package fx

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	appCron "gin-web/app/cron"
	"gin-web/config"
	"gin-web/pkg/cron"
)

// CronModule 定时任务模块（条件加载）
func CronModule(enabled bool) fx.Option {
	if !enabled {
		return fx.Options() // 空模块
	}

	return fx.Module("cron",
		fx.Provide(ProvideCronManager),
		fx.Invoke(StartCron),
	)
}

// ProvideCronManager 提供定时任务管理器
func ProvideCronManager(
	lc fx.Lifecycle,
	cfg *config.Configuration,
	db *gorm.DB,
	redis *redis.Client,
	log *zap.Logger,
) *cron.Manager {
	manager := cron.NewManager(log)

	// 注册定时任务（通过构造函数注入依赖，已移除 global.App）
	manager.Register(appCron.NewCleanupJob(db, redis, log))
	manager.Register(appCron.NewHealthCheckJob(db, redis, log))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting cron manager")
			return manager.Start()
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping cron manager")
			manager.Stop()
			return nil
		},
	})

	return manager
}

// StartCron 启动定时任务（触发依赖注入）
func StartCron(_ *cron.Manager) {
	// manager 会通过 lifecycle 启动
}
