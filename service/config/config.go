package config

import (
	"encoding/json"
	"github.com/lzkking/edr/service/assets"
	"os"
	"path/filepath"
)

const (
	ServiceConfigFile = "service-config.json"
	ServiceLogFile    = "service-log.log"
)

type AgentCenterConfig struct {
	LogFile    string   `json:"log_file"`
	RunMode    string   `json:"run_mode"`
	ListenPort uint16   `json:"listen_port"`
	Hosts      []string `json:"hosts"`
	Weight     uint32   `json:"weight"`
}

func GetServerConfigPath() string {
	appWorkDir := assets.GetAgentRootAppDir()

	serverConfigPath := filepath.Join(appWorkDir, "configs", ServiceConfigFile)

	return serverConfigPath
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
		LogFile:    filepath.Join(assets.GetAgentRootAppDir(), "log", ServiceLogFile),
		RunMode:    "DEBUG",
		ListenPort: 10086,
		Hosts:      []string{},
		Weight:     400,
	}
}
