package pool

import (
	"errors"
	"github.com/lzkking/edr/agent-center/config"
	pb "github.com/lzkking/edr/edrproto"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"time"
)

type GlobalPool struct {
	connectPool *cache.Cache
	tokenChan   chan bool
	confChan    chan string
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

	// 向manager获取下发到agent的插件信息
	go g.checkConfig()

	// 向manager发送任务执行的结果

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

func (g *GlobalPool) checkConfig() {
	for {
		select {
		case agentId := <-g.confChan:
			conn, err := g.GetByID(agentId)
			if err != nil {
				zap.S().Warnf("未在连接池中检索到:%v,GetByID失败,失败原因:%v", agentId, err)
				continue
			}

			//将conn的信息发往manager
			zap.S().Infof("将%v的心跳数据发往manager", conn.AgentId)
		}
	}
}

func (g *GlobalPool) PostLastConfig(agentID string) error {
	select {
	case g.confChan <- agentID:
	default:
		return errors.New("confChan 满的，稍后重试")
	}
	return nil
}
