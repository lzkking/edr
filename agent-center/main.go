package main

import (
	"github.com/lzkking/edr/agent-center/common"
	"github.com/lzkking/edr/agent-center/internal/grpctrans"
	"github.com/lzkking/edr/agent-center/internal/httptrans"
	"github.com/lzkking/edr/agent-center/log"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

func init() {
	signal.Notify(common.Sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
}

func main() {
	log.Init()
	zap.S().Infof("agent-center启动")

	zap.S().Infof("向服务中心注册")

	zap.S().Infof("开启http监听")
	go httptrans.Run()

	zap.S().Infof("开启grpc服务端监听,等待Agent连接")
	go grpctrans.Run()

	sig := <-common.Sig

	zap.S().Debugf("收到%v信号", sig)
	//todo 释放资源
}
