package main

import (
	appCron "gin-web/app/cron"
	"gin-web/bootstrap"
	"gin-web/global"
	"gin-web/pkg/cron"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置和日志
	bootstrap.InitializeConfig()
	global.App.Log = bootstrap.InitializeLog()
	global.App.Log.Info("Cron service initializing...")

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

	// 创建定时任务管理器
	cronManager := cron.NewManager(global.App.Log)

	// 注册定时任务
	cronManager.Register(&appCron.CleanupJob{})
	cronManager.Register(&appCron.HealthCheckJob{})

	// 启动
	if err := cronManager.Start(); err != nil {
		global.App.Log.Fatal("Failed to start cron manager: " + err.Error())
	}
	global.App.Log.Info("Cron service started")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cronManager.Stop()
	global.App.Log.Info("Cron service stopped")
}
