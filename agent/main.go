package main

import (
	"github.com/lzkking/edr/agent/internal/agent"
	"github.com/lzkking/edr/agent/internal/heartbeat"
	"github.com/lzkking/edr/agent/internal/plugin"
	"github.com/lzkking/edr/agent/internal/resource"
	"github.com/lzkking/edr/agent/internal/transport"
	"github.com/lzkking/edr/agent/log"
	"go.uber.org/zap"
	"os"
	"sync"
)

func main() {
	log.Init()
	zap.S().Infof("agent start up")

	cpuPercent, rss, readSpeed, writeSpeed, fds, startAt, err := resource.GetPidInfo(int32(os.Getpid()))
	if err != nil {
		zap.S().Errorf("获取进程资源信息失败")
	} else {
		zap.S().Debugf("cpuPercent: %v, rss: %v, readSpeed: %v, writeSpeed: %v, fds: %v, startAt: %v", cpuPercent, rss, readSpeed, writeSpeed, fds, startAt)
	}

	wg := &sync.WaitGroup{}

	wg.Add(3)
	go heartbeat.StartUp(agent.Context, wg)
	go plugin.StartUp(agent.Context, wg)
	go func() {
		transport.StartUp(agent.Context, wg)
		agent.Cancel()
	}()

	wg.Wait()
}
