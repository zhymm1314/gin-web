// config/queue.go
package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type SingleQueueConfig struct {
	WorkerNum  int    `mapstructure:"worker_num"`
	Name       string `mapstructure:"name"`
	RetryTimes int    `mapstructure:"retry_times"`
}

type QueueConfig struct {
	Queues map[string]SingleQueueConfig `yaml:"queues"`
}

func LoadQueueConfig() (*QueueConfig, error) {

	// 设置配置文件路径
	config := "./config/yaml/queue.yaml"
	// 生产环境可以通过设置环境变量来改变配置文件路径
	if configEnv := os.Getenv("VIPER_CONFIG"); configEnv != "" {
		config = configEnv
	}

	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	// 解析配置到结构体
	var cfg QueueConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &cfg, nil
}
