package main

import (
	"context"
	"github.com/lzkking/edr/manager/internal/route"
	"github.com/lzkking/edr/manager/log"
	"go.uber.org/zap"
	"sync"
)

func main() {
	log.Init()
	zap.S().Infof("管理平台启动")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	route.Route(context.Background(), wg)

	wg.Wait()
}
