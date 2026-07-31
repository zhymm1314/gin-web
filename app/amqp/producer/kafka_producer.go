package producer

import (
	"gin-web/config"
	"gin-web/pkg/kafka"
)

// KafkaBaseProducer Kafka 生产者基类，实现 mq.Producer。
// 与 RabbitMQ 的 BaseProducer 对称，便于按配置切换底层 MQ。
type KafkaBaseProducer struct {
	*kafka.Producer
}

// NewKafkaBaseProducer 创建 Kafka 生产者基类。
func NewKafkaBaseProducer(cfg config.Kafka, topic string) (*KafkaBaseProducer, error) {
	kcfg := &kafka.Config{
		Brokers:       cfg.Brokers,
		Version:       cfg.Version,
		Username:      cfg.Username,
		Password:      cfg.Password,
		GroupID:       cfg.GroupID,
		InitialOffset: cfg.InitialOffset,
	}
	p, err := kafka.NewProducer(kcfg, topic)
	if err != nil {
		return nil, err
	}
	return &KafkaBaseProducer{Producer: p}, nil
}
