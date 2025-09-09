package service

import (
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/service/internal/endpoint"
	"go.uber.org/zap"
	"net/http"
)

func SyncInfo(c *gin.Context) {
	var transInfo endpoint.TransInfo
	err := c.ShouldBindJSON(&transInfo)
	if err != nil {
		zap.S().Errorf("bind json failed,失败原因:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err = endpoint.EI.Recv(transInfo)
	if err != nil {
		zap.S().Warnf("处理接收到的同步信息失败,失败原因:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad sync"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	return
}
