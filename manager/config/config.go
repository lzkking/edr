package config

import (
	"encoding/json"
	"github.com/lzkking/edr/manager/assets"
	"os"
	"path/filepath"
)

const (
	ManagerConfigFile = "manager-config.json"
	ManagerLogFile    = "manager-log.log"
)

type AgentCenterConfig struct {
	LogFile       string        `json:"log_file"`
	RunMode       string        `json:"run_mode"`
	TmpFile       string        `json:"tmp_file"`
	WorkDir       string        `json:"work_dir"`
	ListenPort    uint16        `json:"listen_port"`
	MongodbConfig MongodbConfig `json:"mongodb_config"`
}

func GetServerConfigPath() string {
	appWorkDir := assets.GetManagerRootAppDir()

	serverConfigPath := filepath.Join(appWorkDir, "configs", ManagerConfigFile)

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
		LogFile:    filepath.Join(assets.GetManagerRootAppDir(), "log", ManagerLogFile),
		RunMode:    "DEBUG",
		TmpFile:    filepath.Join(assets.GetManagerRootAppDir(), "tmp"),
		WorkDir:    filepath.Join(assets.GetManagerRootAppDir(), "work"),
		ListenPort: 65000,
		MongodbConfig: MongodbConfig{
			Host:        "10.18.201.56:27017",
			User:        "root",
			Password:    "root",
			AuthDB:      "admin",
			DB:          "test",
			MinPoolSize: 5,
			MaxPoolSize: 50,
			RetryWrites: true,
			EnableAuth:  true,
		},
	}
}
