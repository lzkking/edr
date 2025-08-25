package handler

import (
	pb "github.com/lzkking/edr/edrproto"
)

type TransferHandler struct {
	pb.UnimplementedServiceServer
}

func (t *TransferHandler) Transfer(stream pb.Service_TransferServer) error {

	return nil
}
