package bootstrap

import (
	"fmt"
	"gin-web/global"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func InitializeConfig() *viper.Viper {
	// 设置配置文件路径
	config := "config.yaml"
	// 生产环境可以通过设置环境变量来改变配置文件路径
	if configEnv := os.Getenv("VIPER_CONFIG"); configEnv != "" {
		config = configEnv
	}

	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")

	//env := v.GetString("app.env")

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config file changed:", in.Name)
		// 重载配置
		if err := v.Unmarshal(&global.App.Config); err != nil {
			fmt.Println(err)
		}
	})
	// 将配置赋值给全局变量
	if err := v.Unmarshal(&global.App.Config); err != nil {
		fmt.Println(err)
	}
	return v
}

// 安全版本，带默认值

const (
	EnvDev  = "dev"
	EnvPre  = "pre"
	EnvProd = "prod"
)

func GetCurrentEnv() string {
	env := os.Getenv("APP_ENV")
	switch env {
	case EnvPre, EnvProd:
		return env
	default:
		return EnvDev
	}
}

// GetURL 根据环境和服务名动态获取配置中的 URL
// env: 环境参数（dev/pre/prod）
// service: 服务名称（如 "log_url", "xxx_url"）
func GetURL(env string, service string) string {
	// 构建配置路径（例如：api_url.xxx_url.dev）
	path := "api_url." + service + "." + env

	// 获取配置值，若不存在则返回空字符串
	url := viper.GetString(path)

	// 可选：添加默认值或错误处理
	if url == "" {
		// 可返回默认环境（如 dev）或抛出错误
		// url = viper.GetString("api_url." + service + ".dev")
		panic("Configuration not found for path: " + path)
	}

	return url
}
