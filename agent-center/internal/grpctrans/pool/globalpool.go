package pool

import (
	"errors"
	"github.com/patrickmn/go-cache"
)

type GlobalPool struct {
	connectPool *cache.Cache
	tokenChan   chan bool
}

func NewGlobalPool() *GlobalPool {
	g := &GlobalPool{
		connectPool: cache.New(-1, -1),
		tokenChan:   make(chan bool, 256),
	}

	for i := 0; i < 256; i++ {
		g.tokenChan <- true
	}

	return g
}

func (g *GlobalPool) Add(agentId string, connect *Connection) error {
	_, ok := g.connectPool.Get(agentId)
	if ok {
		return errors.New("agent id conflict")
	}

	g.connectPool.Set(agentId, connect, -1)
	return nil
}

func (g *GlobalPool) Delete(agentId string) {
	g.connectPool.Delete(agentId)
}

func (g *GlobalPool) LoadToken() bool {
	select {
	case _, ok := <-g.tokenChan:
		if ok {
			return true
		}
	default:
	}

	return false
}

func (g *GlobalPool) ReleaseToken() {
	g.tokenChan <- true
}
