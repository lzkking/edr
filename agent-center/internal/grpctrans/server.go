package grpctrans

import (
	"fmt"
	"github.com/lzkking/edr/agent-center/config"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/handler"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
)

func Run() {
	runServer()
}

func runServer() {
	grpcConfig := config.GetServerConfig()

	opts := []grpc.ServerOption{}

	server := grpc.NewServer(opts...)
	pb.RegisterServiceServer(server, &handler.TransferHandler{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcConfig.GrpcListenPort))
	if err != nil {
		zap.S().Errorf("启动grpc的对%v的监听失败,失败原因:%v", grpcConfig.GrpcListenPort, err)
		os.Exit(-1)
	}

	if err = server.Serve(lis); err != nil {
		zap.S().Errorf("启动监听服务失败,失败原因:%v", err)
		os.Exit(-1)
	}
}
