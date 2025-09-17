package main

import (
	"fmt"
	"github.com/lzkking/edr/plugins/collect/engine"
	"github.com/lzkking/edr/plugins/collect/port"
	"github.com/lzkking/edr/plugins/collect/process"
	"github.com/lzkking/edr/plugins/collect/software"
	"github.com/lzkking/edr/plugins/collect/user"
	plgtran "github.com/lzkking/edr/plugins/lib"
	"math/rand"
	"runtime"
	"time"
)

func init() {
	runtime.GOMAXPROCS(8)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	plgClient := plgtran.New()

	//	添加收集引擎
	e := engine.NewEngine(plgClient)

	e.AddHandler(time.Hour*6, &user.UserHandler{})     // 采集用户信息
	e.AddHandler(time.Hour, &process.ProcessHandler{}) // 采集进程信息
	e.AddHandler(time.Hour, &port.PortHandler{})       // 采集端口的数据
	e.AddHandler(-1, &software.SoftwareHandler{})      // 采集软件信息

	fmt.Println("处理采集")
	//  接收命令,运行定时任务,发送采集数据
	e.Run()
}
