package fx

import (
	"go.uber.org/fx"

	"gin-web/app/middleware"
	"gin-web/app/services"
)

// MiddlewareModule 中间件模块
var MiddlewareModule = fx.Module("middleware",
	fx.Provide(
		ProvideJwtMiddleware,
	),
)

// ProvideJwtMiddleware 提供 JWT 中间件
func ProvideJwtMiddleware(jwtSvc *services.JwtService) *middleware.JwtMiddleware {
	return middleware.NewJwtMiddleware(jwtSvc)
}
