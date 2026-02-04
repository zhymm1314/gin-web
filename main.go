package main

import (
	"gin-web/bootstrap"
	"gin-web/global"
	"gin-web/internal/container"
)

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

	// 使用 DI 启动服务器
	bootstrap.RunServerWithDI(app.GetControllers()...)
}
