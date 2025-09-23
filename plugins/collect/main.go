package main

import (
	"github.com/go-logr/zapr"
	"github.com/lzkking/edr/plugins/collect/app"
	"github.com/lzkking/edr/plugins/collect/container"
	"github.com/lzkking/edr/plugins/collect/cron"
	"github.com/lzkking/edr/plugins/collect/engine"
	"github.com/lzkking/edr/plugins/collect/integrity"
	"github.com/lzkking/edr/plugins/collect/kmod"
	"github.com/lzkking/edr/plugins/collect/net_interface"
	"github.com/lzkking/edr/plugins/collect/port"
	"github.com/lzkking/edr/plugins/collect/process"
	"github.com/lzkking/edr/plugins/collect/service"
	"github.com/lzkking/edr/plugins/collect/software"
	"github.com/lzkking/edr/plugins/collect/user"
	"github.com/lzkking/edr/plugins/collect/volume"
	plgtran "github.com/lzkking/edr/plugins/lib"
	"github.com/lzkking/edr/plugins/lib/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	l := log.New(
		log.Config{
			MaxSize:     1,
			Path:        "collect.log",
			FileLevel:   zapcore.InfoLevel,
			RemoteLevel: zapcore.ErrorLevel,
			MaxBackups:  10,
			Compress:    true,
			Client:      plgClient,
		},
	)
	defer l.Sync()
	zap.ReplaceGlobals(l)

	//	添加收集引擎
	e := engine.NewEngine(plgClient, zapr.NewLogger(l))

	e.AddHandler(time.Hour*6, &user.UserHandler{})                   // 采集用户信息
	e.AddHandler(time.Hour, &process.ProcessHandler{})               // 采集进程信息
	e.AddHandler(time.Hour, &port.PortHandler{})                     // 采集端口的数据
	e.AddHandler(engine.BeforeDawn(), &software.SoftwareHandler{})   // 采集软件信息
	e.AddHandler(time.Minute*5, &container.ContainerHandler{})       // 采集容器信息
	e.AddHandler(time.Hour*6, &cron.CronHandler{})                   // 获取定时任务信息
	e.AddHandler(time.Hour*6, &service.ServiceHandler{})             // 采集服务信息
	e.AddHandler(engine.BeforeDawn(), &integrity.IntegrityHandler{}) // 收集安装的包与仓库中的包hash值不一致的包
	e.AddHandler(time.Hour*6, &net_interface.NetInterfaceHandler{})  // 采集网络接口信息
	e.AddHandler(time.Hour*6, &volume.VolumeHandler{})               // 采集挂载的硬盘信息
	e.AddHandler(time.Hour, &kmod.KmodHandler{})                     // 采集安装的内核模块
	e.AddHandler(engine.BeforeDawn(), &app.AppHandler{})             // 采集运行的服务

	//  接收命令,运行定时任务,发送采集数据
	e.Run()
}
