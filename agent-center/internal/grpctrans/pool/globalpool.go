package pool

import (
	"errors"
	"github.com/lzkking/edr/agent-center/config"
	pb "github.com/lzkking/edr/edrproto"
	"github.com/patrickmn/go-cache"
	"time"
)

type GlobalPool struct {
	connectPool *cache.Cache
	tokenChan   chan bool
}

func NewGlobalPool() *GlobalPool {
	grpcConfig := config.GetServerConfig()
	connectLimit := grpcConfig.ConnectLimit
	g := &GlobalPool{
		connectPool: cache.New(-1, -1),
		tokenChan:   make(chan bool, connectLimit),
	}

	for i := uint64(0); i < connectLimit; i++ {
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

func (g *GlobalPool) GetAgentIdList() []string {
	connMap := g.connectPool.Items()
	agentIds := make([]string, 0)
	for _, v := range connMap {
		conn := v.Object.(*Connection)
		agentIds = append(agentIds, conn.AgentId)
	}
	return agentIds
}

func (g *GlobalPool) GetByID(agentID string) (*Connection, error) {
	tmp, ok := g.connectPool.Get(agentID)
	if !ok {
		return nil, errors.New("agent id not found")
	}
	return tmp.(*Connection), nil
}

func (g *GlobalPool) PostCommand(agentID string, command *pb.Command) error {
	conn, err := g.GetByID(agentID)
	if err != nil {
		return err
	}

	cmdToSend := &Command{
		Command: command,
		Error:   nil,
		Ready:   make(chan bool, 1),
	}

	select {
	case conn.Commands <- cmdToSend:
	case <-time.After(2 * time.Second):
		return errors.New("command list is full")
	}

	select {
	case <-cmdToSend.Ready:
		return cmdToSend.Error
	case <-time.After(2 * time.Second):
		return errors.New("the command have been sent,but get results timeout")

	}
}
