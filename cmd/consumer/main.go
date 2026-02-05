package main

import (
	fxmodule "gin-web/internal/fx"
)

func main() {
	// 使用 fx 启动 RabbitMQ 消费者服务
	// fx 会自动管理生命周期：
	// - 配置加载
	// - 日志初始化
	// - 数据库连接
	// - Redis 连接
	// - RabbitMQ 消费者
	// - 优雅关闭
	fxmodule.NewConsumerApp().Run()
}
