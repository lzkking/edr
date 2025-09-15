package kafka

import (
	"context"
	"github.com/IBM/sarama"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

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

type Consumer struct {
	Topics   []string
	Consumer sarama.ConsumerGroup
	GroupID  string
	ClientID string
}

type MessageHandler func(ctx context.Context, m *pb.MQData) error

func baseConsumerConfig(clientID string, enableAuth bool, userName, password string) *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.ClientID = clientID

	// 建议显式设置 Version（与集群匹配或略低）
	// 也可改为 sarama.V3_6_0 / V2_8_0_0 等
	cfg.Version = sarama.V2_6_0_0

	// —— Consumer 关键参数 —— //
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Group.Session.Timeout = 30 * time.Second
	cfg.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	cfg.Consumer.Fetch.Min = 1
	cfg.Consumer.MaxProcessingTime = 500 * time.Millisecond

	// —— SASL —— //
	if enableAuth {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = userName
		cfg.Net.SASL.Password = password
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	// —— 网络健壮性 —— //
	cfg.Net.DialTimeout = 10 * time.Second
	cfg.Net.ReadTimeout = 30 * time.Second
	cfg.Net.WriteTimeout = 30 * time.Second
	cfg.Metadata.Retry.Max = 6
	cfg.Metadata.Retry.Backoff = 2 * time.Second

	return cfg
}

func NewConsumerWithLogger(
	brokers []string,
	groupID string,
	topics []string,
	clientID string,
	logPath string,
	enableAuth bool,
	userName, password string,
) (*Consumer, error) {
	logger, _ := createZapFileLogger(logPath)
	sarama.Logger = zap.NewStdLog(logger)
	cfg := baseConsumerConfig(clientID, enableAuth, userName, password)

	cli, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		zap.S().Warnf("连接kafka失败: %v", err)
		return nil, err
	}

	group, err := sarama.NewConsumerGroupFromClient(groupID, cli)
	if err != nil {
		zap.S().Warnf("创建ConsumerGroup失败: %v", err)
		return nil, err
	}

	return &Consumer{
		Consumer: group,
		Topics:   topics,
		ClientID: clientID,
		GroupID:  groupID,
	}, nil
}

// ===========================
// GroupHandler 实现
// ===========================

type cgHandler struct {
	handle MessageHandler
}

func (h *cgHandler) Setup(s sarama.ConsumerGroupSession) error {
	zap.S().Info("ConsumerGroup Setup")
	return nil
}
func (h *cgHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	zap.S().Info("ConsumerGroup Cleanup")
	return nil
}

func (h *cgHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// 反序列化 protobuf
		pbm := MQMsgPool.Get().(*pb.MQData)
		if err := proto.Unmarshal(msg.Value, pbm); err != nil {
			zap.S().Errorf("解码proto失败: %v, topic=%s partition=%d offset=%d",
				err, msg.Topic, msg.Partition, msg.Offset)
			MQMsgPool.Put(pbm)
			// 不标记，保留给下次消费或移交死信/重试策略
			continue
		}

		// 处理
		if err := h.handle(sess.Context(), pbm); err != nil {
			zap.S().Errorf("业务处理失败: %v, key=%s topic=%s partition=%d offset=%d",
				err, string(msg.Key), msg.Topic, msg.Partition, msg.Offset)
			MQMsgPool.Put(pbm)
			// 失败不 Mark，留待重试（或你在 handle 里将其投递到DLQ）
			continue
		}

		// 成功才提交 offset（幂等/可重入的处理是底线）
		sess.MarkMessage(msg, "")
		MQMsgPool.Put(pbm)
	}
	return nil
}

// ===========================
// 启动 & 优雅退出
// ===========================

func (c *Consumer) Start(ctx context.Context, handler MessageHandler) error {
	h := &cgHandler{handle: handler}

	// 捕获退出信号，优雅关闭
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch
		zap.S().Info("收到退出信号，停止消费...")
		cancel()
	}()

	// 错误监听
	go func() {
		for err := range c.Consumer.Errors() {
			zap.S().Errorf("ConsumerGroup错误: %v", err)
		}
	}()

	// Rebalance 可能导致 Consume 返回，需要循环重进
	for {
		zap.S().Debugf("消费数据")
		if err := c.Consumer.Consume(ctx, c.Topics, h); err != nil {
			zap.S().Errorf("Consume失败: %v", err)
			// 小退后再试，避免打满日志
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
