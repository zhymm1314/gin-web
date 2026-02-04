package main

import (
	"gin-web/app/controllers"
	"gin-web/bootstrap"
	"gin-web/global"
	"gin-web/internal/container"
	"gin-web/pkg/app"

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
	// 1. 初始化基础设施
	bootstrap.InitializeConfig()
	global.App.Log = bootstrap.InitializeLog()
	global.App.Log.Info("log init success!")

	global.App.DB = bootstrap.InitializeDB()
	defer closeDB()

	bootstrap.InitializeValidator()
	global.App.Redis = bootstrap.InitializeRedis()

	// 2. 初始化 Wire 依赖注入
	diApp, err := container.InitializeApp()
	if err != nil {
		global.App.Log.Fatal("Failed to initialize app: " + err.Error())
	}

	// 3. 创建应用并注册模块
	application := app.NewApplication(global.App.Log)
	app.RegisterModules(application)

	// 4. 初始化并启动所有模块
	if err := application.Init(); err != nil {
		global.App.Log.Fatal("Failed to init modules: " + err.Error())
	}
	if err := application.Start(); err != nil {
		global.App.Log.Fatal("Failed to start modules: " + err.Error())
	}
	defer application.Stop()

	// 5. 组装控制器
	allControllers := diApp.GetControllers()
	if wsModule := app.GetWebSocketModule(application); wsModule != nil {
		wsController := controllers.NewWebSocketController(wsModule.Manager())
		allControllers = append(allControllers, wsController)
	}

	// 6. 启动 HTTP 服务
	bootstrap.RunServer(allControllers...)
}

func closeDB() {
	if global.App.DB != nil {
		db, _ := global.App.DB.DB()
		db.Close()
	}
}
