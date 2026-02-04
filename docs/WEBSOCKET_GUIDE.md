# WebSocket 指南

本指南介绍如何使用框架的 WebSocket 功能，基于 Melody 库实现。

## 目录

- [快速开始](#快速开始)
- [启动方式](#启动方式)
- [API 接口](#api-接口)
- [前端连接](#前端连接)
- [服务端推送](#服务端推送)
- [消息格式](#消息格式)
- [配置说明](#配置说明)
- [最佳实践](#最佳实践)

---

## 快速开始

### 1. 启用 WebSocket

修改 `config.yaml`：

```yaml
websocket:
  enable: true
  port: "8081"           # 独立启动时的端口
  max_connections: 10000
```

### 2. 启动服务

```bash
go run main.go
```

### 3. 前端连接

```javascript
const ws = new WebSocket('ws://localhost:8889/api/ws/connect?user_id=user123');
```

---

## 启动方式

### 方式一：跟随框架启动

配置 `websocket.enable: true` 后，WebSocket 服务会集成到主服务中。

```bash
go run main.go
```

WebSocket 端点: `ws://localhost:{app.port}/api/ws/connect`

### 方式二：独立脚本启动

独立启动 WebSocket 服务：

```bash
go run cmd/websocket/main.go
```

WebSocket 端点: `ws://localhost:{websocket.port}/api/ws/connect`

**独立启动适用场景**：
- 需要独立扩展 WebSocket 服务
- WebSocket 需要更高的资源配置
- 微服务架构部署

---

## API 接口

### 建立连接

```
GET /api/ws/connect?user_id={user_id}
```

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | string | 否 | 用户标识，用于定向推送 |

**响应**：WebSocket 连接升级（101 Switching Protocols）

### 获取状态

```
GET /api/ws/status
```

**响应示例**：

```json
{
    "online_count": 150,
    "online_users": ["user1", "user2", "user3"]
}
```

### 广播消息

```
POST /api/ws/broadcast
```

**请求体**：

```json
{
    "type": "notification",
    "content": "系统维护通知"
}
```

### 发送给指定用户

```
POST /api/ws/send
```

**请求体**：

```json
{
    "type": "private_message",
    "to": "user123",
    "content": "Hello!"
}
```

---

## 前端连接

### JavaScript 原生 WebSocket

```javascript
// 建立连接
const ws = new WebSocket('ws://localhost:8889/api/ws/connect?user_id=user123');

// 连接成功
ws.onopen = () => {
    console.log('WebSocket connected');

    // 发送消息
    ws.send(JSON.stringify({
        type: 'chat',
        to: 'user456',  // 私信目标，留空则广播
        content: 'Hello World!'
    }));
};

// 接收消息
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);

    switch (message.type) {
        case 'notification':
            showNotification(message.content);
            break;
        case 'chat':
            appendChatMessage(message);
            break;
        default:
            console.log('Unknown message type:', message.type);
    }
};

// 连接关闭
ws.onclose = (event) => {
    console.log('WebSocket closed:', event.code, event.reason);
    // 自动重连逻辑
    setTimeout(() => reconnect(), 3000);
};

// 连接错误
ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};
```

### 带自动重连的封装

```javascript
class WebSocketClient {
    constructor(url, options = {}) {
        this.url = url;
        this.reconnectInterval = options.reconnectInterval || 3000;
        this.maxReconnectAttempts = options.maxReconnectAttempts || 10;
        this.reconnectAttempts = 0;
        this.handlers = {};
        this.connect();
    }

    connect() {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
            console.log('Connected');
            this.reconnectAttempts = 0;
            this.trigger('open');
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.trigger('message', message);
            this.trigger(message.type, message);
        };

        this.ws.onclose = () => {
            this.trigger('close');
            this.reconnect();
        };

        this.ws.onerror = (error) => {
            this.trigger('error', error);
        };
    }

    reconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log('Max reconnect attempts reached');
            return;
        }

        this.reconnectAttempts++;
        console.log(`Reconnecting... (${this.reconnectAttempts})`);
        setTimeout(() => this.connect(), this.reconnectInterval);
    }

    send(data) {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        }
    }

    on(event, handler) {
        if (!this.handlers[event]) {
            this.handlers[event] = [];
        }
        this.handlers[event].push(handler);
    }

    trigger(event, data) {
        if (this.handlers[event]) {
            this.handlers[event].forEach(handler => handler(data));
        }
    }

    close() {
        this.maxReconnectAttempts = 0;  // 禁止重连
        this.ws.close();
    }
}

// 使用示例
const client = new WebSocketClient('ws://localhost:8889/api/ws/connect?user_id=user123');

client.on('open', () => {
    console.log('Connected!');
});

client.on('notification', (msg) => {
    alert(msg.content);
});

client.on('chat', (msg) => {
    console.log(`${msg.from}: ${msg.content}`);
});

// 发送消息
client.send({
    type: 'chat',
    to: 'user456',
    content: 'Hello!'
});
```

---

## 服务端推送

### 在 Service 中推送消息

```go
// app/services/notification_service.go
package services

import (
    "gin-web/pkg/websocket"
)

type NotificationService struct {
    wsManager *websocket.Manager
}

func NewNotificationService(wsManager *websocket.Manager) *NotificationService {
    return &NotificationService{wsManager: wsManager}
}

// 发送给指定用户
func (s *NotificationService) SendToUser(userID string, title, content string) {
    s.wsManager.SendToUser(userID, &websocket.Message{
        Type: "notification",
        Content: map[string]string{
            "title":   title,
            "content": content,
        },
    })
}

// 广播给所有用户
func (s *NotificationService) Broadcast(title, content string) {
    s.wsManager.Broadcast(&websocket.Message{
        Type: "notification",
        Content: map[string]string{
            "title":   title,
            "content": content,
        },
    })
}
```

### 在 Controller 中推送

```go
// 订单支付成功后通知用户
func (c *OrderController) PayCallback(ctx *gin.Context) {
    orderID := ctx.Param("id")

    // 处理支付逻辑...
    order, _ := c.orderService.ConfirmPayment(orderID)

    // 推送 WebSocket 通知
    c.wsManager.SendToUser(order.UserID, &websocket.Message{
        Type: "order_paid",
        Content: map[string]interface{}{
            "order_id": order.ID,
            "amount":   order.Amount,
        },
    })

    response.Success(ctx, nil)
}
```

### 在 Consumer 中推送

```go
// app/amqp/consumer/order_consumer.go
type OrderConsumer struct {
    wsManager *websocket.Manager
}

func (c *OrderConsumer) HandleMessage(msg amqp.Delivery) error {
    var order Order
    json.Unmarshal(msg.Body, &order)

    // 处理订单...

    // 推送状态更新
    c.wsManager.SendToUser(order.UserID, &websocket.Message{
        Type: "order_status",
        Content: map[string]interface{}{
            "order_id": order.ID,
            "status":   order.Status,
        },
    })

    return nil
}
```

---

## 消息格式

### 标准消息结构

```go
type Message struct {
    Type    string      `json:"type"`              // 消息类型
    To      string      `json:"to,omitempty"`      // 接收者 ID（可选）
    From    string      `json:"from,omitempty"`    // 发送者 ID（可选）
    Content interface{} `json:"content"`           // 消息内容
}
```

### 消息类型约定

| Type | 说明 | 方向 |
|------|------|------|
| `notification` | 系统通知 | 服务器 -> 客户端 |
| `chat` | 聊天消息 | 双向 |
| `order_status` | 订单状态更新 | 服务器 -> 客户端 |
| `typing` | 正在输入 | 客户端 -> 客户端 |
| `ping` | 心跳检测 | 双向 |

### 示例消息

```json
// 系统通知
{
    "type": "notification",
    "content": {
        "title": "系统公告",
        "message": "系统将于今晚 22:00 进行维护"
    }
}

// 聊天消息
{
    "type": "chat",
    "from": "user123",
    "to": "user456",
    "content": {
        "text": "你好！",
        "timestamp": 1699999999
    }
}

// 订单状态
{
    "type": "order_status",
    "content": {
        "order_id": "ORD123456",
        "status": "shipped",
        "tracking_no": "SF123456789"
    }
}
```

---

## 配置说明

### config.yaml

```yaml
websocket:
  enable: true           # 框架启动时是否启用 WebSocket
  port: "8081"           # 独立启动时的端口
  max_connections: 10000 # 最大连接数
```

### 配置结构体

```go
// config/websocket.go
type WebSocket struct {
    Enable         bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
    Port           string `mapstructure:"port" json:"port" yaml:"port"`
    MaxConnections int    `mapstructure:"max_connections" json:"max_connections" yaml:"max_connections"`
}
```

---

## 最佳实践

### 1. 用户认证

结合 JWT 进行用户认证：

```go
// 在 WebSocket 控制器中
func (c *WebSocketController) Connect(ctx *gin.Context) {
    // 从 JWT 中间件获取用户 ID
    userID := ctx.GetString("id")
    if userID == "" {
        // 或者从查询参数获取（不推荐生产环境使用）
        userID = ctx.Query("user_id")
    }

    if err := c.manager.HandleRequest(ctx.Writer, ctx.Request, userID); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    }
}
```

### 2. 心跳检测

Melody 自动处理 ping/pong 心跳，默认配置：

```go
m.melody.Config.PingPeriod = 54 * time.Second
m.melody.Config.PongWait = 60 * time.Second
```

### 3. 消息大小限制

限制消息大小防止恶意攻击：

```go
m.melody.Config.MaxMessageSize = 512 * 1024  // 512KB
```

### 4. 错误处理

```go
m.melody.HandleError(func(s *melody.Session, err error) {
    userID, _ := s.Get("user_id")
    m.log.Error("websocket error",
        zap.Any("user_id", userID),
        zap.Error(err))
})
```

### 5. 房间/频道支持

扩展 Manager 支持房间功能：

```go
type Manager struct {
    // ...
    rooms map[string]map[*melody.Session]bool  // 房间 -> 会话集合
}

func (m *Manager) JoinRoom(s *melody.Session, roomID string) {
    m.mu.Lock()
    if m.rooms[roomID] == nil {
        m.rooms[roomID] = make(map[*melody.Session]bool)
    }
    m.rooms[roomID][s] = true
    m.mu.Unlock()
}

func (m *Manager) LeaveRoom(s *melody.Session, roomID string) {
    m.mu.Lock()
    delete(m.rooms[roomID], s)
    if len(m.rooms[roomID]) == 0 {
        delete(m.rooms, roomID)
    }
    m.mu.Unlock()
}

func (m *Manager) BroadcastToRoom(roomID string, message *Message) {
    m.mu.RLock()
    sessions := m.rooms[roomID]
    m.mu.RUnlock()

    data, _ := json.Marshal(message)
    for session := range sessions {
        session.Write(data)
    }
}
```

---

## 常见问题

### Q: 连接经常断开？

A: 检查以下几点：
1. 网络稳定性
2. 反向代理（Nginx）的超时配置
3. 客户端是否正确处理 ping/pong

Nginx 配置示例：

```nginx
location /api/ws/ {
    proxy_pass http://backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;
}
```

### Q: 如何处理断线重连？

A: 客户端实现自动重连逻辑（见前端示例），服务端无状态，重连后重新建立会话即可。

### Q: 如何限制连接数？

A: 在连接时检查：

```go
m.melody.HandleConnect(func(s *melody.Session) {
    if m.melody.Len() >= maxConnections {
        s.Close()
        return
    }
    // ...
})
```

---

## 参考资源

- [Melody 官方文档](https://github.com/olahol/melody)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [MDN WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
