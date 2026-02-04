package app

import (
	"sync"

	"go.uber.org/zap"
)

// Module 模块接口，所有功能模块都应实现此接口
type Module interface {
	// Name 返回模块名称
	Name() string
	// Init 初始化模块
	Init() error
	// Start 启动模块
	Start() error
	// Stop 停止模块
	Stop() error
}

// Application 应用程序管理器
type Application struct {
	modules []Module
	log     *zap.Logger
	mu      sync.Mutex
}

// NewApplication 创建应用程序管理器
func NewApplication(log *zap.Logger) *Application {
	return &Application{
		modules: make([]Module, 0),
		log:     log,
	}
}

// Register 注册模块
func (a *Application) Register(m Module) *Application {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.modules = append(a.modules, m)
	return a
}

// RegisterIf 条件注册模块
func (a *Application) RegisterIf(condition bool, m Module) *Application {
	if condition {
		return a.Register(m)
	}
	return a
}

// Init 初始化所有模块
func (a *Application) Init() error {
	for _, m := range a.modules {
		if err := m.Init(); err != nil {
			a.log.Error("module init failed",
				zap.String("module", m.Name()),
				zap.Error(err))
			return err
		}
		a.log.Info("module initialized", zap.String("module", m.Name()))
	}
	return nil
}

// Start 启动所有模块
func (a *Application) Start() error {
	for _, m := range a.modules {
		if err := m.Start(); err != nil {
			a.log.Error("module start failed",
				zap.String("module", m.Name()),
				zap.Error(err))
			return err
		}
		a.log.Info("module started", zap.String("module", m.Name()))
	}
	return nil
}

// Stop 停止所有模块 (逆序停止)
func (a *Application) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 逆序停止，先启动的后停止
	for i := len(a.modules) - 1; i >= 0; i-- {
		m := a.modules[i]
		if err := m.Stop(); err != nil {
			a.log.Error("module stop failed",
				zap.String("module", m.Name()),
				zap.Error(err))
		} else {
			a.log.Info("module stopped", zap.String("module", m.Name()))
		}
	}
}

// Modules 获取所有已注册的模块
func (a *Application) Modules() []Module {
	return a.modules
}
