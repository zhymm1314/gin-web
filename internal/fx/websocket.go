package fx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"gin-web/app/controllers"
	"gin-web/config"
	"gin-web/pkg/websocket"
)

// WebSocketModule WebSocket 模块（条件加载）
func WebSocketModule(enabled bool) fx.Option {
	if !enabled {
		return fx.Options() // 空模块
	}

	return fx.Module("websocket",
		fx.Provide(ProvideWebSocketManager),
		fx.Provide(
			fx.Annotate(
				ProvideWebSocketController,
				fx.ResultTags(`group:"controllers"`),
			),
		),
	)
}

// ProvideWebSocketManager 提供 WebSocket 管理器
func ProvideWebSocketManager(
	lc fx.Lifecycle,
	cfg *config.Configuration,
	log *zap.Logger,
) *websocket.Manager {
	manager := websocket.NewManager(log)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("WebSocket manager started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing WebSocket manager")
			manager.Close()
			return nil
		},
	})

	return manager
}

// ProvideWebSocketController 提供 WebSocket 控制器
func ProvideWebSocketController(manager *websocket.Manager) controllers.Controller {
	return controllers.NewWebSocketController(manager)
}
