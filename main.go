package main

import (
	_ "gin-web/docs" // Swagger docs
	fxmodule "gin-web/internal/fx"
)

// @title           Gin-Web API
// @version         2.0.0
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
	// 使用 fx 依赖注入启动应用
	// fx 会自动管理所有组件的生命周期：
	// - 配置加载
	// - 数据库连接
	// - Redis 连接
	// - HTTP 服务器
	// - RabbitMQ 消费者（如果启用）
	// - 定时任务（如果启用）
	// - WebSocket（如果启用）
	fxmodule.NewApp().Run()
}
