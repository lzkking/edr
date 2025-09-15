package manager

import (
	"fmt"
	"github.com/levigross/grequests"
	"github.com/lzkking/edr/thf/config"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const (
	getManagerUrl = "http://%s/list/manager"
)

type managerHost struct {
	Hosts []string `json:"hosts"`
}

var ManagerHosts []string

func GetManagerHosts() error {
	serviceHosts := config.GetServerConfig().ServiceHosts
	url := fmt.Sprintf(getManagerUrl, serviceHosts[rand.Intn(len(serviceHosts))])
	option := &grequests.RequestOptions{}
	option.RequestTimeout = 2 * time.Second
	resp, err := grequests.Get(url, option)
	if err != nil {
		zap.S().Warnf("和注册中心发送get请求失败")
		return err
	}

	var a managerHost
	err = resp.JSON(&a)
	if err != nil {
		zap.S().Warnf("反序列注册中心返回的agent center host 失败")
		return err
	}

	ManagerHosts = a.Hosts
	return nil
}

func init() {
	rand.Seed(time.Now().Unix())
	GetManagerHosts()
}
