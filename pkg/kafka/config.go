package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

// Config Kafka 配置
type Config struct {
	Brokers       []string
	Version       string // 如 "3.6.0.0"，留空自动协商
	Username      string // SASL 用户名（无认证留空）
	Password      string // SASL 密码（无认证留空）
	GroupID       string // 默认消费组（consumer 未单独配置时使用）
	InitialOffset string // 新消费组起始 offset：newest(默认) / oldest
}

// buildSaramaConfig 根据配置构造 *sarama.Config。
func buildSaramaConfig(cfg *Config) (*sarama.Config, error) {
	sc := sarama.NewConfig()
	sc.ClientID = "gin-web"

	// 版本
	if cfg.Version != "" {
		v, err := sarama.ParseKafkaVersion(cfg.Version)
		if err != nil {
			return nil, fmt.Errorf("parse kafka version %q: %w", cfg.Version, err)
		}
		sc.Version = v
	}

	// SASL 认证
	if cfg.Username != "" {
		sc.Net.SASL.Enable = true
		sc.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		sc.Net.SASL.User = cfg.Username
		sc.Net.SASL.Password = cfg.Password
	}

	// 消费者
	sc.Consumer.Return.Errors = true
	switch cfg.InitialOffset {
	case "oldest", "earliest":
		sc.Consumer.Offsets.Initial = sarama.OffsetOldest
	default: // 留空或 newest
		sc.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	if err := sc.Validate(); err != nil {
		return nil, fmt.Errorf("validate sarama config: %w", err)
	}
	return sc, nil
}
