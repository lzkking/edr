package route

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/config"
	"github.com/lzkking/edr/service/internal/handler/agent_center"
	"github.com/lzkking/edr/service/internal/handler/manager"
	"net/http"
	"sync"
)

func StartUp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	listenPort := config.GetServerConfig().ListenPort
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})

	})

	registerGroup := r.Group("/register")
	{
		registerGroup.POST("/agent-center-grpc", agent_center.AgentCenterGrpcRegister)
		registerGroup.POST("/agent-center-http", agent_center.AgentCenterHttpRegister)
		registerGroup.POST("/manager", manager.ManagerRegister)
	}

	listGroup := r.Group("/list")
	{
		listGroup.POST("/agent-center-grpc", agent_center.ListAgentCenterGrpc)
		listGroup.POST("/agent-center-http", agent_center.ListAgentCenterHttp)
		listGroup.POST("/manager", manager.ListManager)
	}

	if err := r.Run(fmt.Sprintf(":%v", listenPort)); err != nil {
		panic(fmt.Sprintf("服务中心启动监听失败,失败原因:%v", err))
	}
}
