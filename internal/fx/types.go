package fx

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-web/config"
)

// Infrastructure 基础设施容器（替代 global.App）
// 该结构体仅用于文档说明，实际使用时各组件通过 fx 独立注入
type Infrastructure struct {
	Config *config.Configuration
	DB     *gorm.DB
	Redis  *redis.Client
	Log    *zap.Logger
}
