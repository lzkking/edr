package main

import (
	"github.com/lzkking/edr/agent/internal/agent"
	"github.com/lzkking/edr/agent/internal/heartbeat"
	"github.com/lzkking/edr/agent/internal/plugin"
	"github.com/lzkking/edr/agent/internal/transport"
	"github.com/lzkking/edr/agent/log"
	"go.uber.org/zap"
	"sync"
)

func main() {
	log.Init()
	zap.S().Infof("agent start up")

	wg := &sync.WaitGroup{}

	wg.Add(3)
	go heartbeat.StartUp(agent.Context, wg)
	go plugin.StartUp(agent.Context, wg)
	go func() {
		//	传输信息
		transport.StartUp(agent.Context, wg)
		agent.Cancel()
	}()

	wg.Wait()
}
