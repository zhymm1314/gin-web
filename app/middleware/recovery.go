package middleware

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"

	"gin-web/app/dto"
	"gin-web/config"
)

// RecoveryMiddleware 恢复中间件（支持依赖注入）
type RecoveryMiddleware struct {
	cfg *config.Configuration
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(cfg *config.Configuration) *RecoveryMiddleware {
	return &RecoveryMiddleware{cfg: cfg}
}

// Handler 返回 gin.HandlerFunc
func (m *RecoveryMiddleware) Handler() gin.HandlerFunc {
	return gin.RecoveryWithWriter(
		&lumberjack.Logger{
			Filename:   m.cfg.Log.RootDir + "/" + m.cfg.Log.Filename,
			MaxSize:    m.cfg.Log.MaxSize,
			MaxBackups: m.cfg.Log.MaxBackups,
			MaxAge:     m.cfg.Log.MaxAge,
			Compress:   m.cfg.Log.Compress,
		},
		dto.ServerError)
}

// CustomRecovery 创建恢复中间件（兼容旧代码，使用默认配置）
// Deprecated: 请使用 NewRecoveryMiddleware 并通过依赖注入获取配置
func CustomRecovery(cfg *config.Configuration) gin.HandlerFunc {
	return gin.RecoveryWithWriter(
		&lumberjack.Logger{
			Filename:   cfg.Log.RootDir + "/" + cfg.Log.Filename,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		},
		dto.ServerError)
}
