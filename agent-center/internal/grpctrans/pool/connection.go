package pool

import (
	"context"
	pb "github.com/lzkking/edr/edrproto"
	"sync"
)

type Connection struct {
	Ctx              context.Context    `json:"-"`
	CancelFuc        context.CancelFunc `json:"-"`
	AgentId          string             `json:"agent_id"`
	SourceAddr       string             `json:"source_addr"`
	CreateAt         int64              `json:"create_at"`
	Commands         chan *Command      `json:"-"`
	agentDetailLock  sync.RWMutex
	AgentDetail      map[string]interface{} `json:"agent_detail"`
	pluginDetailLock sync.RWMutex
	PluginDetail     map[string]map[string]interface{} `json:"plugin_detail"`
}

type Command struct {
	Command *pb.Command
	Error   error
	Ready   chan bool
}

func (c *Connection) GetAgentDetail() map[string]interface{} {
	c.agentDetailLock.RLock()
	defer c.agentDetailLock.RUnlock()
	if c.AgentDetail == nil {
		return map[string]interface{}{}
	}

	return c.AgentDetail
}

func (c *Connection) SetAgentDetail(agentDetail map[string]interface{}) {
	c.agentDetailLock.Lock()
	defer c.agentDetailLock.Unlock()
	c.AgentDetail = agentDetail
}

func (c *Connection) GetPlgDetail(pluginName string) map[string]interface{} {
	c.pluginDetailLock.Lock()
	defer c.pluginDetailLock.Unlock()
	if c.PluginDetail == nil {
		return map[string]interface{}{}
	}

	return c.PluginDetail[pluginName]
}

func (c *Connection) SetPlgDetail(pluginName string, detail map[string]interface{}) {
	c.pluginDetailLock.Lock()
	defer c.pluginDetailLock.Unlock()

	if c.PluginDetail == nil {
		c.PluginDetail = map[string]map[string]interface{}{}
	}

	c.PluginDetail[pluginName] = detail
}
