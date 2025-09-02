package handler

import (
	"context"
	"errors"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/pool"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"

	"google.golang.org/grpc/peer"
	"time"
)

var (
	GlobalPool *pool.GlobalPool
)

func init() {
	GlobalPool = pool.NewGlobalPool()
}

type TransferHandler struct {
	pb.UnimplementedServiceServer
}

func (t *TransferHandler) Transfer(stream pb.Service_TransferServer) error {
	if !GlobalPool.LoadToken() {
		zap.S().Warnf("超出agent-center可以处理的连接上限")
		err := errors.New("out of max connection limit")
		return err
	}
	defer func() {
		GlobalPool.ReleaseToken()
	}()

	event, err := stream.Recv()
	if err != nil {
		zap.S().Warnf("接收agent的数据流失败")
		return err
	}
	agentId := event.AgentId
	zap.S().Infof("接收到%v的连接请求", agentId)

	p, ok := peer.FromContext(stream.Context())
	if !ok {
		zap.S().Warnf("提取%v的连接信息失败", agentId)
		return errors.New("peer error")
	}
	ctx, cancelButton := context.WithCancel(context.Background())
	createAt := time.Now().UnixNano() / (1000 * 1000 * 1000)
	conn := pool.Connection{
		Ctx:        ctx,
		CancelFuc:  cancelButton,
		AgentId:    agentId,
		SourceAddr: p.Addr.String(),
		CreateAt:   createAt,
		Commands:   make(chan *pool.Command),
	}

	err = GlobalPool.Add(agentId, &conn)
	if err != nil {
		return err
	}
	defer func() {
		GlobalPool.Delete(agentId)
	}()

	//	处理第一次的连接
	handleRawData(event)

	go recvData(stream, &conn)

	go sendData(stream, &conn)

	<-conn.Ctx.Done()
	return nil
}

func recvData(stream pb.Service_TransferServer, conn *pool.Connection) {
	defer conn.CancelFuc()
	for {
		select {
		case <-conn.Ctx.Done():
			return
		default:
			data, err := stream.Recv()
			if err != nil {
				return
			}
			handleRawData(data)
		}
	}
}

func sendData(stream pb.Service_TransferServer, conn *pool.Connection) {
	defer conn.CancelFuc()
	for {
		select {
		case <-conn.Ctx.Done():
			return
		case cmd := <-conn.Commands:
			if cmd == nil {
				zap.S().Infof("get close signal, now close the send direction,%v", conn.AgentId)
				return
			}

			err := stream.Send(cmd.Command)
			if err != nil {
				zap.S().Errorf("send Command to %v error, command is %v, error is %v", conn.AgentId, cmd, err)
				cmd.Error = err
				close(cmd.Ready)
				return
			}

			zap.S().Infof("send command to %v, command is %v", conn.AgentId, cmd)
			cmd.Error = nil
			close(cmd.Ready)
		}
	}
}

func handleRawData(event *pb.Event) {

}
