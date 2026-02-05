package cron

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthCheckJob 健康检查任务
type HealthCheckJob struct {
	db    *gorm.DB
	redis *redis.Client
	log   *zap.Logger
}

// NewHealthCheckJob 创建健康检查任务（通过 fx 注入依赖）
func NewHealthCheckJob(db *gorm.DB, redis *redis.Client, log *zap.Logger) *HealthCheckJob {
	return &HealthCheckJob{
		db:    db,
		redis: redis,
		log:   log,
	}
}

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查数据库连接
	if j.db != nil {
		sqlDB, err := j.db.DB()
		if err == nil {
			if err := sqlDB.PingContext(ctx); err != nil {
				j.log.Warn("database health check failed", zap.Error(err))
			}
		}
	}

	// 检查 Redis 连接
	if j.redis != nil {
		if err := j.redis.Ping(ctx).Err(); err != nil {
			j.log.Warn("redis health check failed", zap.Error(err))
		}
	}
}
