package agent_center

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"net/http"
)

func ListAgentCenterHttp(c *gin.Context) {
	agentCenterHosts := endpoint.EI.GetGreenHosts(endpoint.TypeAgentCenterHttp)
	c.JSON(http.StatusOK, gin.H{"hosts": agentCenterHosts})
}
