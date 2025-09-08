package pool

import (
	"errors"
	"github.com/k0kubun/pp/v3"
	"github.com/lzkking/edr/agent-center/config"
	"github.com/lzkking/edr/agent-center/internal/manager"
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
		confChan:    make(chan string),
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
			agentDetail := conn.GetAgentDetail()
			zap.S().Debugf("%v", pp.Sprintf("%v", agentDetail))

			plgConfigs, err := manager.GetConfigFromManager(agentId, agentDetail)
			if err != nil {
				continue
			}

			if len(plgConfigs) > 0 {
				pbCommand := &pb.Command{
					AgentCtrl: 0,
					Config:    plgConfigs,
				}

				err = g.PostCommand(agentId, pbCommand)
				if err != nil {
					zap.S().Errorf("向agent发送加载的插件的命令失败,失败的原因:%v", err)
				}
			} else {
				zap.S().Infof("目前没有需要下发到agent的插件")
			}
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
