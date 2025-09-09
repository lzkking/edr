package manager

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"go.uber.org/zap"
	"net/http"
)

// ManagerRegister - 处理manager的注册信息
func ManagerRegister(c *gin.Context) {
	agentCenterRegisterInfo := endpoint.RegisterInfo{}
	err := c.ShouldBindJSON(&agentCenterRegisterInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request data"})
		zap.S().Warnf("绑定到json失败,失败原因:%v", err)
		return
	}

	agentCenterRegisterInfo.Type = endpoint.TypeManager

	endpoint.EI.Register(agentCenterRegisterInfo)

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
