package config

// WebSocket WebSocket 配置
type WebSocket struct {
	Enable         bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
	Port           string `mapstructure:"port" json:"port" yaml:"port"`
	MaxConnections int    `mapstructure:"max_connections" json:"max_connections" yaml:"max_connections"`
}
