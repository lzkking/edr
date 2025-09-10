package route

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/config"
	"github.com/lzkking/edr/manager/internal/handler/agentcenter"
	"github.com/lzkking/edr/manager/internal/handler/manager/plugins"
	"go.uber.org/zap"
	"sync"
)

func Route(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	zap.S().Infof("开始管理端的路由监听")

	r := gin.Default()
	listenPort := config.GetServerConfig().ListenPort

	//agent-center连接到的组
	agentCenter := r.Group("agent-center")
	{
		agentCenter.POST("get-agent-config", agentcenter.GetAgentConfig)
	}

	manager := r.Group("manager")
	{
		manager.POST("upload-plugin", plugins.UploadPlugin)
	}

	r.Run(fmt.Sprintf(":%d", listenPort))
}
