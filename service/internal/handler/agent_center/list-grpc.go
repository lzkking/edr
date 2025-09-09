package agent_center

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"net/http"
)

func ListAgentCenterGrpc(c *gin.Context) {
	agentCenterHosts := endpoint.EI.GetGreenHosts(endpoint.TypeAgentCenterGrpc)
	c.JSON(http.StatusOK, gin.H{"hosts": agentCenterHosts})
}
