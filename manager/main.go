package main

import (
	"context"
	"github.com/lzkking/edr/manager/internal/route"
	"github.com/lzkking/edr/manager/internal/service_register"
	"github.com/lzkking/edr/manager/log"
	"go.uber.org/zap"
	"sync"
)

func main() {
	log.Init()
	zap.S().Infof("管理平台启动")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go route.Route(context.Background(), wg)

	svrRegister := service_register.NewManagerServiceRegister()
	defer func() {
		svrRegister.Stop()
	}()

	wg.Wait()
}
