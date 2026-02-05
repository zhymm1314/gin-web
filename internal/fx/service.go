package fx

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"gin-web/app/services"
	"gin-web/config"
	"gin-web/internal/repository"
)

// ServiceModule 服务模块
var ServiceModule = fx.Module("service",
	fx.Provide(
		ProvideUserService,
		ProvideJwtService,
		ProvideModService,
	),
)

// ProvideUserService 提供用户服务
func ProvideUserService(
	repo repository.UserRepository,
	log *zap.Logger,
) *services.UserService {
	return services.NewUserService(repo, log)
}

// ProvideJwtService 提供 JWT 服务
func ProvideJwtService(
	cfg *config.Configuration,
	redisClient *redis.Client,
	userSvc *services.UserService,
) *services.JwtService {
	return services.NewJwtService(
		&jwtConfigAdapter{cfg: cfg},
		&redisAdapter{client: redisClient},
		&userGetterAdapter{svc: userSvc},
	)
}

// ProvideModService 提供 Mod 服务
func ProvideModService(
	repo repository.ModRepository,
	log *zap.Logger,
) *services.ModService {
	return services.NewModService(repo, log)
}

// ========== 适配器实现 ==========

// jwtConfigAdapter 适配 config.Configuration 到 services.JwtConfig 接口
type jwtConfigAdapter struct {
	cfg *config.Configuration
}

func (a *jwtConfigAdapter) GetSecret() string {
	return a.cfg.Jwt.Secret
}

func (a *jwtConfigAdapter) GetTtl() int64 {
	return a.cfg.Jwt.JwtTtl
}

func (a *jwtConfigAdapter) GetBlacklistGracePeriod() int64 {
	return a.cfg.Jwt.JwtBlacklistGracePeriod
}

func (a *jwtConfigAdapter) GetRefreshGracePeriod() int64 {
	return a.cfg.Jwt.RefreshGracePeriod
}

// redisAdapter 适配 *redis.Client 到 services.RedisClient 接口
type redisAdapter struct {
	client *redis.Client
}

func (a *redisAdapter) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.client.SetNX(ctx, key, value, expiration).Err()
}

func (a *redisAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.client.Get(ctx, key).Result()
}

// userGetterAdapter 适配 *UserService 到 services.UserGetter 接口
type userGetterAdapter struct {
	svc *services.UserService
}

func (a *userGetterAdapter) GetUserInfo(id string) (services.JwtUser, error) {
	return a.svc.GetUserInfo(id)
}
