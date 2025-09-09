package service_register

import (
	"fmt"
	"github.com/levigross/grequests"
	config2 "github.com/lzkking/edr/agent-center/config"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/handler"
	"github.com/lzkking/edr/agent-center/utils"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"time"
)

var (
	LocalIp = ""
)

const (
	TypeGrpc = 1
	TypeHttp = 2
)

func init() {
	localIp, err := utils.GetOutboundIP()
	if err != nil {
		panic("获取本机的出口ip失败")
	}
	LocalIp = localIp

	rand.Seed(time.Now().Unix())
}

type SvrRegister struct {
	Name         string
	Ip           string
	Port         uint32
	Weight       uint32
	Type         int
	ServiceHosts []string
	stop         chan bool
}

type RegisterInfo struct {
	Name   string `json:"name"`
	Ip     string `json:"ip"`
	Port   uint32 `json:"port"`
	Weight uint32 `json:"weight"`
}

func (s *SvrRegister) RandomServiceHost() string {
	return s.ServiceHosts[rand.Intn(len(s.ServiceHosts))]
}

func (s *SvrRegister) Stop() {
	close(s.stop)
}

func (s *SvrRegister) registerAgain() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			registerInfo := RegisterInfo{
				Name:   s.Name,
				Ip:     s.Ip,
				Port:   s.Port,
				Weight: 0,
			}

			switch s.Type {
			case TypeHttp:
				weight := 0
				registerInfo.Weight = uint32(weight)
				url := fmt.Sprintf("http://%s/register/agent-center-http", s.RandomServiceHost())
				option := &grequests.RequestOptions{}
				option.JSON = registerInfo
				option.RequestTimeout = 2 * time.Second
				_, err := grequests.Post(url, option)
				if err != nil {
					zap.S().Warnf("发送agent-center grpc 心跳信息到注册中心失败")
					continue
				}
			case TypeGrpc:
				weight := len(handler.GlobalPool.GetAgentIdList())
				registerInfo.Weight = uint32(weight)
				url := fmt.Sprintf("http://%s/register/agent-center-grpc", s.RandomServiceHost())
				option := &grequests.RequestOptions{}
				option.JSON = registerInfo
				option.RequestTimeout = 2 * time.Second
				_, err := grequests.Post(url, option)
				if err != nil {
					zap.S().Warnf("发送agent-center grpc 心跳信息到注册中心失败")
					continue
				}
			}
		case <-s.stop:
			zap.S().Warnf("注册退出")
			return
		}
	}
}

func newServerRegister(name string, ip string, port uint32, serviceHosts []string, Type int) *SvrRegister {
	svr := &SvrRegister{
		Name:         name,
		Ip:           ip,
		Port:         port,
		Weight:       0,
		Type:         Type,
		ServiceHosts: serviceHosts,
		stop:         make(chan bool),
	}

	registerInfo := RegisterInfo{
		Name:   svr.Name,
		Ip:     svr.Ip,
		Port:   svr.Port,
		Weight: 0,
	}

	switch svr.Type {
	case TypeHttp:
		weight := 0
		registerInfo.Weight = uint32(weight)
		url := fmt.Sprintf("http://%s/register/agent-center-http", svr.RandomServiceHost())
		option := &grequests.RequestOptions{}
		option.JSON = registerInfo
		option.RequestTimeout = 2 * time.Second
		_, err := grequests.Post(url, option)
		if err != nil {
			zap.S().Warnf("发送agent-center grpc 心跳信息到注册中心失败")
		}
	case TypeGrpc:
		weight := len(handler.GlobalPool.GetAgentIdList())
		registerInfo.Weight = uint32(weight)
		url := fmt.Sprintf("http://%s/register/agent-center-grpc", svr.RandomServiceHost())
		option := &grequests.RequestOptions{}
		option.JSON = registerInfo
		option.RequestTimeout = 2 * time.Second
		_, err := grequests.Post(url, option)
		if err != nil {
			zap.S().Warnf("发送agent-center grpc 心跳信息到注册中心失败")
		}
	}

	go svr.registerAgain()

	return svr
}

func NewGrpcServiceRegister() *SvrRegister {
	config := config2.GetServerConfig()
	portInt, _ := strconv.Atoi(config.GrpcListenPort)
	svr := newServerRegister(fmt.Sprintf("%s_grpc", LocalIp), LocalIp, uint32(portInt), config.ServiceHosts, TypeGrpc)
	return svr
}

func NewHttpServiceRegister() *SvrRegister {
	config := config2.GetServerConfig()
	portInt, _ := strconv.Atoi(config.HttpListenPort)
	svr := newServerRegister(fmt.Sprintf("%s_http", LocalIp), LocalIp, uint32(portInt), config.ServiceHosts, TypeHttp)
	return svr
}
