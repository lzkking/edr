package pool

import (
	"context"
	pb "github.com/lzkking/edr/edrproto"
)

type Connection struct {
	Ctx        context.Context
	CancelFuc  context.CancelFunc
	AgentId    string
	SourceAddr string
	CreateAt   int64
	Commands   chan *Command
}

type Command struct {
	Command *pb.Command
	Error   error
	Ready   chan bool
}
