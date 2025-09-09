package route

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/config"
	"github.com/lzkking/edr/service/internal/handler/agent_center"
	"github.com/lzkking/edr/service/internal/handler/manager"
	"github.com/lzkking/edr/service/internal/handler/service"
	"net/http"
	"sync"
)

func StartUp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	listenPort := config.GetServerConfig().ListenPort
	r := gin.Default()

	registerGroup := r.Group("/register")
	{
		registerGroup.POST("/agent-center-grpc", agent_center.AgentCenterGrpcRegister)
		registerGroup.POST("/agent-center-http", agent_center.AgentCenterHttpRegister)
		registerGroup.POST("/manager", manager.ManagerRegister)
	}

	listGroup := r.Group("/list")
	{
		listGroup.GET("/agent-center-grpc", agent_center.ListAgentCenterGrpc)
		listGroup.GET("/agent-center-http", agent_center.ListAgentCenterHttp)
		listGroup.GET("/manager", manager.ListManager)
	}

	serviceGroup := r.Group("/service")
	{
		serviceGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
		serviceGroup.POST("sync", service.SyncInfo)
	}

	if err := r.Run(fmt.Sprintf(":%v", listenPort)); err != nil {
		panic(fmt.Sprintf("服务中心启动监听失败,失败原因:%v", err))
	}
}
