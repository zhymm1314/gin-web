package kafka

import (
	"context"
	"sync"
	"time"

	"gin-web/pkg/mq"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// reconnectBackoff 消费出错后的重试间隔。
const reconnectBackoff = 5 * time.Second

// Manager Kafka 消费者管理器，实现 mq.Manager。
type Manager struct {
	cfg       *Config
	consumers []mq.ConsumerConfig
	handlers  map[string]mq.Handler
	log       *zap.Logger
	done      chan struct{}
	wg        sync.WaitGroup

	mu     sync.Mutex
	groups []sarama.ConsumerGroup
}

// NewManager 创建 Kafka 管理器。
func NewManager(cfg *Config, consumers []mq.ConsumerConfig, handlers map[string]mq.Handler, log *zap.Logger) *Manager {
	return &Manager{
		cfg:       cfg,
		consumers: consumers,
		handlers:  handlers,
		log:       log,
		done:      make(chan struct{}),
	}
}

// Start 启动所有消费者（阻塞调用方，应在独立 goroutine 中运行）。
func (m *Manager) Start() {
	for _, c := range m.consumers {
		handler, ok := m.handlers[c.Handler]
		if !ok {
			m.log.Warn("handler not registered", zap.String("handler", c.Handler))
			continue
		}

		group := c.Group
		if group == "" {
			group = m.cfg.GroupID
		}
		if group == "" {
			// 兜底：按 topic 生成默认消费组
			group = "gin-web-" + c.Topic
		}

		n := c.Concurrency
		if n <= 0 {
			n = 1
		}

		// 启动 n 个消费者实例加入同一消费组，分区在组内再平衡，
		// 与 RabbitMQ 多消费者 round-robin 的并发模型对应。
		for i := 0; i < n; i++ {
			m.wg.Add(1)
			go m.runConsumer(c.Topic, group, handler, i)
		}
	}
}

// runConsumer 运行单个消费者实例的消费循环。
func (m *Manager) runConsumer(topic, group string, handler mq.Handler, idx int) {
	defer m.wg.Done()

	saramaCfg, err := buildSaramaConfig(m.cfg)
	if err != nil {
		m.log.Error("build sarama config failed", zap.Error(err))
		return
	}

	cg, err := sarama.NewConsumerGroup(m.cfg.Brokers, group, saramaCfg)
	if err != nil {
		m.log.Error("create kafka consumer group failed",
			zap.String("topic", topic),
			zap.String("group", group),
			zap.Error(err))
		return
	}
	m.mu.Lock()
	m.groups = append(m.groups, cg)
	m.mu.Unlock()

	// done 关闭时取消消费上下文
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-m.done
		cancel()
	}()

	h := &consumerGroupHandler{
		handler: handler,
		topic:   topic,
		log:     m.log,
		done:    m.done,
	}

	m.log.Info("kafka consumer started",
		zap.String("topic", topic),
		zap.String("group", group),
		zap.Int("instance", idx+1))

	for {
		// Consume 阻塞直到会话结束（再平衡 / 出错 / 上下文取消）
		if err := cg.Consume(ctx, []string{topic}, h); err != nil {
			m.log.Error("kafka consume error",
				zap.String("topic", topic),
				zap.String("group", group),
				zap.Error(err))
		}

		// 上下文已取消（正在停止）
		if ctx.Err() != nil {
			return
		}

		// 出错后短暂退避，避免紧密循环
		select {
		case <-m.done:
			return
		case <-time.After(reconnectBackoff):
		}
	}
}

// Stop 停止管理器并关闭所有消费者组。
func (m *Manager) Stop() {
	close(m.done)

	m.mu.Lock()
	for _, cg := range m.groups {
		if err := cg.Close(); err != nil {
			m.log.Warn("close kafka consumer group failed", zap.Error(err))
		}
	}
	m.groups = nil
	m.mu.Unlock()

	m.wg.Wait()
	m.log.Info("kafka manager stopped")
}

// RegisterHandler 注册消费者处理器。
func (m *Manager) RegisterHandler(name string, handler mq.Handler) {
	if m.handlers == nil {
		m.handlers = make(map[string]mq.Handler)
	}
	m.handlers[name] = handler
}
