package main

import (
	"gin-web/bootstrap"
	"gin-web/global"
)

func main() {

	//初始化配置
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
	bootstrap.InitializeValidator()
	// 初始化Redis
	global.App.Redis = bootstrap.InitializeRedis()
	go bootstrap.InitRabbitmq()
	bootstrap.RunServer()

}
