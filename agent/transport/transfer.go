package transport

import (
	"context"
	"github.com/lzkking/edr/agent/agent"
	"github.com/lzkking/edr/agent/buffer"
	"github.com/lzkking/edr/agent/host"
	pb "github.com/lzkking/edr/edrproto"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func startTransfer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	subWg := &sync.WaitGroup{}
	defer subWg.Wait()

	//	连接agent-center
	conn, err := grpc.Dial("localhost:10981", grpc.WithInsecure())
	if err != nil {
		return
	}

	client := pb.NewServiceClient(conn)
	stream, err := client.Transfer(ctx)
	if err != nil {
		return
	}
	subCtx, cancel := context.WithCancel(ctx)
	subWg.Add(2)

	go handleSend(subCtx, subWg, stream)
	go func() {
		handleReceive(subCtx, subWg, stream)
		cancel()
	}()
	subWg.Wait()
	cancel()
}

func handleSend(ctx context.Context, wg *sync.WaitGroup, stream pb.Service_TransferClient) {
	defer wg.Done()
	defer stream.CloseSend()

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			{
				recs := buffer.ReadEncodedRecords()
				if len(recs) > 0 {
					err := stream.Send(&pb.PackagedData{
						Records:      recs,
						AgentId:      agent.ID,
						IntranetIpv4: host.PrivateIpv4.Load().([]string),
						ExtranetIpv4: host.PublicIpv4.Load().([]string),
						IntranetIpv6: host.PrivateIpv6.Load().([]string),
						ExtranetIpv6: host.PublicIpv6.Load().([]string),
						Hostname:     host.Name.Load().(string),
						Version:      agent.Version,
						Product:      agent.Product,
					})
					if err != nil {

					}

					for _, rec := range recs {
						buffer.PutEncodedRecord(rec)
					}
				}
			}
		}
	}
}

func handleReceive(ctx context.Context, wg *sync.WaitGroup, stream pb.Service_TransferClient) {
	defer wg.Done()
	for {
		_, err := stream.Recv()
		if err != nil {
			return
		}
	}
}
