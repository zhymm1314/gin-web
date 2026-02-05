package main

import (
	fxmodule "gin-web/internal/fx"
)

func main() {
	// 使用 fx 启动 WebSocket 服务
	// fx 会自动管理生命周期：
	// - 配置加载
	// - 日志初始化
	// - Redis 连接
	// - WebSocket 管理器
	// - HTTP 服务器
	// - 优雅关闭
	fxmodule.NewWebSocketApp().Run()
}
