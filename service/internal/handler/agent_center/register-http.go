package agent_center

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"net/http"
)

// AgentCenterHttpRegister - 处理agent center的注册函数
func AgentCenterHttpRegister(c *gin.Context) {
	agentCenterRegisterInfo := endpoint.RegisterInfo{}
	err := c.ShouldBindJSON(agentCenterRegisterInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request data"})
		return
	}
}
