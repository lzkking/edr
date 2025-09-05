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
