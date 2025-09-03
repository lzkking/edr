package config

import (
	"encoding/json"
	"github.com/lzkking/edr/agent/assets"
	"os"
	"path/filepath"
)

const (
	AgentConfigFile = "agent-config.json"
	AgentLogFile    = "agent-log.log"
)

type AgentCenterConfig struct {
	LogFile string `json:"log_file"`
	RunMode string `json:"run_mode"`
	TmpFile string `json:"tmp_file"`
	WorkDir string `json:"work_dir"`
}

func GetServerConfigPath() string {
	appWorkDir := assets.GetAgentRootAppDir()

	serverConfigPath := filepath.Join(appWorkDir, "configs", AgentConfigFile)

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
		LogFile: filepath.Join(assets.GetAgentRootAppDir(), "log", AgentLogFile),
		RunMode: "DEBUG",
		TmpFile: filepath.Join(assets.GetAgentRootAppDir(), "tmp"),
		WorkDir: filepath.Join(assets.GetAgentRootAppDir(), "work"),
	}
}
