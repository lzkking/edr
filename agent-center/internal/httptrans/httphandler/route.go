package httphandler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Route(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/agent_list", GetAgentList)
	r.POST("/post_command", PostCommand)
	r.POST("/have_agent", HaveAgent)
	r.POST("/command", PostCommand)

	return
}
