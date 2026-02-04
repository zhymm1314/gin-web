package cron

import (
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// JobHandler 定时任务接口
type JobHandler interface {
	Name() string // 任务名称
	Spec() string // cron 表达式
	Run()         // 执行方法
}

// Manager 定时任务管理器
type Manager struct {
	cron     *cron.Cron
	jobs     map[string]cron.EntryID
	handlers []JobHandler
	mu       sync.RWMutex
	log      *zap.Logger
}

// NewManager 创建定时任务管理器
func NewManager(log *zap.Logger) *Manager {
	return &Manager{
		cron: cron.New(cron.WithSeconds()), // 支持秒级调度
		jobs: make(map[string]cron.EntryID),
		log:  log,
	}
}

// Register 注册定时任务
func (m *Manager) Register(handler JobHandler) {
	m.handlers = append(m.handlers, handler)
}

// Start 启动所有定时任务
func (m *Manager) Start() error {
	for _, handler := range m.handlers {
		h := handler // 避免闭包问题
		entryID, err := m.cron.AddFunc(h.Spec(), func() {
			defer func() {
				if r := recover(); r != nil {
					m.log.Error("cron job panic",
						zap.String("name", h.Name()),
						zap.Any("error", r))
				}
			}()
			m.log.Debug("cron job running", zap.String("name", h.Name()))
			h.Run()
		})
		if err != nil {
			m.log.Error("add cron job failed",
				zap.String("name", h.Name()),
				zap.String("spec", h.Spec()),
				zap.Error(err))
			continue
		}
		m.mu.Lock()
		m.jobs[h.Name()] = entryID
		m.mu.Unlock()
		m.log.Info("cron job registered",
			zap.String("name", h.Name()),
			zap.String("spec", h.Spec()))
	}
	m.cron.Start()
	m.log.Info("cron manager started", zap.Int("job_count", len(m.jobs)))
	return nil
}

// Stop 停止定时任务管理器
func (m *Manager) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
	m.log.Info("cron manager stopped")
}

// Remove 移除指定任务
func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entryID, ok := m.jobs[name]; ok {
		m.cron.Remove(entryID)
		delete(m.jobs, name)
		m.log.Info("cron job removed", zap.String("name", name))
	}
}

// GetEntries 获取所有任务
func (m *Manager) GetEntries() []cron.Entry {
	return m.cron.Entries()
}
