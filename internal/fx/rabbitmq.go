package fx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"gin-web/app/amqp/consumer"
	appConfig "gin-web/config"
	"gin-web/pkg/rabbitmq"
)

// RabbitMQModule RabbitMQ 模块（条件加载）
func RabbitMQModule(enabled bool) fx.Option {
	if !enabled {
		return fx.Options() // 空模块
	}

	return fx.Module("rabbitmq",
		fx.Provide(ProvideConsumerHandlers),
		fx.Provide(ProvideRabbitMQManager),
		fx.Invoke(StartRabbitMQ),
	)
}

// ProvideConsumerHandlers 提供消费者处理器
func ProvideConsumerHandlers() map[string]consumer.ConsumerHandler {
	return map[string]consumer.ConsumerHandler{
		"LogConsumer": &consumer.LogConsumer{},
		// 在这里注册更多消费者处理器
	}
}

// ProvideRabbitMQManager 提供 RabbitMQ 管理器
func ProvideRabbitMQManager(
	lc fx.Lifecycle,
	cfg *appConfig.Configuration,
	log *zap.Logger,
	handlers map[string]consumer.ConsumerHandler,
) (*rabbitmq.Manager, error) {
	// 加载消费者配置
	consumerCfg, err := appConfig.LoadConfig("./config/yaml/consumer.yaml")
	if err != nil {
		log.Warn("load consumer config failed, using empty config", zap.Error(err))
		consumerCfg = &appConfig.AppConfig{
			Consumers: []appConfig.ConsumerConfig{},
		}
	}

	// 转换消费者配置
	consumers := make([]rabbitmq.ConsumerConfig, len(consumerCfg.Consumers))
	for i, c := range consumerCfg.Consumers {
		consumers[i] = rabbitmq.ConsumerConfig{
			Queue:       c.Queue,
			Handler:     c.Handler,
			Concurrency: c.Concurrency,
		}
	}

	// 创建 RabbitMQ 配置
	managerCfg := &rabbitmq.Config{
		Host:              cfg.RabbitMQ.Host,
		Port:              cfg.RabbitMQ.Port,
		Username:          cfg.RabbitMQ.Username,
		Password:          cfg.RabbitMQ.Password,
		Vhost:             cfg.RabbitMQ.Vhost,
		ReconnectInterval: 5,
	}

	manager := rabbitmq.NewManager(managerCfg, consumers, handlers, log)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting RabbitMQ consumer manager")
			go manager.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping RabbitMQ consumer manager")
			manager.Stop()
			return nil
		},
	})

	return manager, nil
}

// StartRabbitMQ 启动 RabbitMQ（触发依赖注入）
func StartRabbitMQ(_ *rabbitmq.Manager) {
	// manager 会通过 lifecycle 启动
}
