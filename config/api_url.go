package config

// ApiUrls API 地址配置
type ApiUrls struct {
	UserService  ServiceEndpoint `yaml:"user_service" mapstructure:"user_service"`
	OrderService ServiceEndpoint `yaml:"order_service" mapstructure:"order_service"`
	PayService   ServiceEndpoint `yaml:"pay_service" mapstructure:"pay_service"`
}

// ServiceEndpoint 服务端点配置
type ServiceEndpoint struct {
	Local      string `yaml:"local" mapstructure:"local"`
	Dev        string `yaml:"dev" mapstructure:"dev"`
	Test       string `yaml:"test" mapstructure:"test"`
	Production string `yaml:"production" mapstructure:"production"`
}

// GetURL 根据环境获取 URL
func (s ServiceEndpoint) GetURL(env string) string {
	switch env {
	case "local":
		return s.Local
	case "dev":
		return s.Dev
	case "test":
		return s.Test
	case "production":
		return s.Production
	default:
		return s.Local
	}
}
