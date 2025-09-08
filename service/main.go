package main

import (
	"github.com/lzkking/edr/service/log"
	"go.uber.org/zap"
)

func main() {
	log.Init()
	zap.S().Infof("service 启动")

}
