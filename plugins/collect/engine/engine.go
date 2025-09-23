package engine

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	plugins "github.com/lzkking/edr/plugins/lib"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"hash/fnv"
	"math/rand"
	"sync"
	"time"
)

type Records map[string]map[string]string

func (r Records) Find(key, value string) (string, map[string]string, bool) {
	for k, v := range r {
		if v[key] == value {
			return k, v, true
		}
	}
	return "", map[string]string{}, false
}

type Cache struct {
	// DataType-Key-Record
	m  map[int]Records
	mu *sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		m:  map[int]Records{},
		mu: &sync.RWMutex{},
	}
}

// don't modify returned map
func (c *Cache) Get(dt int, key string) (map[string]string, bool) {
	c.mu.RLock()
	res, ok := c.m[dt][key]
	c.mu.RUnlock()
	return res, ok
}
func (c *Cache) Put(dt int, key string, value map[string]string) {
	c.mu.Lock()
	c.m[dt][key] = value
	c.mu.Unlock()
}
func (c *Cache) clear(dt int) {
	c.mu.Lock()
	c.m[dt] = map[string]map[string]string{}
	c.mu.Unlock()
}

type Handler interface {
	//	Handle - 具体的处理函数 c - 具体的发送数据的对象, seq - 当前执行任务的唯一标识符
	Handle(c *plugins.Client, cache *Cache, seq string)
	Name() string
	DataType() int
}

type handler struct {
	l *zap.SugaredLogger
	Handler
	done     chan struct{}
	interval time.Duration
}

func (h *handler) Handle(c *plugins.Client, cache *Cache) {
	fmt.Println("开始处理采集用户数据任务")
	var t struct{}
	select {
	// 一开始 h.done 管道中有数据,来一个任务就会被消耗, 然后如果再来相同的任务就不会进入执行分支,
	// 而是进入不执行分支,但这时候就存在一个问题,后来的任务会前于后来的任务执行完成,这是不可接受的
	// 所以只需要让它阻塞就行,等待当前任务执行完成,后再返回数据,这建立在并行运行采集任务的前提下,
	// 目前存在定时任务,所以存在同一时间执行相同任务的情况。
	case t = <-h.done:
		// 生成当前采集任务的序列号,单次采集任务的序列号一样
		f := fnv.New32()
		binary.Write(f, binary.LittleEndian, time.Now().UnixNano())
		seq := hex.EncodeToString(f.Sum(nil))
		h.l.Infof("do work")

		h.Handler.Handle(c, cache, seq)
	default:
		//在此处阻塞等待
		h.l.Info("wait work")
		t = <-h.done
	}
	h.l.Info("work done")
	h.done <- t
	h.l.Info("handled")
}

type Engine struct {
	m     map[int]*handler
	s     *cron.Cron
	c     *plugins.Client
	cache *Cache
}

func (e *Engine) AddHandler(interval time.Duration, h Handler) {
	e.m[h.DataType()] = &handler{
		l:        zap.S().With("name", h.Name()),
		Handler:  h,
		done:     make(chan struct{}, 1),
		interval: interval,
	}

	//	添加这个接收是为了一开始的采集任务可以进入采集函数,不会出现任务阻塞的bug
	e.m[h.DataType()].done <- struct{}{}
}

func BeforeDawn() time.Duration {
	return -1
}

// Run - 初始化各组件,并接收命令并处理,为各个组件添加定时任务
func (e *Engine) Run() {
	// 	首次运行先进行初始化,初始化完成后将采集任务加入定时器
	//	接收agent传递来的数据并进行处理

	zap.S().Info("采集引擎启动")
	for _, h := range e.m {
		go func(h *handler) {
			var spec string
			var r int
			minutes := int(h.interval.Minutes())
			if h.interval == BeforeDawn() {
				spec = fmt.Sprintf("%d %d * * *", rand.Intn(60), rand.Intn(6))
				r = rand.Intn(14400) + 7200
			} else if minutes > 0 {
				r = rand.Intn(minutes * 60)
				spec = fmt.Sprintf("@every %dm", int(minutes))
			} else {
				panic("unknown interval")
			}

			h.l.Infof("init call will after %d secs\n", r)
			time.Sleep(time.Second * time.Duration(r))
			h.l.Info("init call")

			h.Handle(e.c, e.cache)

			time.Sleep(time.Minute * time.Duration(minutes))
			//	将采集任务模块加到定时器模块中
			e.s.AddFunc(spec, func() {
				h.Handle(e.c, e.cache)
			})

			h.l.Info("add func to scheduler successfully")
		}(h)
	}

	go func() {
		zap.S().Info("定时器模块启动")
		e.s.Run()
	}()

	for {
		t, err := e.c.ReceiveTask()
		if err != nil {
			break
		}

		//	看是否有相关处理模块
		zap.S().Infof("received task")
		if h, ok := e.m[int(t.DataType)]; ok {
			h.Handle(e.c, e.cache)

			//发送完成采集任务
			e.c.SendRecord(
				&plugins.Record{
					DataType:  5100,
					Timestamp: time.Now().Unix(),
					Data: &plugins.Payload{
						Fields: map[string]string{
							"status": "succeed",
							"msg":    "",
							"token":  t.Token,
						},
					}})

		} else {
			// 没有相关的处理结构,退出
			// can't find handler
			e.c.SendRecord(
				&plugins.Record{
					DataType:  5100,
					Timestamp: time.Now().Unix(),
					Data: &plugins.Payload{
						Fields: map[string]string{
							"status": "failed",
							"msg":    "the data_type hasn't been implemented",
							"token":  t.Token,
						},
					}})
		}
	}

	//	停止与agent的通信,关闭定时器模块
	zap.S().Warn("engine will stop")
	e.c.Close()
	e.s.Stop()
}

func NewEngine(c *plugins.Client, l cron.Logger) *Engine {
	return &Engine{
		m: make(map[int]*handler),
		s: cron.New(cron.WithChain(cron.SkipIfStillRunning(l)), cron.WithLogger(l)),
		c: c,
	}
}
