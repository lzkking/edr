package httptrans

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/agent-center/config"
	"github.com/lzkking/edr/agent-center/internal/httptrans/httphandler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func Run() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	gin.DefaultWriter = zapcore.AddSync(os.Stdout)
	gin.DefaultErrorWriter = zapcore.AddSync(os.Stderr)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	httphandler.Route(r)

	httpConfig := config.GetServerConfig()

	if err := r.Run(fmt.Sprintf(":%v", httpConfig.HttpListenPort)); err != nil {
		zap.S().Errorf("开启http监听失败,失败原因:%v", err)
		os.Exit(-2)
	}
}
