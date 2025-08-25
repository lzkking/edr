package handler

import (
	"context"
	pb "github.com/lzkking/edr/edrproto"
)

type TransferHandler struct {
	pb.UnimplementedServiceServer
}

func (t *TransferHandler) UploadEvent(ctx context.Context, req *pb.Event) (*pb.Command, error) {

	return nil, nil
}

func (t *TransferHandler) Transfer(stream pb.Service_TransferServer)
