# 消息队列使用指南（Kafka / RabbitMQ）

本文档说明 Gin-Web 的消息队列用法。消息队列已抽象到 `pkg/mq`，业务代码面向统一接口编程，通过配置在 **Kafka（主要维护）** 与 **RabbitMQ** 之间切换，handler 无需改动。

---

## 目录

- [概述](#概述)
- [架构说明](#架构说明)
- [配置说明](#配置说明)
- [选择哪一种 MQ](#选择哪一种-mq)
- [本地启动 Kafka（Docker）](#本地启动-kafkadocker)
- [开发消费者](#开发消费者)
- [开发生产者](#开发生产者)
- [从 RabbitMQ 切换到 Kafka](#从-rabbitmq-切换到-kafka)
- [可靠性与错误处理](#可靠性与错误处理)
- [本地测试](#本地测试)
- [常见问题](#常见问题)

---

## 概述

- 统一抽象：`pkg/mq` 定义 `Message` / `Handler` / `ConsumerConfig` / `Manager` / `Producer`。
- 两种实现：
  - `pkg/kafka`：基于 [`github.com/IBM/sarama`](https://github.com/IBM/sarama)，**后期主要维护**。
  - `pkg/rabbitmq`：基于 `amqp091-go`，保留兼容。
- 业务 handler 接收 `*mq.Message`（不再耦合 `amqp.Delivery`），同一份 handler 代码可在两种 MQ 上运行。

典型用途：异步任务处理（日志记录、邮件发送）、服务解耦、流量削峰。

## 架构说明

```
┌──────────┐      ┌──────────────────────┐      ┌──────────┐
│ Producer │─────▶│  pkg/mq (抽象接口)   │─────▶│ Consumer │
│  生产者   │      │  ┌─ pkg/kafka       │      │  消费者   │
└──────────┘      │  └─ pkg/rabbitmq    │      └──────────┘
                  └──────────────────────┘
```

| 组件 | 位置 | 说明 |
|------|------|------|
| 抽象接口 | `pkg/mq/mq.go` | `Message`/`Handler`/`Manager`/`Producer` |
| Kafka 实现 | `pkg/kafka/` | Manager / Consumer / Producer（sarama） |
| RabbitMQ 实现 | `pkg/rabbitmq/` | Manager / Consumer |
| 消费者 handler | `app/amqp/consumer/` | 实现 `mq.Handler`，如 `LogConsumer` |
| 生产者 | `app/amqp/producer/` | `BaseProducer`(RabbitMQ) / `KafkaBaseProducer` |
| fx 装配 | `internal/fx/mq.go` | `MQModule` 按配置选择实现 |
| 消费者配置 | `config/yaml/consumer.yaml` | topic / group / handler / concurrency |

## 配置说明

### config.yaml

```yaml
# Kafka（主要维护）
kafka:
  enable: true                       # 启用 Kafka（与 rabbitmq 二选一，同时启用优先 Kafka）
  brokers:                           # broker 地址列表
    - 127.0.0.1:9092
  version:                           # 如 "3.7.1"；留空自动协商
  username:                          # SASL 用户名（无认证留空）
  password:                          # SASL 密码（无认证留空）
  group_id: gin-web-consumer         # 默认消费组（可在 consumer.yaml 按 consumer 覆盖）
  initial_offset: newest             # 新消费组起始 offset：newest(默认) / oldest

# RabbitMQ（兼容保留）
rabbitmq:
  enable: false                      # 启用 RabbitMQ
  host: 127.0.0.1
  port: 5672
  username: guest
  password: guest
  vhost: /
```

### consumer.yaml

```yaml
consumers:
  - topic: "order_topic"             # Kafka topic；RabbitMQ 下为队列名
    # queue: "order_queue"           # RabbitMQ 队列名（向后兼容：topic 为空时使用）
    group: "order-consumer-group"    # Kafka 消费组（为空则用 config 的 kafka.group_id；RabbitMQ 忽略）
    concurrency: 3                   # 并发消费者数量
    handler: "OrderConsumer"         # 处理器名称（需在代码中注册）
```

## 选择哪一种 MQ

通过 `enable` 开关选择（与 Cron/WebSocket 模式一致）：

| 配置 | 生效的 MQ |
|------|-----------|
| `kafka.enable: true` | Kafka（同时启用时优先 Kafka 并告警） |
| 仅 `rabbitmq.enable: true` | RabbitMQ |
| 两者都未启用 | 主进程不加载 MQ；消费者专用进程（`cmd/consumer`）回退 RabbitMQ |

消费者专用进程 `cmd/consumer` 强制启用 MQ：按配置选择，两者都没开时回退 RabbitMQ（历史默认）。如需让消费者进程跑 Kafka，把 `kafka.enable` 设为 `true` 即可。

## 本地启动 Kafka（Docker）

项目根目录已提供 `docker-compose.yml`（Kafka 采用 KRaft 模式，无需 Zookeeper）：

```bash
# 仅启动 Kafka（其余基础设施已独立运行时推荐）
docker compose up -d kafka

# 或一次性启动全部基础设施（mysql/redis/postgres/kafka）
docker compose up -d

# 查看日志
docker logs -f gin-web-kafka

# 停止
docker compose down          # 保留数据卷
docker compose down -v       # 清除数据卷
```

主机连接地址：`127.0.0.1:9092`（已在 `KAFKA_ADVERTISED_LISTENERS` 中通告）。主题默认自动创建、3 分区。

> 若 `docker pull` 拉取 `apache/kafka` 镜像失败（网络问题），可改用镜像加速，例如：
> `docker pull docker.m.daocloud.io/apache/kafka:3.7.1 && docker tag docker.m.daocloud.io/apache/kafka:3.7.1 apache/kafka:3.7.1`

## 开发消费者

### Step 1：实现 `mq.Handler`

**文件位置**：`app/amqp/consumer/order_consumer.go`

```go
package consumer

import (
	"encoding/json"

	"gin-web/config"
	"gin-web/pkg/mq"

	"go.uber.org/zap"
)

type OrderMessage struct {
	OrderID uint   `json:"order_id"`
	Action  string `json:"action"`
}

type OrderConsumer struct {
	cfg *config.Configuration
	log *zap.Logger
}

func NewOrderConsumer(cfg *config.Configuration, log *zap.Logger) *OrderConsumer {
	return &OrderConsumer{cfg: cfg, log: log}
}

// HandleMessage 处理消息。返回 error 会触发重投（at-least-once）。
func (c *OrderConsumer) HandleMessage(msg *mq.Message) error {
	// 1. defer + recover 防止 panic 导致消费中断
	defer func() {
		if r := recover(); r != nil {
			c.log.Error("OrderConsumer panic recovered", zap.Any("panic", r))
		}
	}()

	// 2. 解析消息（msg.Body 为原始字节，与底层 MQ 无关）
	var order OrderMessage
	if err := json.Unmarshal(msg.Body, &order); err != nil {
		c.log.Error("解析订单消息失败", zap.Error(err))
		return nil // 格式错误不重试
	}

	// 3. 业务逻辑
	c.log.Info("处理订单", zap.Uint("order_id", order.OrderID), zap.String("action", order.Action))
	return nil
}
```

### Step 2：注册处理器

**文件位置**：`internal/fx/mq.go` 的 `ProvideConsumerHandlers`

```go
func ProvideConsumerHandlers() map[string]mq.Handler {
	return map[string]mq.Handler{
		"LogConsumer":    &consumer.LogConsumer{},
		"OrderConsumer":  &consumer.LogConsumer{}, // 替换为你的 NewOrderConsumer(...)
	}
}
```

### Step 3：在 `consumer.yaml` 配置

```yaml
consumers:
  - topic: "order_topic"
    group: "order-consumer-group"
    concurrency: 3
    handler: "OrderConsumer"
```

## 开发生产者

### 方式一：直接使用 `pkg/kafka`（推荐）

```go
import "gin-web/pkg/kafka"

kcfg := &kafka.Config{
	Brokers: []string{"127.0.0.1:9092"},
}
p, err := kafka.NewProducer(kcfg, "order_topic")
if err != nil {
	return err
}
defer p.Close()

if err := p.Publish([]byte(`{"order_id":1,"action":"create"}`)); err != nil {
	// 发送失败记录日志，通常不影响主流程
}
```

### 方式二：使用 `app/amqp/producer` 的基类（与 RabbitMQ 对称）

```go
import "gin-web/app/amqp/producer"

// Kafka
bp, err := producer.NewKafkaBaseProducer(cfg.Kafka, "order_topic")
// RabbitMQ
// bp, err := producer.NewBaseProducer(cfg.RabbitMQ, "order_queue")

bp.Publish([]byte(`{"order_id":1,"action":"create"}`))
```

两者都实现 `mq.Producer`（`Publish([]byte) error` / `Topic() string`），可按配置切换。

## 从 RabbitMQ 切换到 Kafka

1. 在 `config.yaml` 把 `kafka.enable` 改为 `true`、`rabbitmq.enable` 改为 `false`。
2. 在 `consumer.yaml` 用 `topic`/`group` 描述消费者（`queue` 字段可保留，`topic` 为空时回退到 `queue`）。
3. handler 代码无需改动（面向 `*mq.Message`）。
4. 启动消费者进程：`go run cmd/consumer/main.go`。

> Kafka 的并发受分区数约束：同一消费组内，活跃消费者数不超过分区数。`concurrency` 超过分区数的部分会闲置（与 RabbitMQ 多消费者 round-robin 语义对应）。

## 可靠性与错误处理

| handler 返回 | Kafka 行为 | RabbitMQ 行为 |
|--------------|-----------|---------------|
| `nil` | 标记 offset，视为消费成功 | `Ack` |
| `error` | 不提交 offset，结束当前会话并退避后重投（at-least-once） | `Nack(requeue)` 重投 |

- 两种实现均为 **at-least-once**：消息至少被处理一次，业务应做幂等设计。
- Kafka 消费出错会退避（默认 5s）后重投，避免毒消息导致紧密循环。
- 建议消息体包含唯一 ID，用 Redis/DB 做幂等去重。

## 本地测试

`pkg/kafka/kafka_test.go` 含集成测试 `TestKafkaProduceAndConsume`，验证投递与消费端到端链路：

```bash
docker compose up -d kafka
go test ./pkg/kafka/ -run TestKafkaProduceAndConsume -v
```

Kafka 不可达时该测试自动 `Skip`，不影响 `go test ./...` 在无 Kafka 环境下的运行。

## 常见问题

**Q: 消费者启动后没有消费消息？**
检查：`kafka.enable`（或 `rabbitmq.enable`）是否为 `true`；`consumer.yaml` 中 `handler` 名称是否与 `ProvideConsumerHandlers` 注册的 key 一致；`topic`/`brokers` 是否正确。

**Q: 如何优雅关闭？**
`mq.Manager.Stop()` 会等待消费循环退出并关闭连接/消费者组，fx 生命周期已自动在 `OnStop` 调用。

**Q: 消费失败会丢消息吗？**
不会。返回 `error` 会触发重投（at-least-once）；但毒消息可能反复重投，建议设置最大重试次数 + 死信处理。

**Q: 同一消息被消费多次？**
这是 at-least-once 的正常表现。请通过消息唯一 ID 做幂等。
