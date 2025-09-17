package consumer

import (
	"context"
	pb "github.com/lzkking/edr/edrproto"
	"github.com/lzkking/edr/thf/internal/collect"
)

func CustomerData(ctx context.Context, m *pb.MQData) error {
	switch m.DataType {
	case 7310, 7311, 7312, 7313, 7314, 7315, 7316, 7317, 7318, 7319, 7320, 7321, 7322:
		collect.DealCollectTaskData(m)
	}
	return nil
}
