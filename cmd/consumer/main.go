package main

import (
	"gin-web/bootstrap"
	"gin-web/global"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置和日志
	bootstrap.InitializeConfig()
	global.App.Log = bootstrap.InitializeLog()
	global.App.Log.Info("Consumer service initializing...")

	// 初始化数据库
	global.App.DB = bootstrap.InitializeDB()
	defer func() {
		if global.App.DB != nil {
			db, _ := global.App.DB.DB()
			db.Close()
		}
	}()

	// 初始化 Redis
	global.App.Redis = bootstrap.InitializeRedis()

	// 启动 RabbitMQ 消费者
	cm := bootstrap.InitRabbitmq()
	if cm == nil {
		global.App.Log.Fatal("Failed to start RabbitMQ consumer manager")
	}
	global.App.Log.Info("RabbitMQ consumer service started")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cm.Stop()
	global.App.Log.Info("RabbitMQ consumer service stopped")
}
