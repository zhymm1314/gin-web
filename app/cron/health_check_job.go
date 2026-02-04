package cron

import (
	"context"
	"gin-web/global"
)

// HealthCheckJob 健康检查任务
type HealthCheckJob struct{}

// Name 返回任务名称
func (j *HealthCheckJob) Name() string {
	return "health_check"
}

// Spec 返回 cron 表达式 (每 30 秒执行一次)
func (j *HealthCheckJob) Spec() string {
	return "*/30 * * * * *"
}

// Run 执行健康检查
func (j *HealthCheckJob) Run() {
	// 检查数据库连接
	if global.App.DB != nil {
		db, err := global.App.DB.DB()
		if err == nil {
			if err := db.Ping(); err != nil {
				global.App.Log.Warn("database ping failed")
			}
		}
	}

	// 检查 Redis 连接
	if global.App.Redis != nil {
		if err := global.App.Redis.Ping(context.Background()).Err(); err != nil {
			global.App.Log.Warn("redis ping failed")
		}
	}
}
