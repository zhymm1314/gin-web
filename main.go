package main

import (
	"gin-web/app/controllers"
	appCron "gin-web/ap
	"gin-web/bootstrap"
	"gin-web/global"
	"gin-web/internal/container"
	"gin-web/pkg/cron"
	"gin-web/pkg/websocket"

	_ "gin-web/docs" // Swagger docs
)

// @title           Gin-Web API
// @version         1.6.0
// @description     Gin-Web 脚手架 API 文档 - 基于 Gin 框架的企业级 Go 语言后端 API 脚手架
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入 Bearer {token}

func main() {
	// 初始化配置
	bootstrap.InitializeConfig()

	// 初始化日志
	global.App.Log = bootstrap.InitializeLog()
	global.App.Log.Info("log init success!")

	// 初始化数据库
	global.App.DB = bootstrap.InitializeDB()

	// 程序关闭前，释放数据库连接
	defer func() {
		if global.App.DB != nil {
			db, _ := global.App.DB.DB()
			db.Close()
		}
	}()

	// 初始化验证器
	bootstrap.InitializeValidator()

	// 初始化 Redis
	global.App.Redis = bootstrap.InitializeRedis()

	// 使用 Wire 初始化应用
	app, err := container.InitializeApp()
	if err != nil {
		global.App.Log.Fatal("Failed to initialize app: " + err.Error())
	}

	// 启动 RabbitMQ 消费者 (根据配置)
	var consumerManager *bootstrap.ConsumerManager
	if global.App.Config.RabbitMQ.Enable {
		consumerManager = bootstrap.InitRabbitmq()
		if consumerManager != nil {
			global.App.Log.Info("RabbitMQ consumer manager started")
		}
	}

	// 启动定时任务 (根据配置)
	var cronManager *cron.Manager
	if global.App.Config.Cron.Enable {
		cronManager = cron.NewManager(global.App.Log)
		cronManager.Register(&appCron.CleanupJob{})
		cronManager.Register(&appCron.HealthCheckJob{})
		cronManager.Start()
		global.App.Log.Info("Cron manager started")
	}

	// 启动 WebSocket (根据配置)
	var wsManager *websocket.Manager
	var wsController *controllers.WebSocketController
	if global.App.Config.WebSocket.Enable {
		wsManager = websocket.NewManager(global.App.Log)
		wsController = controllers.NewWebSocketController(wsManager)
		global.App.Log.Info("WebSocket manager started")
	}

	// 组装控制器列表
	allControllers := app.GetControllers()
	if wsController != nil {
		allControllers = append(allControllers, wsController)
	}

	// 启动服务器
	bootstrap.RunServer(allControllers...)

	// 清理资源
	if consumerManager != nil {
		consumerManager.Stop()
	}
	if cronManager != nil {
		cronManager.Stop()
	}
	if wsManager != nil {
		wsManager.Close()
	}
}
