package agent_center

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"go.uber.org/zap"
	"net/http"
)

// AgentCenterGrpcRegister - 处理agent center的注册函数
func AgentCenterGrpcRegister(c *gin.Context) {
	agentCenterRegisterInfo := endpoint.RegisterInfo{}
	err := c.ShouldBindJSON(&agentCenterRegisterInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request data"})
		zap.S().Warnf("绑定到json失败,失败原因:%v", err)
		return
	}
	agentCenterRegisterInfo.Type = endpoint.TypeAgentCenterGrpc

	endpoint.EI.Register(agentCenterRegisterInfo)

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
