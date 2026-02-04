package config

// Cron 定时任务配置
type Cron struct {
	Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`
}
