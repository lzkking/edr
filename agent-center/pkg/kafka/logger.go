package kafka

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func createZapFileLogger(logPath string) (*zap.Logger, error) {
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	})

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	core := zapcore.NewCore(fileEncoder, fileWriter, level)
	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}
