package main

import (
	"github.com/lzkking/edr/plugins/collect/engine"
	plgtran "github.com/lzkking/edr/plugins/lib"
)

func main() {
	plgClient := plgtran.New()

	//	添加收集引擎
	e := engine.NewEngine(plgClient)

	//  接收命令,运行定时任务,发送采集数据
	e.Run()
}
