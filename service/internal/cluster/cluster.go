package cluster

import (
	"github.com/lzkking/edr/service/config"
	"sync"
)

var (
	GlobalCluster *Cluster
)

type Cluster struct {
	mu    sync.RWMutex
	Hosts []string
}

func (c *Cluster) GetOtherHost() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var hosts []string

	hosts = c.Hosts

	return hosts
}

//todo 检查服务注册中心是否健康

func NewClusterFromConfig() *Cluster {
	hosts := config.GetServerConfig().Hosts
	cluster := &Cluster{
		Hosts: hosts,
	}

	return cluster
}

func init() {
	GlobalCluster = NewClusterFromConfig()
}
