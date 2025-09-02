package httphandler

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/handler"
	"net/http"
)

func GetAgentList(c *gin.Context) {
	agentLists := handler.GlobalPool.GetAgentIdList()
	c.JSON(http.StatusOK, gin.H{
		"agent list": agentLists,
	})
}
