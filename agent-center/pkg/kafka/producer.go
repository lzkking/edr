package kafka

import (
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Producer struct {
	Topic    string
	Producer sarama.AsyncProducer
}

var (
	MQMsgPool = &sync.Pool{
		New: func() interface{} {
			return &pb.MQData{}
		},
	}

	mqProducerMessagePool = &sync.Pool{
		New: func() interface{} {
			return &sarama.ProducerMessage{}
		},
	}
)

func NewProducerWithLogger(kafkaAdders []string, topic string, clientID string, logPath string, enableAuth bool, userName, password string) (*Producer, error) {
	logger, _ := createZapFileLogger(logPath)
	sarama.Logger = zap.NewStdLog(logger)

	config := sarama.NewConfig()
	config.ClientID = clientID
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = 1024 * 1024 * 4
	config.Producer.Timeout = 6 * time.Second
	config.Producer.Flush.Bytes = 1024 * 1024 * 4
	config.Producer.Flush.MaxMessages = 1024 * 1024 * 4
	config.Producer.Flush.Frequency = 10 * time.Second

	if enableAuth {
		config.Net.SASL.User = userName
		config.Net.SASL.Password = password
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.Enable = enableAuth
	}
	return newProducerWithConfig(kafkaAdders, topic, config)
}

func NewProducer(kafkaAdders []string, topic string, clientID string, logPath string, enableAuth bool, userName, password string) (*Producer, error) {
	config := sarama.NewConfig()
	config.ClientID = clientID
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = 1024 * 1024 * 4
	config.Producer.Timeout = 6 * time.Second
	config.Producer.Flush.Bytes = 1024 * 1024 * 4
	config.Producer.Flush.MaxMessages = 1024 * 1024 * 4
	config.Producer.Flush.Frequency = 10 * time.Second

	if enableAuth {
		config.Net.SASL.User = userName
		config.Net.SASL.Password = password
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.Enable = enableAuth
	}

	return newProducerWithConfig(kafkaAdders, topic, config)
}

func newProducerWithConfig(kafkaAdders []string, topic string, config *sarama.Config) (*Producer, error) {
	client, err := sarama.NewClient(kafkaAdders, config)
	if err != nil {
		zap.S().Warnf("连接kafka失败，失败原因:%v", err)
		return nil, err
	}

	producer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		zap.S().Warnf("创建kafka的生产者失败失败原因:%v", err)
		return nil, err
	}

	go func() {
		select {
		case success := <-producer.Successes():

		}
	}()
}
