package manager

import (
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
)

const GetConfigUrl = "http://%s/agent-center/get-agent-config"

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

	resp, err := grequests.Post(fmt.Sprintf(GetConfigUrl, "localhost:65000"), data)
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
