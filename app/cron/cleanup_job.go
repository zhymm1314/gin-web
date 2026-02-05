package cron

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CleanupJob 清理过期数据任务
type CleanupJob struct {
	db    *gorm.DB
	redis *redis.Client
	log   *zap.Logger
}

// NewCleanupJob 创建清理任务（通过 fx 注入依赖）
func NewCleanupJob(db *gorm.DB, redis *redis.Client, log *zap.Logger) *CleanupJob {
	return &CleanupJob{
		db:    db,
		redis: redis,
		log:   log,
	}
}

// Name 返回任务名称
func (j *CleanupJob) Name() string {
	return "cleanup_expired_tokens"
}

// Spec 返回 cron 表达式 (每天凌晨 2 点执行)
func (j *CleanupJob) Spec() string {
	return "0 0 2 * * *"
}

// Run 执行清理逻辑
func (j *CleanupJob) Run() {
	startTime := time.Now()
	j.log.Info("cleanup job started")

	// 清理过期的 JWT 黑名单
	if j.redis != nil {
		// 示例：清理过期的 token 黑名单
		// 实际实现根据业务需求
		ctx := context.Background()
		_ = ctx // 使用 redis 清理过期数据
	}

	// 清理数据库中的过期数据
	if j.db != nil {
		// 示例：清理 30 天前的软删除数据
		// j.db.Unscoped().Where("deleted_at < ?", time.Now().AddDate(0, 0, -30)).Delete(&models.SomeModel{})
	}

	j.log.Info("cleanup job completed",
		zap.String("job", j.Name()),
		zap.Duration("duration", time.Since(startTime)),
	)
}
