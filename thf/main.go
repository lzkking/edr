package main

import (
	"context"
	"github.com/lzkking/edr/thf/config"
	"github.com/lzkking/edr/thf/internal/consumer"
	"github.com/lzkking/edr/thf/log"
	"github.com/lzkking/edr/thf/pkg/kafka"
	"go.uber.org/zap"
)

func main() {
	log.Init()
	zap.S().Debugf("数据处理中心")

	kafkaConfig := config.GetServerConfig().KafkaConfig
	c, err := kafka.NewConsumerWithLogger(
		kafkaConfig.KafkaAdders,
		kafkaConfig.GroupID,
		kafkaConfig.Topics,
		kafkaConfig.ClientID,
		kafkaConfig.LogPath,
		kafkaConfig.EnableAuth,
		kafkaConfig.UserName,
		kafkaConfig.Password,
	)
	if err != nil {
		panic(err)
	}

	zap.S().Infof("开始消费kafka中的数据")

	if err = c.Start(context.Background(), consumer.CustomerData); err != nil {
		zap.S().Errorf("消费kafka的数据失败")
	}

}
