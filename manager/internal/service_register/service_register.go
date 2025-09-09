package service_register

import (
	"fmt"
	"github.com/levigross/grequests"
	config2 "github.com/lzkking/edr/manager/config"
	"github.com/lzkking/edr/manager/utils"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

var (
	LocalIp = ""
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

			weight := 0
			registerInfo.Weight = uint32(weight)
			url := fmt.Sprintf("http://%s/register/manager", s.RandomServiceHost())
			option := &grequests.RequestOptions{}
			option.JSON = registerInfo
			option.RequestTimeout = 2 * time.Second
			_, err := grequests.Post(url, option)
			if err != nil {
				zap.S().Warnf("发送agent-center grpc 心跳信息到注册中心失败")
				continue
			}
		case <-s.stop:
			zap.S().Warnf("注册退出")
			return
		}
	}
}

func newServerRegister(name string, ip string, port uint32, serviceHosts []string) *SvrRegister {
	svr := &SvrRegister{
		Name:         name,
		Ip:           ip,
		Port:         port,
		Weight:       0,
		ServiceHosts: serviceHosts,
		stop:         make(chan bool),
	}

	registerInfo := RegisterInfo{
		Name:   svr.Name,
		Ip:     svr.Ip,
		Port:   svr.Port,
		Weight: 0,
	}

	weight := 0
	registerInfo.Weight = uint32(weight)
	url := fmt.Sprintf("http://%s/register/manager", svr.RandomServiceHost())
	option := &grequests.RequestOptions{}
	option.JSON = registerInfo
	option.RequestTimeout = 2 * time.Second
	_, err := grequests.Post(url, option)
	if err != nil {
		zap.S().Warnf("发送agent-center grpc 心跳信息到注册中心失败")
	}

	go svr.registerAgain()

	return svr
}

func NewManagerServiceRegister() *SvrRegister {
	config := config2.GetServerConfig()
	svr := newServerRegister(fmt.Sprintf("%s_manager", LocalIp), LocalIp, uint32(config.ListenPort), config.ServiceHosts)
	return svr
}
