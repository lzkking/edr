package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/pool"
	pb "github.com/lzkking/edr/edrproto"

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
	if GlobalPool.LoadToken() {
		err := errors.New("out of max connection limit")
		return err
	}

	defer func() {
		GlobalPool.ReleaseToken()
	}()

	event, err := stream.Recv()
	if err != nil {
		return err
	}

	agentId := event.AgentId

	p, ok := peer.FromContext(stream.Context())
	if !ok {
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
	}

	err = GlobalPool.Add(agentId, &conn)
	if err != nil {
		return err
	}
	defer func() {
		GlobalPool.Delete(agentId)
	}()

	//	处理第一次的连接

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
			fmt.Printf("recv data from %v", data.AgentId)
		}
	}
}

func sendData(stream pb.Service_TransferServer, conn *pool.Connection) {
	defer conn.CancelFuc()
	for {
		time.Sleep(30)
		select {
		case <-conn.Ctx.Done():
			return
		default:
			fmt.Println("发送命令")
		}
	}
}
