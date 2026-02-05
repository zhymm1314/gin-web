package fx

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"gin-web/app/controllers"
	"gin-web/app/middleware"
	"gin-web/config"
	_ "gin-web/docs" // Swagger 文档
	"gin-web/routes"
)

// RouterModule 路由模块
var RouterModule = fx.Module("router",
	fx.Provide(ProvideGinEngine),
	fx.Provide(ProvideHTTPServer),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(StartHTTPServer), // 触发 HTTP 服务器启动
)

// ControllerParams 控制器参数（分组注入）
type ControllerParams struct {
	fx.In
	Controllers []controllers.Controller `group:"controllers"`
}

// ProvideGinEngine 提供 Gin 引擎
func ProvideGinEngine(cfg *config.Configuration) *gin.Engine {
	// 禁用 Gin 的 debug 日志输出
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	// 只在非生产环境使用默认 Logger（请求日志）
	if cfg.App.Env != "production" {
		r.Use(gin.Logger())
	}
	r.Use(middleware.CustomRecovery(cfg))
	r.Use(middleware.Cors())

	// Swagger 文档 (非生产环境)
	if cfg.App.Env != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return r
}

// ProvideHTTPServer 提供 HTTP 服务器
func ProvideHTTPServer(
	lc fx.Lifecycle,
	cfg *config.Configuration,
	engine *gin.Engine,
	log *zap.Logger,
) *http.Server {
	server := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Gracefully shutting down...")
			return server.Shutdown(ctx)
		},
	})

	return server
}

// RegisterRoutes 注册所有路由
func RegisterRoutes(engine *gin.Engine, params ControllerParams) {
	apiGroup := engine.Group("/api")
	routes.SetApiGroupRoutes(apiGroup, params.Controllers...)
}

// StartHTTPServer 触发 HTTP 服务器的创建和启动
// 这个函数通过依赖 *http.Server 来确保 ProvideHTTPServer 被调用
func StartHTTPServer(_ *http.Server) {
	// HTTP 服务器的启动逻辑在 ProvideHTTPServer 的 lifecycle hook 中
}
