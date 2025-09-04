package kafka

import "github.com/lzkking/edr/agent-center/config"

var KafkaProducer *Producer

func init() {
	kafkaConfig := config.GetServerConfig().KafkaConfig
	var err error
	if kafkaConfig.LogPath != "" {
		KafkaProducer, err = NewProducerWithLogger(
			kafkaConfig.KafkaAdders,
			kafkaConfig.Topic,
			kafkaConfig.ClientID,
			kafkaConfig.LogPath,
			kafkaConfig.EnableAuth,
			kafkaConfig.UserName,
			kafkaConfig.Password)
	} else {
		KafkaProducer, err = NewProducer(
			kafkaConfig.KafkaAdders,
			kafkaConfig.Topic,
			kafkaConfig.ClientID,
			kafkaConfig.LogPath,
			kafkaConfig.EnableAuth,
			kafkaConfig.UserName,
			kafkaConfig.Password)
	}

	if err != nil {
		panic("kafka 连接失败")
	}
}
