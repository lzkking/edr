package grpctrans

import (
	"fmt"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/handler"
	pb "github.com/lzkking/edr/edrproto"
	"google.golang.org/grpc"
	"net"
	"os"
)

func Run() {
	runServer()
}

func runServer() {
	opts := []grpc.ServerOption{}

	server := grpc.NewServer(opts...)
	pb.RegisterServiceServer(server, &handler.TransferHandler{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 13423))
	if err != nil {
		os.Exit(-1)
	}

	if err = server.Serve(lis); err != nil {
		os.Exit(-1)
	}
}
