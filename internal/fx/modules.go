package fx

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"gin-web/config"
)

// NewApp 创建主应用（API 服务）
func NewApp() *fx.App {
	// 预加载配置以决定可选模块
	cfg, err := ProvideConfig()
	if err != nil {
		panic(err)
	}

	return fx.New(
		// 基础设施
		InfrastructureModule,

		// 业务层
		RepositoryModule,
		ServiceModule,
		MiddlewareModule,
		ControllerModule,

		// HTTP 路由
		RouterModule,

		// 启动信息
		BannerModule,

		// 可选模块（根据配置动态加载）
		RabbitMQModule(cfg.RabbitMQ.Enable),
		CronModule(cfg.Cron.Enable),
		WebSocketModule(cfg.WebSocket.Enable),

		// 禁用 fx 的 verbose 日志
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
	)
}

// NewConsumerApp 创建消费者应用
func NewConsumerApp() *fx.App {
	return fx.New(
		// 基础设施
		InfrastructureModule,

		// RabbitMQ 消费者（强制启用）
		RabbitMQModule(true),

		// 禁用 fx 的 verbose 日志
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
	)
}

// NewCronApp 创建定时任务应用
func NewCronApp() *fx.App {
	return fx.New(
		// 基础设施
		InfrastructureModule,

		// 业务层（定时任务可能需要）
		RepositoryModule,
		ServiceModule,

		// 定时任务（强制启用）
		CronModule(true),

		// 禁用 fx 的 verbose 日志
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
	)
}

// NewWebSocketApp 创建 WebSocket 应用
func NewWebSocketApp() *fx.App {
	return fx.New(
		// 基础设施
		InfrastructureModule,

		// 业务层
		RepositoryModule,
		ServiceModule,
		MiddlewareModule,

		// WebSocket 模块（强制启用）
		WebSocketModule(true),

		// HTTP 路由
		RouterModule,

		// 禁用 fx 的 verbose 日志
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),

		// 覆盖端口配置
		fx.Decorate(func(c *config.Configuration) *config.Configuration {
			if c.WebSocket.Port != "" {
				c.App.Port = c.WebSocket.Port
			}
			return c
		}),
	)
}
