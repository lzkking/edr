package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lzkking/edr/manager/docs" // docs 包是 swag 生成的
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

// @title Example API
// @version 1.0
// @description 示例 Go + Swagger API
// @host 10.18.201.56:8989
// @BasePath /api
func main() {
	r := gin.Default()

	r.GET("/api/ping", PingHandler)

	// Swagger UI 路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8989")
}

// PingHandler godoc
// @Summary ping 测试
// @Description 测试接口
// @Tags example
// @Success 200 {string} string "pong"
// @Router /ping [get]
func PingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
