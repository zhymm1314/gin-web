---
name: add-mq-consumer
description: 在 gin-web 脚手架中新增 MQ 消费者/生产者（pkg/mq 抽象，Kafka/RabbitMQ，fx 装配，consumer.yaml 配置）
---

# 新增 MQ 消费者 / 生产者

本项目消息队列统一抽象在 `pkg/mq`，业务面向此包编程，不直接依赖 sarama/amqp091。
Kafka(sarama) 为主、RabbitMQ 兼容，`config.yaml` 的 `kafka.enable`/`rabbitmq.enable` 二选一。
范例参照 `app/amqp/consumer/log_consumer.go`。

## 抽象层（`pkg/mq/mq.go`）
```go
type Message struct {
    Body      []byte
    Topic     string              // RabbitMQ=队列名；Kafka=topic
    MessageID string              // RabbitMQ: MessageId；Kafka: 留空
    Headers   map[string][]byte
    Partition int32               // 仅 Kafka
    Offset    int64               // 仅 Kafka
}

type Handler interface {
    HandleMessage(msg *Message) error  // 返回 error 触发重投
}
// 函数式：mq.HandlerFunc(func(msg *mq.Message) error{...})

type ConsumerConfig struct {
    Topic       string // RabbitMQ=队列名；Kafka=topic
    Group       string // 仅 Kafka
    Handler     string // 注册表 key，对应 ProvideConsumerHandlers 的 map key
    Concurrency int    // 并发实例数
}

type Manager interface { Start(); Stop(); RegisterHandler(name string, h Handler) }
type Producer interface { Publish(body []byte) error; Topic() string }
```
类型别名：`app/amqp/consumer/abstract.go` `ConsumerHandler = mq.Handler`；`app/amqp/producer/abstract.go` `Producer = mq.Producer`。

## 新增消费者

### 1. 实现 Handler — `app/amqp/consumer/xxx_consumer.go`
```go
package consumer

import (
	"gin-web/config"
	"gin-web/pkg/mq"
	"go.uber.org/zap"
)

type XxxConsumer struct {
	cfg *config.Configuration
	log *zap.Logger
}

func NewXxxConsumer(cfg *config.Configuration, log *zap.Logger) *XxxConsumer {
	return &XxxConsumer{cfg: cfg, log: log}
}

func (c *XxxConsumer) HandleMessage(msg *mq.Message) error {
	defer func() {
		if r := recover(); r != nil {
			c.log.Error("panic", zap.Any("r", r)) // 参考 log_consumer.go:27-31
		}
	}()
	// 业务处理...
	// 不可恢复的错误返回 nil 吞掉（避免无限重投），见 log_consumer.go:36-38
	return nil
}
```

### 2. 注册到 Manager — `internal/fx/mq.go` `ProvideConsumerHandlers`
```go
func ProvideConsumerHandlers() map[string]mq.Handler {
	return map[string]mq.Handler{
		"LogConsumer": &consumer.LogConsumer{},
		"XxxConsumer": &consumer.XxxConsumer{}, // 新增
	}
}
```
`ProvideConsumerHandlers` 接收 `cfg`/`log`（fx 自动注入），通过 `NewXxxConsumer(cfg, log)` 正确构造。新增消费者在此 map 加一行，**key 必须与 `config/yaml/consumer.yaml` 的 `handler` 字段一致**。

### 3. 配置消费者 — `config/yaml/consumer.yaml`
```yaml
consumers:
  - topic: "xxx_topic"          # Kafka=topic；RabbitMQ=队列名
    # queue: "xxx_queue"        # 向后兼容，topic 为空时用 queue
    group: "xxx-consumer-group" # 仅 Kafka；为空则回退 config.yaml 的 kafka.group_id
    concurrency: 5
    handler: "XxxConsumer"      # 必须与 ProvideConsumerHandlers 的 key 一致
```
该文件由 `loadConsumerConfigs`(`internal/fx/mq.go:66-91`) 读取，转成 `[]mq.ConsumerConfig` 喂给 Manager。Manager 启动时按 `c.Handler` 从 map 取 handler，找不到只 Warn 跳过（`pkg/kafka/manager.go:44`、`pkg/rabbitmq/manager.go:92`）。**无需改 fx wire 逻辑**，consumer.yaml 全部条目自动消费。

