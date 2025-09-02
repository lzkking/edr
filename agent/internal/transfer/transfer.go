package transfer

import (
	"context"
	"fmt"
	pb "github.com/lzkking/edr/edrproto"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func Transfer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := grpc.Dial("localhost:10981", grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewServiceClient(conn)
	stream, err := client.Transfer(ctx)
	if err != nil {
		return
	}
	event := pb.Event{
		AgentId: "123",
	}
	err = stream.Send(&event)
	if err != nil {
		return
	}

	c, cancel := context.WithCancel(context.Background())

	go sendData(stream, c)
	go recvData(stream, c)
	<-ctx.Done()
	cancel()
}

func sendData(stream pb.Service_TransferClient, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("发送数据")
		}
		time.Sleep(1 * time.Minute)
	}
}

func recvData(stream pb.Service_TransferClient, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := stream.Recv()
			if err != nil {
				return
			}
			fmt.Println(data.AgentId)
		}
	}
}
