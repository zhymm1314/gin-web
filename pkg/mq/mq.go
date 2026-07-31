// pkg/mq 提供与底层消息队列实现无关的统一抽象。
// 当前由 RabbitMQ 与 Kafka 两种实现支撑，业务代码面向本包接口编程，
// 通过配置选择具体实现，无需改动 handler。
package mq

// Message 消息队列消息的统一抽象，屏蔽底层 MQ（RabbitMQ / Kafka）的差异。
type Message struct {
	Body      []byte              // 消息体
	Topic     string              // 主题 / 队列名
	MessageID string              // 消息ID（RabbitMQ: MessageId；Kafka: 无原生ID时留空）
	Headers   map[string][]byte   // 消息头

	// 以下为 Kafka 语义字段，RabbitMQ 实现不填充
	Partition int32
	Offset    int64
}

// Handler 消息处理器接口。
// 返回 error 表示处理失败，由具体 MQ 实现决定重投策略
// （RabbitMQ: Nack(requeue)；Kafka: 不提交 offset，触发重投）。
type Handler interface {
	HandleMessage(msg *Message) error
}

// HandlerFunc 函数式 Handler，便于将普通函数适配为 Handler。
type HandlerFunc func(msg *Message) error

// HandleMessage 实现 Handler 接口。
func (f HandlerFunc) HandleMessage(msg *Message) error { return f(msg) }

// ConsumerConfig 消费者配置（与具体 MQ 实现无关）。
type ConsumerConfig struct {
	Topic       string // RabbitMQ: 队列名；Kafka: topic
	Group       string // Kafka 消费组（RabbitMQ 忽略）
	Handler     string // 处理器名称，对应注册表中的 key
	Concurrency int    // 并发消费者数量
}

// Manager 消息队列消费者管理器接口。
type Manager interface {
	// Start 启动消费者管理器（通常在独立 goroutine 中运行）。
	Start()
	// Stop 停止管理器并释放资源。
	Stop()
	// RegisterHandler 注册消费者处理器。
	RegisterHandler(name string, handler Handler)
}

// Producer 消息生产者接口。
type Producer interface {
	// Publish 发布一条消息。
	Publish(body []byte) error
	// Topic 返回目标主题 / 队列名。
	Topic() string
}
