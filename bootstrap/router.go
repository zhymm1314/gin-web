package bootstrap

import (
	"context"
	"gin-web/app/controllers"
	"gin-web/app/middleware"
	"gin-web/global"
	"gin-web/routes"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setupRouter() *gin.Engine {

	if global.App.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger(), middleware.CustomRecovery())
	router.Use(middleware.Cors())

	// 注册 api 分组路由
	apiGroup := router.Group("/api")
	routes.SetApiGroupRoutes(apiGroup)

	return router
}

// SetupRouterWithDI 使用依赖注入设置路由
func SetupRouterWithDI(ctrls ...controllers.Controller) *gin.Engine {
	if global.App.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger(), middleware.CustomRecovery())
	router.Use(middleware.Cors())

	// 注册 api 分组路由 (使用依赖注入)
	apiGroup := router.Group("/api")
	routes.SetApiGroupRoutesWithDI(apiGroup, ctrls...)

	return router
}

// RunServer 启动服务器 (Legacy 版本，保留兼容)
func RunServer() {
	runWithRouter(setupRouter())
}

// RunServerWithDI 使用依赖注入启动服务器
func RunServerWithDI(ctrls ...controllers.Controller) {
	runWithRouter(SetupRouterWithDI(ctrls...))
}

// runWithRouter 通用服务器启动逻辑
func runWithRouter(r *gin.Engine) {
	srv := &http.Server{
		Addr:    ":" + global.App.Config.App.Port,
		Handler: r,
	}

	// 在 goroutine 中启动服务器
	go func() {
		global.App.Log.Info("Server starting on port " + global.App.Config.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
