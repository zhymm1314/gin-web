package kafka

import (
	"github.com/IBM/sarama"
)

// Producer 基于 sarama.SyncProducer 的 Kafka 生产者，实现 mq.Producer。
type Producer struct {
	sync  sarama.SyncProducer
	topic string
}

// NewProducer 创建 Kafka 生产者。
func NewProducer(cfg *Config, topic string) (*Producer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true // SyncProducer 必须开启
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.ClientID = "gin-web"

	if cfg.Version != "" {
		if v, err := sarama.ParseKafkaVersion(cfg.Version); err == nil {
			saramaCfg.Version = v
		}
	}
	if cfg.Username != "" {
		saramaCfg.Net.SASL.Enable = true
		saramaCfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		saramaCfg.Net.SASL.User = cfg.Username
		saramaCfg.Net.SASL.Password = cfg.Password
	}

	if err := saramaCfg.Validate(); err != nil {
		return nil, err
	}

	sp, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, err
	}
	return &Producer{sync: sp, topic: topic}, nil
}

// Publish 发布一条消息。
func (p *Producer) Publish(body []byte) error {
	_, _, err := p.sync.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(body),
	})
	return err
}

// Topic 返回目标主题。
func (p *Producer) Topic() string { return p.topic }

// Close 关闭生产者。
func (p *Producer) Close() error {
	return p.sync.Close()
}
