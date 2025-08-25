package handler

import (
	"context"
	"github.com/lzkking/edr/edrproto"
)

type TransferHandler struct {
	edrproto.UnimplementedServiceServer
}

func (t *TransferHandler) UploadEvent(ctx context.Context, req *edrproto.Event) (*edrproto.Command, error) {

	return nil, nil
}
