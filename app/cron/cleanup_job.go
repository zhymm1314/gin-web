package cron

import (
	"gin-web/global"

	"go.uber.org/zap"
)

// CleanupJob 清理过期数据任务
type CleanupJob struct{}

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
	global.App.Log.Info("cleanup job running")

	// 清理过期的 JWT 黑名单
	// 这里可以添加具体的清理逻辑
	// 例如: global.App.Redis.Del(context.Background(), "expired_keys...")

	global.App.Log.Info("cleanup job completed", zap.String("job", j.Name()))
}
