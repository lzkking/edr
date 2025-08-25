package pool

import "context"

type Connection struct {
	Ctx        context.Context
	CancelFuc  context.CancelFunc
	AgentId    string
	SourceAddr string
	CreateAt   int64
}
