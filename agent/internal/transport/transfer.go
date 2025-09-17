package transport

import (
	"context"
	"github.com/lzkking/edr/agent/internal/agent"
	"github.com/lzkking/edr/agent/internal/buffer"
	"github.com/lzkking/edr/agent/internal/host"
	"github.com/lzkking/edr/agent/internal/plugin"
	"github.com/lzkking/edr/agent/internal/transport/connection"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"sync"
	"time"
)

func startTransfer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	subWg := &sync.WaitGroup{}
	defer subWg.Wait()

	//	连接agent-center
	conn, err := connection.GetConnection(ctx)
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
	defer zap.S().Infof("处理接收agent center数据的模块退出")
	zap.S().Infof("开始接收并处理agent传递的命令")
	for {
		cmd, err := stream.Recv()
		if err != nil {
			zap.S().Error(err)
			return
		}

		zap.S().Infof("接收到了agent center的命令")
		if cmd.Task != nil {
			if cmd.Task.Name == agent.Product {
				// 处理传递给agent的命令
				switch cmd.Task.DataType {
				case 1999:
					zap.S().Infof("将要关闭agent")
					agent.Cancel()
					zap.S().Infof("agent关闭成功")
					return
				}
			} else {
				//	处理传递给插件的命令
				plg, ok := plugin.Get(cmd.Task.Name)
				if !ok || plg == nil {
					zap.S().Errorf("没有指定的插件")
					continue
				}

				err = plg.SendTask(cmd.Task)
				if err != nil {
					plg.Errorf("向插件发送控制命令失败")
				}
			}
		}

		//	处理插件配置数据存在的化
		cfgs := map[string]*pb.ConfigItem{}
		for _, config := range cmd.Config {
			cfgs[config.Name] = config
		}

		// 判断是否需要升级agent

		//	同步插件
		delete(cfgs, agent.Product)

		//	判断是否有需要同步的插件数据
		if len(cfgs) > 0 {
			// 同步plugin
			err = plugin.Sync(cfgs)
			if err != nil {
				zap.S().Error(err)
			}
		}

	}
}
