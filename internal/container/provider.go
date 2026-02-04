package container

import (
	"context"
	"gin-web/app/controllers"
	"gin-web/app/middleware"
	"gin-web/app/services"
	"gin-web/config"
	"gin-web/global"
	"gin-web/internal/repository"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// ========== 基础设施 Provider ==========

// ProvideDB 提供数据库连接
func ProvideDB() *gorm.DB {
	return global.App.DB
}

// ProvideRedis 提供 Redis 客户端
func ProvideRedis() *redis.Client {
	return global.App.Redis
}

// ProvideLog 提供日志实例
func ProvideLog() *zap.Logger {
	return global.App.Log
}

// ProvideConfig 提供配置
func ProvideConfig() *config.Configuration {
	return &global.App.Config
}

// ========== Repository Provider ==========

// ProvideUserRepository 提供用户仓储
func ProvideUserRepository(db *gorm.DB) repository.UserRepository {
	return repository.NewUserRepository(db)
}

// ProvideModRepository 提供 Mod 仓储
func ProvideModRepository(db *gorm.DB) repository.ModRepository {
	return repository.NewModRepository(db)
}

// ========== 适配器 Provider ==========

// ProvideJwtConfig 提供 JWT 配置适配器
func ProvideJwtConfig(cfg *config.Configuration) services.JwtConfig {
	return &JwtConfigAdapter{config: cfg}
}

// ProvideRedisClient 提供 Redis 客户端适配器
func ProvideRedisClient(client *redis.Client) services.RedisClient {
	return &RedisAdapter{client: client}
}

// ProvideUserGetter 提供用户获取适配器
func ProvideUserGetter(userService *services.UserService) services.UserGetter {
	return &UserGetterAdapter{userService: userService}
}

// ========== Service Provider ==========

// ProvideUserService 提供用户服务
func ProvideUserService(repo repository.UserRepository, log *zap.Logger) *services.UserService {
	return services.NewUserService(repo, log)
}

// ProvideJwtService 提供 JWT 服务
func ProvideJwtService(jwtConfig services.JwtConfig, redisClient services.RedisClient, userGetter services.UserGetter) *services.JwtService {
	return services.NewJwtService(jwtConfig, redisClient, userGetter)
}

// ProvideModService 提供 Mod 服务
func ProvideModService(repo repository.ModRepository, db *gorm.DB, log *zap.Logger) *services.ModService {
	return services.NewModService(repo, db, log)
}

// ========== Middleware Provider ==========

// ProvideJwtMiddleware 提供 JWT 中间件
func ProvideJwtMiddleware(jwtService *services.JwtService) *middleware.JwtMiddleware {
	return middleware.NewJwtMiddleware(jwtService)
}

// ========== Controller Provider ==========

// ProvideAuthController 提供认证控制器
func ProvideAuthController(userService *services.UserService, jwtService *services.JwtService, jwtMiddleware *middleware.JwtMiddleware) *controllers.AuthController {
	return controllers.NewAuthController(userService, jwtService, jwtMiddleware)
}

// ProvideModController 提供 Mod 控制器
func ProvideModController(modService *services.ModService) *controllers.ModController {
	return controllers.NewModController(modService)
}

// ========== 适配器实现 ==========

// JwtConfigAdapter JWT配置适配器
type JwtConfigAdapter struct {
	config *config.Configuration
}

func (a *JwtConfigAdapter) GetSecret() string {
	return a.config.Jwt.Secret
}

func (a *JwtConfigAdapter) GetTtl() int64 {
	return a.config.Jwt.JwtTtl
}

func (a *JwtConfigAdapter) GetBlacklistGracePeriod() int64 {
	return a.config.Jwt.JwtBlacklistGracePeriod
}

func (a *JwtConfigAdapter) GetRefreshGracePeriod() int64 {
	return a.config.Jwt.RefreshGracePeriod
}

// RedisAdapter Redis客户端适配器
type RedisAdapter struct {
	client *redis.Client
}

func (a *RedisAdapter) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.client.SetNX(ctx, key, value, expiration).Err()
}

func (a *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.client.Get(ctx, key).Result()
}

// UserGetterAdapter 用户获取适配器
type UserGetterAdapter struct {
	userService *services.UserService
}

func (a *UserGetterAdapter) GetUserInfo(id string) (services.JwtUser, error) {
	return a.userService.GetUserInfo(id)
}
