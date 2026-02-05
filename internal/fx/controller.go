package fx

import (
	"go.uber.org/fx"

	"gin-web/app/controllers"
	"gin-web/app/middleware"
	"gin-web/app/services"
)

// ControllerModule 控制器模块
var ControllerModule = fx.Module("controller",
	fx.Provide(
		// 使用分组注入，自动收集所有控制器
		fx.Annotate(
			NewAuthController,
			fx.ResultTags(`group:"controllers"`),
		),
		fx.Annotate(
			NewModController,
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewAuthController 创建认证控制器
func NewAuthController(
	userSvc *services.UserService,
	jwtSvc *services.JwtService,
	jwtMw *middleware.JwtMiddleware,
) controllers.Controller {
	return controllers.NewAuthController(userSvc, jwtSvc, jwtMw)
}

// NewModController 创建 Mod 控制器
func NewModController(
	modSvc *services.ModService,
) controllers.Controller {
	return controllers.NewModController(modSvc)
}
