package rabbitmq

import (
	"fmt"
	"sync"
	"time"

	"gin-web/app/amqp/consumer"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Config RabbitMQ 配置
type Config struct {
	Host              string
	Port              int
	Username          string
	Password          string
	Vhost             string
	ReconnectInterval int
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Queue       string
	Handler     string
	Concurrency int
}

// Manager RabbitMQ 消费者管理器
type Manager struct {
	cfg            *Config
	consumers      []ConsumerConfig
	handlers       map[string]consumer.ConsumerHandler
	conn           *amqp.Connection
	activeConsumer []*Consumer
	wg             sync.WaitGroup
	done           chan struct{}
	log            *zap.Logger
}

// NewManager 创建 RabbitMQ 管理器
func NewManager(cfg *Config, consumers []ConsumerConfig, handlers map[string]consumer.ConsumerHandler, log *zap.Logger) *Manager {
	return &Manager{
		cfg:       cfg,
		consumers: consumers,
		handlers:  handlers,
		done:      make(chan struct{}),
		log:       log,
	}
}

// Start 启动消费者管理器
func (m *Manager) Start() {
	for {
		select {
		case <-m.done:
			return
		default:
			if err := m.connect(); err != nil {
				m.log.Error("RabbitMQ connection failed", zap.Error(err))
				m.reconnect()
				continue
			}

			m.startConsumers()
			m.monitorConnection()
			return
		}
	}
}

// connect 连接 RabbitMQ
func (m *Manager) connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		m.cfg.Username,
		m.cfg.Password,
		m.cfg.Host,
		m.cfg.Port,
		m.cfg.Vhost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	m.conn = conn
	m.log.Info("RabbitMQ connected",
		zap.String("host", m.cfg.Host),
		zap.Int("port", m.cfg.Port))
	return nil
}

// startConsumers 启动所有消费者
func (m *Manager) startConsumers() {
	for _, cfg := range m.consumers {
		handler, ok := m.handlers[cfg.Handler]
		if !ok {
			m.log.Warn("handler not registered", zap.String("handler", cfg.Handler))
			continue
		}

		for i := 0; i < cfg.Concurrency; i++ {
			c := NewConsumer(m.conn, cfg.Queue, handler, m.log)
			m.activeConsumer = append(m.activeConsumer, c)
			m.wg.Add(1)
			go func(consumer *Consumer) {
				defer m.wg.Done()
				consumer.Start()
			}(c)

			m.log.Info("consumer started",
				zap.String("queue", cfg.Queue),
				zap.String("handler", cfg.Handler),
				zap.Int("instance", i+1))
		}
	}
}

// monitorConnection 监控连接状态
func (m *Manager) monitorConnection() {
	closeChan := make(chan *amqp.Error)
	m.conn.NotifyClose(closeChan)

	select {
	case err := <-closeChan:
		m.log.Error("RabbitMQ connection closed", zap.Error(err))
		m.reconnect()
	case <-m.done:
		return
	}
}

// reconnect 重新连接
func (m *Manager) reconnect() {
	m.stopConsumers()

	interval := time.Duration(m.cfg.ReconnectInterval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	m.log.Info("reconnecting to RabbitMQ", zap.Duration("interval", interval))
	time.Sleep(interval)
	m.Start()
}

// stopConsumers 停止所有消费者
func (m *Manager) stopConsumers() {
	for _, c := range m.activeConsumer {
		c.Stop()
	}
	m.wg.Wait()
	m.activeConsumer = nil
}

// Stop 停止管理器
func (m *Manager) Stop() {
	close(m.done)
	m.stopConsumers()
	if m.conn != nil {
		m.conn.Close()
	}
	m.log.Info("RabbitMQ manager stopped")
}

// RegisterHandler 注册消费者处理器
func (m *Manager) RegisterHandler(name string, handler consumer.ConsumerHandler) {
	if m.handlers == nil {
		m.handlers = make(map[string]consumer.ConsumerHandler)
	}
	m.handlers[name] = handler
}
