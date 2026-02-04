# RabbitMQ 消息队列使用指南

本文档详细说明如何在 Gin-Web 项目中使用 RabbitMQ 消息队列，包括生产者和消费者的完整开发流程。

---

## 目录

- [概述](#概述)
- [架构说明](#架构说明)
- [配置说明](#配置说明)
- [开发消费者](#开发消费者)
  - [Step 1: 实现消费者处理器](#step-1-实现消费者处理器)
  - [Step 2: 注册消费者](#step-2-注册消费者)
  - [Step 3: 配置队列](#step-3-配置队列)
- [开发生产者](#开发生产者)
  - [Step 1: 创建生产者](#step-1-创建生产者)
  - [Step 2: 在业务中使用](#step-2-在业务中使用)
- [完整示例](#完整示例)
- [最佳实践](#最佳实践)
- [错误处理与重试](#错误处理与重试)
- [注意事项](#注意事项)
- [常见问题](#常见问题)

---

## 概述

本项目使用 RabbitMQ 作为消息中间件，采用 `amqp091-go` 库进行操作。消息队列主要用于：

- 异步任务处理（如日志记录、邮件发送）
- 服务解耦
- 流量削峰

## 架构说明

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    Producer     │────▶│    RabbitMQ     │────▶│    Consumer     │
│  (生产者/API)   │     │    (消息队列)    │     │   (消费者)      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**核心组件**:

| 组件 | 位置 | 说明 |
|------|------|------|
| Producer | `app/ampq/producer/` | 消息生产者 |
| Consumer | `app/ampq/consumer/` | 消息消费者 |
| ConsumerManager | `bootstrap/rabbitmq_manager.go` | 消费者管理器 |
| Config | `config.yaml` | RabbitMQ 配置 |

---

## 配置说明

### config.yaml 配置

```yaml
rabbitmq:
  consumer_enable_start: true      # 是否启动消费者
  host: 127.0.0.1                  # RabbitMQ 地址
  port: 5672                       # RabbitMQ 端口
  username: guest                  # 用户名
  password: guest                  # 密码
  vhost: /                         # 虚拟主机
  reconnect_interval: 5            # 重连间隔(秒)

# 消费者配置
consumers:
  - queue: "log_queue"             # 队列名称
    handler: "log_consumer"        # 处理器名称（需在代码中注册）
    concurrency: 2                 # 并发消费者数量
  - queue: "email_queue"
    handler: "email_consumer"
    concurrency: 1
```

---

## 开发消费者

### Step 1: 实现消费者处理器

在 `app/ampq/consumer/` 目录下创建消费者处理器。

**必须实现 `ConsumerHandler` 接口**:

```go
type ConsumerHandler interface {
    HandleMessage(msg amqp.Delivery) error
}
```

**示例 - 创建订单消费者**:

**文件位置**: `app/ampq/consumer/order_consumer.go`

```go
package consumer

import (
    "encoding/json"
    "gin-web/app/models"
    "gin-web/global"
    amqp "github.com/rabbitmq/amqp091-go"
    "go.uber.org/zap"
)

// OrderMessage 订单消息结构
type OrderMessage struct {
    OrderID   uint   `json:"order_id"`
    UserID    uint   `json:"user_id"`
    Amount    float64 `json:"amount"`
    Action    string `json:"action"` // create, update, cancel
}

// OrderConsumer 订单消费者
type OrderConsumer struct{}

// NewOrderConsumer 创建订单消费者
func NewOrderConsumer() *OrderConsumer {
    return &OrderConsumer{}
}

// HandleMessage 处理消息
func (c *OrderConsumer) HandleMessage(msg amqp.Delivery) error {
    // 1. 使用 defer + recover 防止 panic 导致消费者退出
    defer func() {
        if r := recover(); r != nil {
            global.App.Log.Error("OrderConsumer panic recovered",
                zap.Any("panic", r),
                zap.ByteString("body", msg.Body),
            )
        }
    }()

    // 2. 解析消息
    var orderMsg OrderMessage
    if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
        global.App.Log.Error("解析订单消息失败",
            zap.Error(err),
            zap.ByteString("body", msg.Body),
        )
        // 解析失败的消息不重试，直接返回 nil 表示消费成功
        return nil
    }

    // 3. 处理业务逻辑
    global.App.Log.Info("处理订单消息",
        zap.Uint("order_id", orderMsg.OrderID),
        zap.String("action", orderMsg.Action),
    )

    switch orderMsg.Action {
    case "create":
        return c.handleCreate(orderMsg)
    case "update":
        return c.handleUpdate(orderMsg)
    case "cancel":
        return c.handleCancel(orderMsg)
    default:
        global.App.Log.Warn("未知的订单操作", zap.String("action", orderMsg.Action))
    }

    return nil
}

func (c *OrderConsumer) handleCreate(msg OrderMessage) error {
    // 创建订单逻辑
    global.App.Log.Info("创建订单", zap.Uint("order_id", msg.OrderID))
    // ... 业务逻辑
    return nil
}

func (c *OrderConsumer) handleUpdate(msg OrderMessage) error {
    // 更新订单逻辑
    return nil
}

func (c *OrderConsumer) handleCancel(msg OrderMessage) error {
    // 取消订单逻辑
    return nil
}
```

### Step 2: 注册消费者

在 `main.go` 或消费者初始化文件中注册消费者处理器。

**文件位置**: `main.go`

```go
package main

import (
    "gin-web/app/ampq/consumer"
    "gin-web/bootstrap"
    "gin-web/global"
)

func main() {
    // ... 初始化配置、数据库等

    // 注册消费者处理器
    handlers := map[string]consumer.ConsumerHandler{
        "log_consumer":   &consumer.LogConsumer{},
        "order_consumer": consumer.NewOrderConsumer(),  // 新增
        "email_consumer": consumer.NewEmailConsumer(),  // 新增
    }

    // 启动消费者管理器
    if global.App.Config.RabbitMQ.ConsumerEnableStart {
        cm := bootstrap.NewConsumerManager(global.App.Config, handlers)
        go cm.Start()
    }

    // ... 启动 HTTP 服务
}
```

### Step 3: 配置队列

在 `config.yaml` 中添加消费者配置：

```yaml
consumers:
  - queue: "order_queue"
    handler: "order_consumer"
    concurrency: 3   # 并发3个消费者处理
```

---

## 开发生产者

### Step 1: 创建生产者

在 `app/ampq/producer/` 目录下创建生产者。

**文件位置**: `app/ampq/producer/order_producer.go`

```go
package producer

import (
    "encoding/json"
    "gin-web/config"
)

// OrderProducer 订单生产者
type OrderProducer struct {
    *BaseProducer
}

// NewOrderProducer 创建订单生产者
func NewOrderProducer(cfg config.RabbitMQ) (*OrderProducer, error) {
    base, err := NewBaseProducer(cfg, "order_queue")
    if err != nil {
        return nil, err
    }
    return &OrderProducer{BaseProducer: base}, nil
}

// OrderMessage 订单消息
type OrderMessage struct {
    OrderID uint    `json:"order_id"`
    UserID  uint    `json:"user_id"`
    Amount  float64 `json:"amount"`
    Action  string  `json:"action"`
}

// PublishOrder 发布订单消息
func (p *OrderProducer) PublishOrder(msg OrderMessage) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    return p.Publish(body)
}

// QueueName 返回队列名称
func (p *OrderProducer) QueueName() string {
    return "order_queue"
}
```

### Step 2: 在业务中使用

**在 Service 层使用生产者**:

```go
package services

import (
    "gin-web/app/ampq/producer"
    "gin-web/global"
)

type OrderService struct {
    orderProducer *producer.OrderProducer
}

func NewOrderService() (*OrderService, error) {
    orderProducer, err := producer.NewOrderProducer(global.App.Config.RabbitMQ)
    if err != nil {
        return nil, err
    }
    return &OrderService{orderProducer: orderProducer}, nil
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID uint, amount float64) (uint, error) {
    // 1. 创建订单到数据库
    orderID := uint(12345) // 假设这是创建后的订单ID

    // 2. 发送消息到队列进行异步处理
    msg := producer.OrderMessage{
        OrderID: orderID,
        UserID:  userID,
        Amount:  amount,
        Action:  "create",
    }

    if err := s.orderProducer.PublishOrder(msg); err != nil {
        global.App.Log.Error("发送订单消息失败")
        // 注意：消息发送失败不影响主流程，可以记录日志后继续
    }

    return orderID, nil
}
```

**在 Controller 中使用**:

```go
func CreateOrder(c *gin.Context) {
    var req request.CreateOrder
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidateFail(c, request.GetErrorMsg(req, err))
        return
    }

    userID := c.GetUint("user_id")
    orderID, err := services.OrderServiceInstance.CreateOrder(userID, req.Amount)
    if err != nil {
        response.BusinessFail(c, err.Error())
        return
    }

    response.Success(c, gin.H{"order_id": orderID})
}
```

---

## 完整示例

以下是一个完整的邮件发送队列示例：

### 1. 邮件消息结构

```go
// app/ampq/producer/email_producer.go
package producer

type EmailMessage struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
    Type    string `json:"type"` // welcome, reset_password, notification
}
```

### 2. 邮件生产者

```go
// app/ampq/producer/email_producer.go
package producer

import (
    "encoding/json"
    "gin-web/config"
)

type EmailProducer struct {
    *BaseProducer
}

func NewEmailProducer(cfg config.RabbitMQ) (*EmailProducer, error) {
    base, err := NewBaseProducer(cfg, "email_queue")
    if err != nil {
        return nil, err
    }
    return &EmailProducer{BaseProducer: base}, nil
}

func (p *EmailProducer) SendEmail(msg EmailMessage) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    return p.Publish(body)
}
```

### 3. 邮件消费者

```go
// app/ampq/consumer/email_consumer.go
package consumer

import (
    "encoding/json"
    "gin-web/app/ampq/producer"
    "gin-web/global"
    amqp "github.com/rabbitmq/amqp091-go"
    "go.uber.org/zap"
)

type EmailConsumer struct{}

func NewEmailConsumer() *EmailConsumer {
    return &EmailConsumer{}
}

func (c *EmailConsumer) HandleMessage(msg amqp.Delivery) error {
    defer func() {
        if r := recover(); r != nil {
            global.App.Log.Error("EmailConsumer panic", zap.Any("panic", r))
        }
    }()

    var emailMsg producer.EmailMessage
    if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
        global.App.Log.Error("解析邮件消息失败", zap.Error(err))
        return nil
    }

    global.App.Log.Info("发送邮件",
        zap.String("to", emailMsg.To),
        zap.String("subject", emailMsg.Subject),
        zap.String("type", emailMsg.Type),
    )

    // 调用邮件发送服务
    // err := smtp.SendEmail(emailMsg.To, emailMsg.Subject, emailMsg.Body)
    // if err != nil {
    //     return err // 返回错误会触发重试
    // }

    return nil
}
```

### 4. 配置

```yaml
# config.yaml
consumers:
  - queue: "email_queue"
    handler: "email_consumer"
    concurrency: 2
```

### 5. 注册

```go
// main.go
handlers := map[string]consumer.ConsumerHandler{
    "email_consumer": consumer.NewEmailConsumer(),
}
```

---

## 最佳实践

### 1. 消息设计

```go
// 推荐：包含足够的上下文信息
type GoodMessage struct {
    ID        string    `json:"id"`         // 唯一消息ID，用于幂等处理
    Type      string    `json:"type"`       // 消息类型
    Data      any       `json:"data"`       // 业务数据
    Timestamp int64     `json:"timestamp"`  // 消息产生时间
    RetryCount int      `json:"retry_count"` // 重试次数
}
```

### 2. 幂等性处理

```go
func (c *OrderConsumer) HandleMessage(msg amqp.Delivery) error {
    var orderMsg OrderMessage
    json.Unmarshal(msg.Body, &orderMsg)

    // 检查是否已处理过（使用 Redis 或数据库）
    key := fmt.Sprintf("processed:order:%d", orderMsg.OrderID)
    exists, _ := global.App.Redis.Exists(context.Background(), key).Result()
    if exists > 0 {
        global.App.Log.Info("消息已处理，跳过", zap.Uint("order_id", orderMsg.OrderID))
        return nil
    }

    // 处理业务逻辑
    // ...

    // 标记已处理（设置过期时间）
    global.App.Redis.Set(context.Background(), key, "1", 24*time.Hour)

    return nil
}
```

### 3. 日志记录

```go
func (c *OrderConsumer) HandleMessage(msg amqp.Delivery) error {
    startTime := time.Now()

    // 处理逻辑...

    global.App.Log.Info("消息处理完成",
        zap.String("queue", "order_queue"),
        zap.ByteString("body", msg.Body),
        zap.Duration("duration", time.Since(startTime)),
    )

    return nil
}
```

---

## 错误处理与重试

### 返回值说明

| 返回值 | 行为 |
|--------|------|
| `return nil` | 消息处理成功，确认消费 |
| `return err` | 消息处理失败，会根据配置重试 |

### 自定义重试逻辑

```go
func (c *OrderConsumer) HandleMessage(msg amqp.Delivery) error {
    var orderMsg OrderMessage
    json.Unmarshal(msg.Body, &orderMsg)

    err := c.processOrder(orderMsg)
    if err != nil {
        // 检查重试次数
        retryCount := getRetryCount(msg.Headers)
        if retryCount >= 3 {
            // 超过重试次数，发送到死信队列
            global.App.Log.Error("消息处理失败，已达最大重试次数",
                zap.Int("retry_count", retryCount),
            )
            return nil // 返回 nil 确认消息，避免无限重试
        }

        // 返回错误触发重试
        return err
    }

    return nil
}
```

---

## 注意事项

### 必须遵守

1. **必须使用 defer + recover**：防止 panic 导致消费者退出
2. **必须实现 ConsumerHandler 接口**：所有消费者处理器必须实现 `HandleMessage` 方法
3. **必须在配置中注册**：消费者需要在 `config.yaml` 的 `consumers` 中配置
4. **必须在代码中注册处理器**：处理器名称需要在 `handlers` map 中注册

### 建议遵守

1. **幂等设计**：同一消息多次消费应产生相同结果
2. **合理设置并发数**：根据业务复杂度和机器性能调整 `concurrency`
3. **记录处理日志**：便于问题排查
4. **设置消息过期时间**：避免消息积压
5. **使用结构化消息**：便于解析和扩展

### 避免

1. **避免在消费者中进行耗时操作**：如有必要，再次拆分为子任务
2. **避免忽略错误**：即使选择不重试，也应记录日志
3. **避免直接使用全局连接**：使用 ConsumerManager 管理连接

---

## 常见问题

### Q: 消费者启动后没有消费消息？

A: 检查以下配置：
1. `consumer_enable_start` 是否为 `true`
2. 处理器名称是否与配置中的 `handler` 一致
3. 队列名称是否正确

### Q: 如何优雅关闭消费者？

A: 使用 `ConsumerManager.Stop()` 方法：

```go
cm.Stop() // 会等待所有消费者处理完当前消息后关闭
```

### Q: 消息消费失败如何处理？

A: 返回 `error` 会触发重试机制，返回 `nil` 表示消费成功。建议设置最大重试次数，超过后发送到死信队列。

### Q: 如何查看队列状态？

A: 访问 RabbitMQ 管理界面 `http://localhost:15672`，默认账号密码 `guest/guest`。
