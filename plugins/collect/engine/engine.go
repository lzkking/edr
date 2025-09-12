package engine

import (
	"encoding/binary"
	"encoding/hex"
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
	e.m[h.DataType()] = &handler{
		Handler: h,
		done:    make(chan struct{}),
	}

	//	添加这个接收是为了一开始的采集任务可以进入采集函数,不会出现任务阻塞的bug
	e.m[h.DataType()].done <- struct{}{}
}

// Run - 初始化各组件,并接收命令并处理,为各个组件添加定时任务
func (e *Engine) Run() {

}

func NewEngine(c *plugins.Client) *Engine {
	return &Engine{
		m: make(map[int]*handler),
		s: cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger))),
		c: c,
	}
}
