package connection

import (
	"context"
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/lzkking/edr/agent/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
	"time"
)

const (
	getAgentCenterUrl = "http://%s/list/agent-center-grpc"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type agentCenterHosts struct {
	Hosts []string `json:"hosts"`
}

func GetConnection(ctx context.Context) (*grpc.ClientConn, error) {
	serviceHosts := config.GetServerConfig().ServiceHosts

	url := fmt.Sprintf(getAgentCenterUrl, serviceHosts[rand.Intn(len(serviceHosts))])
	option := &grequests.RequestOptions{}
	option.RequestTimeout = 2 * time.Second
	resp, err := grequests.Get(url, option)
	if err != nil {
		zap.S().Warnf("和注册中心发送get请求失败")
		return nil, err
	}

	var a agentCenterHosts
	err = resp.JSON(&a)
	if err != nil {
		zap.S().Warnf("反序列注册中心返回的agent center host 失败")
		return nil, err
	}

	for _, host := range a.Hosts {
		zap.S().Debugf("host is %v", host)
		conn, err := grpc.Dial(host, grpc.WithInsecure())
		if err == nil {
			return conn, nil
		} else {
			zap.S().Errorf("连接grpc服务端失败,失败原因:%v", err)
		}
	}

	return nil, errors.New("connect failed")
}
