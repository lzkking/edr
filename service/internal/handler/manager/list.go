package manager

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"net/http"
)

func ListManager(c *gin.Context) {
	managerHosts := endpoint.EI.GetGreenHosts(endpoint.TypeManager)
	c.JSON(http.StatusOK, gin.H{"hosts": managerHosts})
}
