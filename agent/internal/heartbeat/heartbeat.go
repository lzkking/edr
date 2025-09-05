package heartbeat

import (
	"context"
	"github.com/lzkking/edr/agent/internal/buffer"
	"github.com/lzkking/edr/agent/internal/host"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"sync"
	"time"
)

// getAgentStat - 填充心跳包并传递到服务端
func getAgentStat(now time.Time) {
	rec := &pb.Record{
		DataType:  2025,
		Timestamp: now.Unix(),
		Data: &pb.Payload{Fields: make(
			map[string]string),
		},
	}

	rec.Data.Fields["kernel_version"] = host.KernelVersion
	rec.Data.Fields["arch"] = host.Arch
	rec.Data.Fields["platform"] = host.Platform
	rec.Data.Fields["platform_family"] = host.PlatformFamily
	rec.Data.Fields["platform_version"] = host.PlatformVersion

	buffer.WriteRecord(rec)
}

func StartUp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	getAgentStat(time.Now())
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			{
				zap.S().Infof("向agent-center发送心跳包")
				host.RefreshHost()
				getAgentStat(t)
			}
		}
	}
}
