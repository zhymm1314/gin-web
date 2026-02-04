//go:build wireinject
// +build wireinject

package container

import (
	"gin-web/app/controllers"
	"github.com/google/wire"
)

// InfrastructureSet 基础设施 Provider 集合
var InfrastructureSet = wire.NewSet(
	ProvideDB,
	ProvideRedis,
	ProvideLog,
	ProvideConfig,
)

// RepositorySet Repository Provider 集合
var RepositorySet = wire.NewSet(
	ProvideUserRepository,
	ProvideModRepository,
)

// AdapterSet 适配器 Provider 集合
var AdapterSet = wire.NewSet(
	ProvideJwtConfig,
	ProvideRedisClient,
	ProvideUserGetter,
)

// ServiceSet Service Provider 集合
var ServiceSet = wire.NewSet(
	ProvideUserService,
	ProvideJwtService,
	ProvideModService,
)

// MiddlewareSet Middleware Provider 集合
var MiddlewareSet = wire.NewSet(
	ProvideJwtMiddleware,
)

// ControllerSet Controller Provider 集合
var ControllerSet = wire.NewSet(
	ProvideAuthController,
	ProvideModController,
)

// App 应用程序依赖集合
type App struct {
	AuthController *controllers.AuthController
	ModController  *controllers.ModController
}

// ProvideApp 提供应用程序
func ProvideApp(auth *controllers.AuthController, mod *controllers.ModController) *App {
	return &App{
		AuthController: auth,
		ModController:  mod,
	}
}

// GetControllers 获取所有控制器
func (a *App) GetControllers() []controllers.Controller {
	return []controllers.Controller{
		a.AuthController,
		a.ModController,
	}
}

// InitializeApp 初始化应用程序 - Wire 将生成此函数的实现
func InitializeApp() (*App, error) {
	wire.Build(
		InfrastructureSet,
		RepositorySet,
		AdapterSet,
		ServiceSet,
		MiddlewareSet,
		ControllerSet,
		ProvideApp,
	)
	return nil, nil
}
