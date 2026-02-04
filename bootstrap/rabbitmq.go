package bootstrap

import (
	"gin-web/app/amqp/consumer"
	"gin-web/config"
	"gin-web/global"
	"gin-web/pkg/rabbitmq"
	"log"
)

// InitRabbitmq 初始化 RabbitMQ 消费者并返回管理器
func InitRabbitmq() *rabbitmq.Manager {
	// 加载消费者配置
	cfgConsumer, err := config.LoadConfig("./config/yaml/consumer.yaml")
	if err != nil {
		log.Printf("Failed to load consumer config: %v", err)
		return nil
	}

	// 注册消费者处理器
	handlers := map[string]consumer.ConsumerHandler{
		"LogConsumer": &consumer.LogConsumer{},
		// 添加更多消费者处理器...
	}

	// 构建 RabbitMQ 配置
	cfg := &rabbitmq.Config{
		Host:              global.App.Config.RabbitMQ.Host,
		Port:              global.App.Config.RabbitMQ.Port,
		Username:          global.App.Config.RabbitMQ.Username,
		Password:          global.App.Config.RabbitMQ.Password,
		Vhost:             global.App.Config.RabbitMQ.Vhost,
		ReconnectInterval: 5,
	}

	// 转换消费者配置
	consumers := make([]rabbitmq.ConsumerConfig, len(cfgConsumer.Consumers))
	for i, c := range cfgConsumer.Consumers {
		consumers[i] = rabbitmq.ConsumerConfig{
			Queue:       c.Queue,
			Handler:     c.Handler,
			Concurrency: c.Concurrency,
		}
	}

	// 创建消费者管理器
	manager := rabbitmq.NewManager(cfg, consumers, handlers, global.App.Log)
	go manager.Start()

	return manager
}
