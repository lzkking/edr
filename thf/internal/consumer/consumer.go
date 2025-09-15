package consumer

import (
	"context"
	pb "github.com/lzkking/edr/edrproto"
	"github.com/lzkking/edr/thf/internal/collect"
)

func CustomerData(ctx context.Context, m *pb.MQData) error {
	switch m.DataType {
	case 7310:
		collect.DealCollectTaskData(m)
	}
	return nil
}
