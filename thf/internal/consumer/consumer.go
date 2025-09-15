package consumer

import (
	"context"
	pb "github.com/lzkking/edr/edrproto"
)

func CustomerData(ctx context.Context, m *pb.MQData) error {
	switch m.DataType {
	case 7310:
	}

	return nil
}