## 新增生产者 — `app/amqp/producer/`
生产者**未接入 fx DI**，按需手动构造。两条对称基类：

RabbitMQ（嵌入 `*BaseProducer`，`abstract.go`）：
```go
type XxxProducer struct{ *BaseProducer }

func NewXxxProducer(cfg config.RabbitMQ) (*XxxProducer, error) {
	base, err := NewBaseProducer(cfg, "xxx_queue") // abstract.go:22
	if err != nil {
		return nil, err
	}
	return &XxxProducer{base}, nil
}
// p.Publish([]byte(`{...}`))  // 默认 exchange="", routing key=队列名, Persistent
```

Kafka（嵌入 `*KafkaBaseProducer`，`kafka_producer.go`）：
```go
type XxxKafkaProducer struct{ *KafkaBaseProducer }

func NewXxxKafkaProducer(cfg config.Kafka) (*XxxKafkaProducer, error) {
	base, err := NewKafkaBaseProducer(cfg, "xxx_topic") // kafka_producer.go:15
	if err != nil {
		return nil, err
	}
	return &XxxKafkaProducer{base}, nil
}
```
**注意**：生产者侧**没有**像消费者侧 `decideMQ` 那样的自动切换工厂，调用方需自行按 `cfg.Kafka.Enable` 分支或建工厂。

## 入口与 enable 开关
- 入口：`cmd/consumer/main.go` -> `fxmodule.NewConsumerApp().Run()`（`internal/fx/modules.go:47`：只装 `InfrastructureModule` + `MQModule(cfg, true)`，force=true）
- API 主进程：`NewApp()` 用 `MQModule(cfg, false)`
- 选型优先级（`internal/fx/mq.go:43-55` `decideMQ`）：
  1. `kafka.enable==true` → Kafka（若 RabbitMQ 也开则告警）
  2. 否则 `rabbitmq.enable==true` → RabbitMQ
  3. 否则 `force==true`（消费者专用应用）→ 回退 RabbitMQ
  4. 否则不加载 MQ

## Kafka vs RabbitMQ 差异
| 维度 | Kafka | RabbitMQ |
|---|---|---|
| topic/queue | `topic` | 队列名（同字段或回退 `queue`） |
| 消费组 | `group`（空则回退 `kafka.group_id`） | 无，忽略 |
| 并发 | N 个 ConsumerGroup 加入同组，分区再平衡 | N 个 Consumer 共享连接，round-robin |
| 成功确认 | `session.MarkMessage` 自动提交 offset | `msg.Ack(false)` |
| 失败处理 | 不提交 offset + 结束会话 + 5s 退避重投 | `msg.Nack(false, true)` requeue |
| QoS/声明 | 无（topic 预存在） | `Qos(120,0,false)` + `QueueDeclare(durable=true)` |

## 要点 / 坑
1. **`config.yaml` 的 `rabbitmq.enable` / `kafka.enable` 控制主进程是否加载 MQ**（二选一，同时启用优先 Kafka）。独立进程 `cmd/consumer` 强制启用（force=true），不受开关影响。
2. **Kafka = at-least-once，有毒消息风险**：handler 返回 error 时不提交 offset、5s 后从上次提交点重投，持续失败会无限重投。不可恢复错误应返回 nil 吞掉或加死信逻辑。
3. **优雅关闭已实现**：Kafka `Stop()`(`pkg/kafka/manager.go:138-152`) 关 ConsumerGroup + `wg.Wait()`；RabbitMQ `Stop()`(`pkg/rabbitmq/manager.go:153-160`) `stopConsumers` + `conn.Close()`。均挂 fx `OnStop`。
4. **并发已实现**：`Concurrency` 在 Kafka 起 N 个 ConsumerGroup goroutine，RabbitMQ 起 N 个 Consumer goroutine。
5. **活跃路径是 `internal/fx`**。另有 `pkg/app/modules.go` + `bootstrap/rabbitmq.go` 的旧 MQ 路径（非 fx 体系，含重复 handler 注册逻辑），**新增消费者只改 `internal/fx/mq.go`，勿动 `bootstrap/rabbitmq.go`**。
