package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/lzkking/edr/agent-center/config"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const GetConfigUrl = "http://%s/agent-center/get-agent-config"

const (
	getManagerUrl = "http://%s/list/manager"
)

var (
	managerHosts []string
)

type PlgConfig struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Version      string   `json:"version"`
	SHA256       string   `json:"sha256"`
	Signature    string   `json:"signature"`
	DownloadUrls []string `json:"download_urls"`
	Detail       string   `json:"detail"`
}

func GetConfigFromManager(agentID string, detail map[string]interface{}) ([]*pb.ConfigItem, error) {
	data := &grequests.RequestOptions{
		JSON: detail,
	}

	if len(managerHosts) == 0 {
		err := GetManagerHosts()
		if err != nil {
			return nil, errors.New("获取manager地址信息失败")
		}
	}

	resp, err := grequests.Post(fmt.Sprintf(GetConfigUrl, managerHosts[rand.Intn(len(managerHosts))]), data)
	if err != nil {
		zap.S().Errorf("从manager获取插件配置信息失败,失败的原因为:%v", err)
		return nil, err
	}

	var plgConfigs []PlgConfig
	err = json.Unmarshal(resp.Bytes(), &plgConfigs)
	if err != nil {
		zap.S().Errorf("反序列需要下发的插件数据失败,失败原因:%v", err)
		return nil, err
	}

	var pbPlgConfigs []*pb.ConfigItem
	for _, plgConfig := range plgConfigs {
		tmp := &pb.ConfigItem{
			Name:        plgConfig.Name,
			Type:        plgConfig.Type,
			Version:     plgConfig.Version,
			SHA256:      plgConfig.SHA256,
			Signature:   plgConfig.Signature,
			DownloadURL: plgConfig.DownloadUrls,
			Detail:      plgConfig.Detail,
		}
		pbPlgConfigs = append(pbPlgConfigs, tmp)
	}

	return pbPlgConfigs, nil
}

type managerHost struct {
	Hosts []string `json:"hosts"`
}

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

	managerHosts = a.Hosts
	return nil
}

func init() {
	rand.Seed(time.Now().Unix())
	GetManagerHosts()
}
