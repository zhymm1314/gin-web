package fx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"gin-web/app/amqp/consumer"
	appConfig "gin-web/config"
	"gin-web/pkg/kafka"
	"gin-web/pkg/mq"
	"gin-web/pkg/rabbitmq"
)

// MQModule 消息队列模块（条件加载，按配置选择 Kafka 或 RabbitMQ）。
// kafka.enable 优先；否则 rabbitmq.enable；两者均未启用时，force=true 回退 RabbitMQ，
// 否则不加载。同时启用时优先 Kafka 并在 provider 中告警。
func MQModule(cfg *appConfig.Configuration, force bool) fx.Option {
	enabled, useKafka := decideMQ(cfg, force)
	if !enabled {
		return fx.Options() // 空模块
	}

	if useKafka {
		return fx.Module("mq",
			fx.Provide(ProvideConsumerHandlers),
			fx.Provide(ProvideKafkaManager),
			fx.Invoke(StartMQ),
		)
	}

	return fx.Module("mq",
		fx.Provide(ProvideConsumerHandlers),
		fx.Provide(ProvideRabbitMQManager),
		fx.Invoke(StartMQ),
	)
}

// decideMQ 返回是否启用，以及是否使用 Kafka（同时启用时优先 Kafka）。
// force=true（消费者专用应用）时，若两者均未显式启用，回退到 RabbitMQ 以保持向后兼容；
// 如需切换到 Kafka，请在配置中显式设置 kafka.enable: true。
func decideMQ(cfg *appConfig.Configuration, force bool) (enabled, useKafka bool) {
	switch {
	case cfg.Kafka.Enable:
		return true, true
	case cfg.RabbitMQ.Enable:
		return true, false
	case force:
		// 消费者专用应用：两者均未显式启用时回退到 RabbitMQ（历史默认）
		return true, false
	default:
		return false, false
	}
}

// ProvideConsumerHandlers 提供消费者处理器
func ProvideConsumerHandlers() map[string]mq.Handler {
	return map[string]mq.Handler{
		"LogConsumer": &consumer.LogConsumer{},
		// 在这里注册更多消费者处理器
	}
}

// loadConsumerConfigs 加载 consumer.yaml 并转换为统一的消费者配置。
func loadConsumerConfigs(cfg *appConfig.Configuration, log *zap.Logger) []mq.ConsumerConfig {
	consumerCfg, err := appConfig.LoadConfig("./config/yaml/consumer.yaml")
	if err != nil {
		log.Warn("load consumer config failed, using empty config", zap.Error(err))
		return nil
	}

	consumers := make([]mq.ConsumerConfig, 0, len(consumerCfg.Consumers))
	for _, c := range consumerCfg.Consumers {
		topic := c.Topic
		if topic == "" {
			topic = c.Queue // 向后兼容旧配置
		}
		group := c.Group
		if group == "" {
			group = cfg.Kafka.GroupID
		}
		consumers = append(consumers, mq.ConsumerConfig{
			Topic:       topic,
			Group:       group,
			Handler:     c.Handler,
			Concurrency: c.Concurrency,
		})
	}
	return consumers
}

// ProvideKafkaManager 提供 Kafka 管理器
func ProvideKafkaManager(
	lc fx.Lifecycle,
	cfg *appConfig.Configuration,
	log *zap.Logger,
	handlers map[string]mq.Handler,
) (mq.Manager, error) {
	if cfg.RabbitMQ.Enable {
		log.Warn("both kafka and rabbitmq enabled, using kafka")
	}

	kcfg := &kafka.Config{
		Brokers:       cfg.Kafka.Brokers,
		Version:       cfg.Kafka.Version,
		Username:      cfg.Kafka.Username,
		Password:      cfg.Kafka.Password,
		GroupID:       cfg.Kafka.GroupID,
		InitialOffset: cfg.Kafka.InitialOffset,
	}

	manager := kafka.NewManager(kcfg, loadConsumerConfigs(cfg, log), handlers, log)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting Kafka consumer manager",
				zap.Strings("brokers", cfg.Kafka.Brokers))
			go manager.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping Kafka consumer manager")
			manager.Stop()
			return nil
		},
	})

	return manager, nil
}

// ProvideRabbitMQManager 提供 RabbitMQ 管理器
func ProvideRabbitMQManager(
	lc fx.Lifecycle,
	cfg *appConfig.Configuration,
	log *zap.Logger,
	handlers map[string]mq.Handler,
) (mq.Manager, error) {
	rcfg := &rabbitmq.Config{
		Host:              cfg.RabbitMQ.Host,
		Port:              cfg.RabbitMQ.Port,
		Username:          cfg.RabbitMQ.Username,
		Password:          cfg.RabbitMQ.Password,
		Vhost:             cfg.RabbitMQ.Vhost,
		ReconnectInterval: 5,
	}

	manager := rabbitmq.NewManager(rcfg, loadConsumerConfigs(cfg, log), handlers, log)

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

// StartMQ 启动消息队列（触发依赖注入）
func StartMQ(_ mq.Manager) {
	// manager 会通过 lifecycle 启动
}
