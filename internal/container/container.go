package container

import (
	"context"
	"gin-web/app/controllers"
	"gin-web/app/services"
	"gin-web/config"
	"gin-web/global"
	"gin-web/internal/repository"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// Container 依赖注入容器
type Container struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Log    *zap.Logger
	Config *config.Configuration

	// Repositories
	UserRepo repository.UserRepository

	// Services
	UserService *services.UserService
	JwtService  *services.JwtServiceDI

	// Controllers
	AuthController *controllers.AuthController
}

// NewContainer 创建容器实例
func NewContainer() *Container {
	c := &Container{
		DB:     global.App.DB,
		Redis:  global.App.Redis,
		Log:    global.App.Log,
		Config: &global.App.Config,
	}

	// 初始化各层依赖
	c.initRepositories()
	c.initServices()
	c.initControllers()

	return c
}

// initRepositories 初始化仓储层
func (c *Container) initRepositories() {
	c.UserRepo = repository.NewUserRepository(c.DB)
}

// initServices 初始化服务层
func (c *Container) initServices() {
	c.UserService = services.NewUserService(c.UserRepo, c.Log)
	c.JwtService = services.NewJwtServiceDI(
		NewJwtConfigAdapter(c.Config),
		NewRedisAdapter(c.Redis),
	)
}

// initControllers 初始化控制器层
func (c *Container) initControllers() {
	c.AuthController = controllers.NewAuthController(c.UserService, c.JwtService)
}

// GetControllers 获取所有控制器
func (c *Container) GetControllers() []controllers.Controller {
	return []controllers.Controller{
		c.AuthController,
	}
}

// ========== 适配器实现 ==========

// JwtConfigAdapter JWT配置适配器
type JwtConfigAdapter struct {
	config *config.Configuration
}

func NewJwtConfigAdapter(cfg *config.Configuration) *JwtConfigAdapter {
	return &JwtConfigAdapter{config: cfg}
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

func NewRedisAdapter(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (a *RedisAdapter) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.client.SetNX(ctx, key, value, expiration).Err()
}

func (a *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.client.Get(ctx, key).Result()
}
