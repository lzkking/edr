package main

import (
	"context"
	"github.com/lzkking/edr/service/internal/route"
	"github.com/lzkking/edr/service/log"
	"go.uber.org/zap"
	"sync"
)

func main() {
	log.Init()
	zap.S().Infof("service 启动")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go route.StartUp(context.Background(), wg)

	wg.Wait()
}
