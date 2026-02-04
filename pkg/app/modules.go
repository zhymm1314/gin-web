package app

import (
	appCron "gin-web/app/cron"
	"gin-web/bootstrap"
	"gin-web/global"
	"gin-web/pkg/cron"
	"gin-web/pkg/rabbitmq"
	"gin-web/pkg/websocket"

	"go.uber.org/zap"
)

// ==================== Cron 模块 ====================

// CronModule 定时任务模块
type CronModule struct {
	manager *cron.Manager
	log     *zap.Logger
}

// NewCronModule 创建定时任务模块
func NewCronModule(log *zap.Logger) *CronModule {
	return &CronModule{log: log}
}

func (m *CronModule) Name() string { return "cron" }

func (m *CronModule) Init() error {
	m.manager = cron.NewManager(m.log)

	// 注册定时任务
	m.manager.Register(&appCron.CleanupJob{})
	m.manager.Register(&appCron.HealthCheckJob{})

	return nil
}

func (m *CronModule) Start() error {
	return m.manager.Start()
}

func (m *CronModule) Stop() error {
	m.manager.Stop()
	return nil
}

// Manager 获取 Cron Manager
func (m *CronModule) Manager() *cron.Manager {
	return m.manager
}

// ==================== RabbitMQ 模块 ====================

// RabbitMQModule RabbitMQ 消费者模块
type RabbitMQModule struct {
	manager *rabbitmq.Manager
	log     *zap.Logger
}

// NewRabbitMQModule 创建 RabbitMQ 模块
func NewRabbitMQModule(log *zap.Logger) *RabbitMQModule {
	return &RabbitMQModule{log: log}
}

func (m *RabbitMQModule) Name() string { return "rabbitmq" }

func (m *RabbitMQModule) Init() error {
	return nil
}

func (m *RabbitMQModule) Start() error {
	m.manager = bootstrap.InitRabbitmq()
	if m.manager == nil {
		m.log.Warn("RabbitMQ consumer manager not started (config missing or error)")
	}
	return nil
}

func (m *RabbitMQModule) Stop() error {
	if m.manager != nil {
		m.manager.Stop()
	}
	return nil
}

// Manager 获取 RabbitMQ Manager
func (m *RabbitMQModule) Manager() *rabbitmq.Manager {
	return m.manager
}

// ==================== WebSocket 模块 ====================

// WebSocketModule WebSocket 模块
type WebSocketModule struct {
	manager *websocket.Manager
	log     *zap.Logger
}

// NewWebSocketModule 创建 WebSocket 模块
func NewWebSocketModule(log *zap.Logger) *WebSocketModule {
	return &WebSocketModule{log: log}
}

func (m *WebSocketModule) Name() string { return "websocket" }

func (m *WebSocketModule) Init() error {
	m.manager = websocket.NewManager(m.log)
	return nil
}

func (m *WebSocketModule) Start() error {
	return nil // WebSocket 是被动的，由 HTTP 请求触发
}

func (m *WebSocketModule) Stop() error {
	if m.manager != nil {
		m.manager.Close()
	}
	return nil
}

// Manager 获取 WebSocket Manager
func (m *WebSocketModule) Manager() *websocket.Manager {
	return m.manager
}

// ==================== HTTP 服务模块 ====================

// HTTPModule HTTP 服务模块
type HTTPModule struct {
	controllers []interface{}
	log         *zap.Logger
}

// NewHTTPModule 创建 HTTP 模块
func NewHTTPModule(controllers []interface{}, log *zap.Logger) *HTTPModule {
	return &HTTPModule{
		controllers: controllers,
		log:         log,
	}
}

func (m *HTTPModule) Name() string { return "http" }

func (m *HTTPModule) Init() error {
	return nil
}

func (m *HTTPModule) Start() error {
	// HTTP 服务在 RunServer 中阻塞启动，这里不处理
	return nil
}

func (m *HTTPModule) Stop() error {
	return nil
}

// ==================== 模块工厂函数 ====================

// RegisterModules 根据配置注册所有模块
func RegisterModules(application *Application) *Application {
	cfg := global.App.Config
	log := global.App.Log

	// 按条件注册模块
	application.
		RegisterIf(cfg.RabbitMQ.Enable, NewRabbitMQModule(log)).
		RegisterIf(cfg.Cron.Enable, NewCronModule(log)).
		RegisterIf(cfg.WebSocket.Enable, NewWebSocketModule(log))

	return application
}

// GetWebSocketModule 从应用中获取 WebSocket 模块
func GetWebSocketModule(application *Application) *WebSocketModule {
	for _, m := range application.Modules() {
		if ws, ok := m.(*WebSocketModule); ok {
			return ws
		}
	}
	return nil
}
