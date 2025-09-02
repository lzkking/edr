package main

import (
	"github.com/lzkking/edr/agent/agent"
	"github.com/lzkking/edr/agent/heartbeat"
	"github.com/lzkking/edr/agent/transport"
	"sync"
)

func main() {

	wg := &sync.WaitGroup{}

	wg.Add(2)

	go heartbeat.StartUp(agent.Context, wg)
	go func() {
		transport.StartUp(agent.Context, wg)
		agent.Cancel()
	}()

	wg.Wait()
}
