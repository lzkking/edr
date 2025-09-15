// Package config
//
//	Author:		龙泽坤
//	Company:	奇富科技
//	Email:		longzekun-jk@360shuke.com
//				lzk1342325850@163.com
//	Date:		2025-03-10
//	File:		config.go
//
//	Description(获取全局配置:默认):
//		LogFile:日志记录文件(工作路径/log/schedule.log)
//		RunMode:日志记录模式(调试模式)
//
//
//	Change:
package config

import (
	"encoding/json"
	"github.com/lzkking/edr/thf/assets"
	"os"
	"path/filepath"
)

const (
	AgentCenterConfigFile = "agent-center-config.json"
	AgentCenterLogFile    = "agent-center-log.log"
	KafkaLogFile          = "kafka-log.log"
)

func GetServerConfigPath() string {
	appWorkDir := assets.GetRootAppDir()

	serverConfigPath := filepath.Join(appWorkDir, "configs", AgentCenterConfigFile)

	return serverConfigPath
}

type AgentCenterConfig struct {
	LogFile        string      `json:"log_file"`
	RunMode        string      `json:"run_mode"`
	ServiceHosts   []string    `json:"service_hosts"`
	GrpcListenPort string      `json:"grpc-listen_port"`
	HttpListenPort string      `json:"http-listen-port"`
	ServiceCenter  []string    `json:"service_center"`
	ConnectLimit   uint64      `json:"connect_limit"`
	KafkaConfig    KafkaConfig `json:"kafka_config"`
}

func (c *AgentCenterConfig) Save() error {
	configPath := GetServerConfigPath()
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0700)
		if err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		panic(err)
	}

	return nil
}

func GetServerConfig() *AgentCenterConfig {
	configPath := GetServerConfigPath()
	config := getDefaultServerConfig()

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return config
		}

		err = json.Unmarshal(data, config)
		if err != nil {
			return config
		}
	}
	err := config.Save()
	if err != nil {
		panic(err)
	}
	return config
}

func getDefaultServerConfig() *AgentCenterConfig {
	return &AgentCenterConfig{
		LogFile: filepath.Join(assets.GetRootAppDir(), "log", AgentCenterLogFile),
		RunMode: "DEBUG",
		ServiceHosts: []string{
			"10.18.201.56:10086",
		},
		GrpcListenPort: "10981",
		HttpListenPort: "10982",
		ConnectLimit:   10,
		KafkaConfig: KafkaConfig{
			KafkaAdders: []string{
				"10.18.201.56:9092",
			},
			Topics: []string{
				"test",
			},
			GroupID:    "test-group",
			ClientID:   "10.18.201.56",
			LogPath:    filepath.Join(assets.GetRootAppDir(), "log", KafkaLogFile),
			EnableAuth: false,
			UserName:   "",
			Password:   "",
		},
	}
}
