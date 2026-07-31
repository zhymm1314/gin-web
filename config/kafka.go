package config

// Kafka Kafka 消息队列配置
type Kafka struct {
	Enable        bool     `mapstructure:"enable" json:"enable" yaml:"enable"`
	Brokers       []string `mapstructure:"brokers" json:"brokers" yaml:"brokers"`
	Version       string   `mapstructure:"version" json:"version" yaml:"version"`             // Kafka 版本，如 "3.6.0.0"；留空则由客户端自动协商
	Username      string   `mapstructure:"username" json:"username" yaml:"username"`          // SASL 用户名（无认证留空）
	Password      string   `mapstructure:"password" json:"password" yaml:"password"`          // SASL 密码（无认证留空）
	GroupID       string   `mapstructure:"group_id" json:"group_id" yaml:"group_id"`          // 默认消费组（可在 consumer.yaml 中按 consumer 覆盖）
	InitialOffset string   `mapstructure:"initial_offset" json:"initial_offset" yaml:"initial_offset"` // 新消费组的起始 offset：newest(默认) / oldest
}
