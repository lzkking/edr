package log

import (
	"github.com/lzkking/edr/manager/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

func Init() {
	serverConfig := config.GetServerConfig()

	fileEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   serverConfig.LogFile,
		MaxSize:    1,
		MaxBackups: 10,
		MaxAge:     10,
		Compress:   true,
	})

	//	替换全局日志记录器
	var core zapcore.Core
	if serverConfig.RunMode == "DEBUG" {
		//debug模式使用终端替换文件记录
		consoleWriter := zapcore.AddSync(os.Stdout)
		consoleConfig := zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			EncodeLevel:   zapcore.CapitalColorLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		}
		consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)
		core = zapcore.NewTee(zapcore.NewCore(consoleEncoder, consoleWriter, zap.DebugLevel))
	} else {
		core = zapcore.NewTee(
			zapcore.NewSamplerWithOptions(zapcore.NewCore(fileEncoder, fileWriter, zap.InfoLevel), time.Second, 4, 1))
	}

	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
}
