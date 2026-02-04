package config

// Configuration 应用程序配置
type Configuration struct {
	App       App       `mapstructure:"app" json:"app" yaml:"app"`
	Log       Log       `mapstructure:"log" json:"log" yaml:"log"`
	Database  Database  `mapstructure:"database" json:"database" yaml:"database"`
	Jwt       Jwt       `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Redis     Redis     `mapstructure:"redis" json:"redis" yaml:"redis"`
	RabbitMQ  RabbitMQ  `mapstructure:"rabbitmq" json:"rabbitMQ" yaml:"rabbitMQ"`
	Cron      Cron      `mapstructure:"cron" json:"cron" yaml:"cron"`
	WebSocket WebSocket `mapstructure:"websocket" json:"websocket" yaml:"websocket"`
	ApiUrls   ApiUrls   `mapstructure:"api_url" json:"api_url" yaml:"api_url"`
}
