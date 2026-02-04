package main

import (
	"context"
	"gin-web/app/controllers"
	"gin-web/app/middleware"
	"gin-web/bootstrap"
	"gin-web/global"
	"gin-web/pkg/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置和日志
	bootstrap.InitializeConfig()
	global.App.Log = bootstrap.InitializeLog()
	global.App.Log.Info("WebSocket service initializing...")

	// 初始化 Redis (WebSocket 可能需要用于分布式场景)
	global.App.Redis = bootstrap.InitializeRedis()

	// 创建 WebSocket Manager
	wsManager := websocket.NewManager(global.App.Log)

	// 设置 Gin 模式
	if global.App.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 路由
	router := gin.New()
	router.Use(gin.Logger(), middleware.CustomRecovery())
	router.Use(middleware.Cors())

	// 注册 WebSocket 控制器
	wsController := controllers.NewWebSocketController(wsManager)
	apiGroup := router.Group("/api")
	controllers.RegisterController(apiGroup, wsController)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":       "ok",
			"service":      "websocket",
			"online_count": wsManager.OnlineCount(),
		})
	})

	// 获取端口
	port := global.App.Config.WebSocket.Port
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		global.App.Log.Info("WebSocket server starting on port " + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	global.App.Log.Info("Shutting down WebSocket server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("WebSocket server shutdown error:", err)
	}

	wsManager.Close()
	global.App.Log.Info("WebSocket service stopped")
}
