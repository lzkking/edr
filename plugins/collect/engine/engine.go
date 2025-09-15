package engine

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	plugins "github.com/lzkking/edr/plugins/lib"
	"github.com/robfig/cron/v3"
	"hash/fnv"
	"time"
)

type Handler interface {
	//	Handle - 具体的处理函数 c - 具体的发送数据的对象, seq - 当前执行任务的唯一标识符
	Handle(c *plugins.Client, seq string)
	Name() string
	DataType() int
}

type handler struct {
	Handler
	done chan struct{}
}

func (h *handler) Handle(c *plugins.Client) {
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
		h.Handler.Handle(c, seq)
	default:
		//在此处阻塞等待
		t = <-h.done
	}

	h.done <- t
}

type Engine struct {
	m map[int]*handler
	s *cron.Cron
	c *plugins.Client
}

func (e *Engine) AddHandler(interval time.Duration, h Handler) {
	fmt.Println("添加插件")
	e.m[h.DataType()] = &handler{
		Handler: h,
		done:    make(chan struct{}, 1),
	}

	fmt.Println("向管道塞数据")
	//	添加这个接收是为了一开始的采集任务可以进入采集函数,不会出现任务阻塞的bug
	e.m[h.DataType()].done <- struct{}{}

	fmt.Println("塞数据完成")
}

// Run - 初始化各组件,并接收命令并处理,为各个组件添加定时任务
func (e *Engine) Run() {
	// 	首次运行先进行初始化,初始化完成后将采集任务加入定时器
	//	接收agent传递来的数据并进行处理

	fmt.Println("采集数据")
	for _, h := range e.m {
		fmt.Println("采集数据")
		go func(h *handler) {
			h.Handle(e.c)

			//	将采集任务模块加到定时器模块中
		}(h)
	}

	for {
		t, err := e.c.ReceiveTask()
		if err != nil {
			break
		}

		//	看是否有相关处理模块
		if h, ok := e.m[int(t.DataType)]; ok {
			h.Handle(e.c)

			//发送完成采集任务

		} else {
			// 没有相关的处理结构,退出

		}
	}

	//	停止与agent的通信,关闭定时器模块
	e.c.Close()
	e.s.Stop()
}

func NewEngine(c *plugins.Client) *Engine {
	return &Engine{
		m: make(map[int]*handler),
		s: cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger))),
		c: c,
	}
}
