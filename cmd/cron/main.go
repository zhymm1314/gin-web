package main

import (
	fxmodule "gin-web/internal/fx"
)

func main() {
	// 使用 fx 启动定时任务服务
	// fx 会自动管理生命周期：
	// - 配置加载
	// - 日志初始化
	// - 数据库连接
	// - Redis 连接
	// - 定时任务管理器
	// - 优雅关闭
	fxmodule.NewCronApp().Run()
}
