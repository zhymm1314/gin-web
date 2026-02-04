# P3 - 核心功能增强 (Core Enhancement)

> 优先级：高
> 预计工时：3-4 天
> 影响范围：开发体验 & 核心功能

---

## 概述

P3 阶段专注于三个核心功能的实现：API 文档自动生成、定时任务封装、WebSocket 封装。这些功能对于生产环境的 Web 应用至关重要。

---

## TODO 列表

### 1. Swagger API 文档 (最高优先级)

- [ ] **任务完成**

**目标**: 启动项目后直接通过 URL 访问 API 文档

#### 1.1 安装依赖

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

#### 1.2 main.go 添加 Swagger 注释

```go
// @title           Gin-Web API
// @version         1.5.0
// @description     Gin-Web 脚手架 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入 Bearer {token}

func main() {
    // ...
}
```

#### 1.3 控制器方法添加注释

```go
// Register 用户注册
// @Summary      用户注册
// @Description  创建新用户账号
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body request.Register true "注册信息"
// @Success      200 {object} response.Response{data=models.User}
// @Failure      400 {object} response.Response
// @Router       /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
    // ...
}
```

#### 1.4 注册 Swagger 路由

**文件**: `bootstrap/router.go`

```go
import (
    _ "gin-web/docs"  // swagger docs
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(ctrls ...controllers.Controller) *gin.Engine {
    // ...

    // Swagger 文档 (非生产环境)
    if global.App.Config.App.Env != "production" {
        router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    return router
}
```

#### 1.5 生成并访问文档

```bash
# 生成文档
swag init

# 启动服务后访问
# http://localhost:8080/swagger/index.html
```

#### 1.6 Makefile 命令

```makefile
swagger:
	swag init
	@echo "Swagger docs generated"
	@echo "Visit: http://localhost:8080/swagger/index.html"
```

**预期效果**:
- 启动项目后访问 `http://localhost:8080/swagger/index.html`
- 自动展示所有 API 接口文档
- 支持在线调试 API

---

### 2. 定时任务封装 (Cron Job)

- [ ] **任务完成**

**目标**: 提供简单易用的定时任务封装，支持独立启动和框架集成启动两种模式

#### 2.1 技术选型

推荐使用 [robfig/cron](https://github.com/robfig/cron) v3，这是 Go 生态中最成熟的定时任务库。

```bash
go get github.com/robfig/cron/v3
```

**选型理由**:
- Star 数 12k+，社区活跃
- 支持标准 cron 表达式
- 支持秒级调度
- 支持时区设置
- 轻量无依赖

#### 2.2 定时任务管理器

**新建文件**: `pkg/cron/manager.go`

```go
package cron

import (
    "gin-web/global"
    "github.com/robfig/cron/v3"
    "go.uber.org/zap"
    "sync"
)

type JobHandler interface {
    Name() string
    Spec() string  // cron 表达式
    Run()
}

type Manager struct {
    cron     *cron.Cron
    jobs     map[string]cron.EntryID
    handlers []JobHandler
    mu       sync.RWMutex
    log      *zap.Logger
}

func NewManager(log *zap.Logger) *Manager {
    return &Manager{
        cron: cron.New(cron.WithSeconds()),  // 支持秒级
        jobs: make(map[string]cron.EntryID),
        log:  log,
    }
}

func (m *Manager) Register(handler JobHandler) {
    m.handlers = append(m.handlers, handler)
}

func (m *Manager) Start() error {
    for _, handler := range m.handlers {
        entryID, err := m.cron.AddFunc(handler.Spec(), func() {
            defer func() {
                if r := recover(); r != nil {
                    m.log.Error("cron job panic", zap.Any("error", r))
                }
            }()
            handler.Run()
        })
        if err != nil {
            m.log.Error("add cron job failed",
                zap.String("name", handler.Name()),
                zap.Error(err))
            continue
        }
        m.mu.Lock()
        m.jobs[handler.Name()] = entryID
        m.mu.Unlock()
        m.log.Info("cron job registered",
            zap.String("name", handler.Name()),
            zap.String("spec", handler.Spec()))
    }
    m.cron.Start()
    m.log.Info("cron manager started")
    return nil
}

func (m *Manager) Stop() {
    m.cron.Stop()
    m.log.Info("cron manager stopped")
}
```

#### 2.3 定时任务示例

**新建文件**: `app/cron/cleanup_job.go`

```go
package cron

import (
    "gin-web/global"
    "go.uber.org/zap"
)

// CleanupJob 清理过期数据任务
type CleanupJob struct{}

func (j *CleanupJob) Name() string {
    return "cleanup_expired_tokens"
}

func (j *CleanupJob) Spec() string {
    return "0 0 2 * * *"  // 每天凌晨 2 点执行
}

func (j *CleanupJob) Run() {
    global.App.Log.Info("running cleanup job")
    // 清理逻辑...
}
```

**新建文件**: `app/cron/health_check_job.go`

```go
package cron

import (
    "gin-web/global"
)

// HealthCheckJob 健康检查任务
type HealthCheckJob struct{}

func (j *HealthCheckJob) Name() string {
    return "health_check"
}

func (j *HealthCheckJob) Spec() string {
    return "*/30 * * * * *"  // 每 30 秒执行
}

func (j *HealthCheckJob) Run() {
    global.App.Log.Debug("health check running")
    // 健康检查逻辑...
}
```

#### 2.4 启动方式一：跟随框架启动

**修改 main.go**:

```go
import (
    appCron "gin-web/app/cron"
    "gin-web/pkg/cron"
)

func main() {
    // ... 其他初始化 ...

    // 初始化定时任务 (可选)
    if global.App.Config.App.EnableCron {
        cronManager := cron.NewManager(global.App.Log)
        cronManager.Register(&appCron.CleanupJob{})
        cronManager.Register(&appCron.HealthCheckJob{})
        cronManager.Start()
        defer cronManager.Stop()
    }

    // 启动服务器
    bootstrap.RunServer(app.GetControllers()...)
}
```

#### 2.5 启动方式二：独立脚本启动

**新建文件**: `cmd/cron/main.go`

```go
package main

import (
    appCron "gin-web/app/cron"
    "gin-web/bootstrap"
    "gin-web/global"
    "gin-web/pkg/cron"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 初始化配置和日志
    bootstrap.InitializeConfig()
    global.App.Log = bootstrap.InitializeLog()
    global.App.DB = bootstrap.InitializeDB()
    global.App.Redis = bootstrap.InitializeRedis()

    // 创建定时任务管理器
    cronManager := cron.NewManager(global.App.Log)

    // 注册定时任务
    cronManager.Register(&appCron.CleanupJob{})
    cronManager.Register(&appCron.HealthCheckJob{})

    // 启动
    cronManager.Start()
    global.App.Log.Info("Cron service started")

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    cronManager.Stop()
    global.App.Log.Info("Cron service stopped")
}
```

**运行独立脚本**:

```bash
# 独立启动定时任务服务
go run cmd/cron/main.go
```

#### 2.6 配置文件

**config.yaml 添加**:

```yaml
app:
  # ...
  enable_cron: true  # 是否启用定时任务 (框架集成模式)
```

---

### 3. WebSocket 封装 (基于 Melody)

- [ ] **任务完成**

**目标**: 使用成熟的 Melody 库实现 WebSocket 封装，支持独立启动和框架集成启动

#### 3.1 技术选型

推荐使用 [Melody](https://github.com/olahol/melody)，这是 Go 生态中与 Gin 风格最接近的 WebSocket 库。

```bash
go get github.com/olahol/melody
```

**选型理由**:
- Star 3.5k+，成熟稳定
- API 风格与 Gin 一致，学习成本低
- 基于 gorilla/websocket，性能可靠
- 自动处理 ping/pong 心跳和超时
- 内置广播、会话过滤、并发安全
- 代码简洁，易于维护

**备选方案**: 如需支持百万级连接或复杂 Pub/Sub，可考虑 [Centrifuge](https://github.com/centrifugal/centrifuge)

#### 3.2 WebSocket 管理器封装

**新建文件**: `pkg/websocket/manager.go`

```go
package websocket

import (
    "encoding/json"
    "github.com/olahol/melody"
    "go.uber.org/zap"
    "net/http"
    "sync"
)

// Message WebSocket 消息结构
type Message struct {
    Type    string      `json:"type"`
    To      string      `json:"to,omitempty"`
    From    string      `json:"from,omitempty"`
    Content interface{} `json:"content"`
}

// Manager WebSocket 管理器
type Manager struct {
    melody      *melody.Melody
    userSessions map[string]map[*melody.Session]bool
    mu          sync.RWMutex
    log         *zap.Logger
}

// NewManager 创建 WebSocket 管理器
func NewManager(log *zap.Logger) *Manager {
    m := &Manager{
        melody:       melody.New(),
        userSessions: make(map[string]map[*melody.Session]bool),
        log:          log,
    }

    // 配置 Melody
    m.melody.Config.MaxMessageSize = 512 * 1024  // 512KB
    m.melody.Config.MessageBufferSize = 256

    // 注册事件处理器
    m.setupHandlers()

    return m
}

func (m *Manager) setupHandlers() {
    // 连接建立
    m.melody.HandleConnect(func(s *melody.Session) {
        userID, _ := s.Get("user_id")
        m.log.Info("client connected", zap.Any("user_id", userID))

        if uid, ok := userID.(string); ok && uid != "" {
            m.mu.Lock()
            if m.userSessions[uid] == nil {
                m.userSessions[uid] = make(map[*melody.Session]bool)
            }
            m.userSessions[uid][s] = true
            m.mu.Unlock()
        }
    })

    // 连接断开
    m.melody.HandleDisconnect(func(s *melody.Session) {
        userID, _ := s.Get("user_id")
        m.log.Info("client disconnected", zap.Any("user_id", userID))

        if uid, ok := userID.(string); ok && uid != "" {
            m.mu.Lock()
            delete(m.userSessions[uid], s)
            if len(m.userSessions[uid]) == 0 {
                delete(m.userSessions, uid)
            }
            m.mu.Unlock()
        }
    })

    // 收到消息
    m.melody.HandleMessage(func(s *melody.Session, msg []byte) {
        var message Message
        if err := json.Unmarshal(msg, &message); err != nil {
            m.log.Error("invalid message format", zap.Error(err))
            return
        }

        userID, _ := s.Get("user_id")
        if uid, ok := userID.(string); ok {
            message.From = uid
        }

        // 处理消息路由
        if message.To != "" {
            m.SendToUser(message.To, &message)
        } else {
            m.Broadcast(&message)
        }
    })

    // 错误处理
    m.melody.HandleError(func(s *melody.Session, err error) {
        m.log.Error("websocket error", zap.Error(err))
    })
}

// HandleRequest 处理 WebSocket 升级请求
func (m *Manager) HandleRequest(w http.ResponseWriter, r *http.Request, userID string) error {
    return m.melody.HandleRequestWithKeys(w, r, map[string]interface{}{
        "user_id": userID,
    })
}

// Broadcast 广播消息给所有客户端
func (m *Manager) Broadcast(message *Message) {
    data, _ := json.Marshal(message)
    m.melody.Broadcast(data)
}

// BroadcastFilter 按条件广播
func (m *Manager) BroadcastFilter(message *Message, filter func(s *melody.Session) bool) {
    data, _ := json.Marshal(message)
    m.melody.BroadcastFilter(data, filter)
}

// SendToUser 发送消息给指定用户
func (m *Manager) SendToUser(userID string, message *Message) {
    m.mu.RLock()
    sessions := m.userSessions[userID]
    m.mu.RUnlock()

    if len(sessions) == 0 {
        return
    }

    data, _ := json.Marshal(message)
    for session := range sessions {
        session.Write(data)
    }
}

// OnlineCount 获取在线人数
func (m *Manager) OnlineCount() int {
    return m.melody.Len()
}

// OnlineUsers 获取在线用户列表
func (m *Manager) OnlineUsers() []string {
    m.mu.RLock()
    defer m.mu.RUnlock()

    users := make([]string, 0, len(m.userSessions))
    for userID := range m.userSessions {
        users = append(users, userID)
    }
    return users
}

// Close 关闭管理器
func (m *Manager) Close() {
    m.melody.Close()
}
```

#### 3.3 WebSocket 控制器

**新建文件**: `app/controllers/websocket_controller.go`

```go
package controllers

import (
    "gin-web/pkg/websocket"
    "github.com/gin-gonic/gin"
    "net/http"
)

type WebSocketController struct {
    manager *websocket.Manager
}

func NewWebSocketController(manager *websocket.Manager) *WebSocketController {
    return &WebSocketController{manager: manager}
}

func (c *WebSocketController) Prefix() string {
    return "/ws"
}

func (c *WebSocketController) Routes() []Route {
    return []Route{
        {Method: "GET", Path: "/connect", Handler: c.Connect},
        {Method: "GET", Path: "/status", Handler: c.Status},
        {Method: "POST", Path: "/broadcast", Handler: c.Broadcast},
        {Method: "POST", Path: "/send", Handler: c.SendToUser},
    }
}

// Connect WebSocket 连接
// @Summary      WebSocket 连接
// @Description  建立 WebSocket 连接
// @Tags         WebSocket
// @Param        user_id query string false "用户ID"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws/connect [get]
func (c *WebSocketController) Connect(ctx *gin.Context) {
    userID := ctx.Query("user_id")
    // 或从 JWT 中间件获取: userID := ctx.GetString("id")

    if err := c.manager.HandleRequest(ctx.Writer, ctx.Request, userID); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    }
}

// Status 获取 WebSocket 状态
// @Summary      WebSocket 状态
// @Description  获取在线人数等状态信息
// @Tags         WebSocket
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /ws/status [get]
func (c *WebSocketController) Status(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, gin.H{
        "online_count": c.manager.OnlineCount(),
        "online_users": c.manager.OnlineUsers(),
    })
}

// Broadcast 广播消息
// @Summary      广播消息
// @Description  向所有在线用户广播消息
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body websocket.Message true "消息内容"
// @Success      200 {object} map[string]interface{}
// @Router       /ws/broadcast [post]
func (c *WebSocketController) Broadcast(ctx *gin.Context) {
    var msg websocket.Message
    if err := ctx.ShouldBindJSON(&msg); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.manager.Broadcast(&msg)
    ctx.JSON(http.StatusOK, gin.H{"message": "broadcast sent"})
}

// SendToUser 发送消息给指定用户
// @Summary      发送消息给用户
// @Description  向指定用户发送消息
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body websocket.Message true "消息内容"
// @Success      200 {object} map[string]interface{}
// @Router       /ws/send [post]
func (c *WebSocketController) SendToUser(ctx *gin.Context) {
    var msg websocket.Message
    if err := ctx.ShouldBindJSON(&msg); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if msg.To == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "to field is required"})
        return
    }

    c.manager.SendToUser(msg.To, &msg)
    ctx.JSON(http.StatusOK, gin.H{"message": "message sent"})
}
```

#### 3.4 启动方式一：跟随框架启动

**修改 main.go**:

```go
import (
    "gin-web/pkg/websocket"
)

func main() {
    // ... 其他初始化 ...

    // 初始化 WebSocket (根据配置)
    var wsManager *websocket.Manager
    var wsController *controllers.WebSocketController
    if global.App.Config.WebSocket.Enable {
        wsManager = websocket.NewManager(global.App.Log)
        wsController = controllers.NewWebSocketController(wsManager)
        global.App.Log.Info("WebSocket manager started")
    }

    // 组装控制器列表
    allControllers := app.GetControllers()
    if wsController != nil {
        allControllers = append(allControllers, wsController)
    }

    // 启动服务器
    bootstrap.RunServer(allControllers...)

    // 清理资源
    if wsManager != nil {
        wsManager.Close()
    }
}
```

#### 3.5 启动方式二：独立脚本启动

**新建文件**: `cmd/websocket/main.go`

```go
package main

import (
    "gin-web/app/controllers"
    "gin-web/app/middleware"
    "gin-web/bootstrap"
    "gin-web/global"
    "gin-web/pkg/websocket"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 初始化
    bootstrap.InitializeConfig()
    global.App.Log = bootstrap.InitializeLog()
    global.App.Redis = bootstrap.InitializeRedis()

    // 创建 WebSocket Manager
    wsManager := websocket.NewManager(global.App.Log)

    // 创建 Gin 路由
    router := gin.New()
    router.Use(gin.Logger(), middleware.CustomRecovery())
    router.Use(middleware.Cors())

    // 注册 WebSocket 控制器
    wsController := controllers.NewWebSocketController(wsManager)
    controllers.RegisterController(router.Group("/api"), wsController)

    // 健康检查
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":       "ok",
            "online_count": wsManager.OnlineCount(),
        })
    })

    // 获取端口
    port := global.App.Config.WebSocket.Port
    if port == "" {
        port = "8081"
    }

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }

    go func() {
        global.App.Log.Info("WebSocket server starting on port " + port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    wsManager.Close()
    global.App.Log.Info("WebSocket server stopped")
}
```

#### 3.6 使用示例

**前端连接示例**:

```javascript
// 建立连接
const ws = new WebSocket('ws://localhost:8080/api/ws/connect?user_id=user123');

// 接收消息
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
};

// 发送消息
ws.send(JSON.stringify({
    type: 'chat',
    to: 'user456',  // 私信，留空则广播
    content: 'Hello!'
}));
```

**服务端主动推送**:

```go
// 在任意 Service 或 Controller 中
wsManager.SendToUser("user123", &websocket.Message{
    Type:    "notification",
    Content: "You have a new message",
})

// 广播给所有人
wsManager.Broadcast(&websocket.Message{
    Type:    "system",
    Content: "Server will restart in 5 minutes",
})
```

#### 3.7 配置文件

**config.yaml 添加**:

```yaml
websocket:
  enable: true
  port: "8081"              # 独立启动时的端口
  max_connections: 10000
```

---

## 目录结构变更

```
gin-web/
├── cmd/                          # 独立启动脚本
│   ├── consumer/
│   │   └── main.go              # RabbitMQ 消费者独立启动
│   ├── cron/
│   │   └── main.go              # 定时任务独立启动
│   └── websocket/
│       └── main.go              # WebSocket 独立启动
├── config/
│   ├── cron.go                  # 定时任务配置 (新增)
│   └── websocket.go             # WebSocket 配置 (新增)
├── pkg/                          # 公共包
│   ├── cron/
│   │   └── manager.go           # 定时任务管理器
│   └── websocket/
│       ├── hub.go               # WebSocket Hub
│       └── client.go            # WebSocket Client
├── app/
│   ├── cron/                    # 定时任务实现
│   │   ├── cleanup_job.go
│   │   └── health_check_job.go
│   └── controllers/
│       └── websocket_controller.go
└── docs/                         # Swagger 生成的文档
    ├── docs.go
    ├── swagger.json
    └── swagger.yaml
```

---

## 完成检查清单

### Swagger 文档
- [ ] 安装 swag 工具
- [ ] main.go 添加 Swagger 注释
- [ ] 所有控制器方法添加 API 注释
- [ ] 注册 Swagger 路由
- [ ] 运行 `swag init` 生成文档
- [ ] 验证 `http://localhost:8080/swagger/index.html` 可访问

### 定时任务
- [ ] 安装 robfig/cron
- [ ] 实现 cron Manager
- [ ] 创建示例定时任务
- [ ] 创建独立启动脚本 `cmd/cron/main.go`
- [ ] 添加配置项 `cron.enable`
- [ ] 框架集成启动模式测试
- [ ] 独立脚本启动模式测试

### WebSocket (Melody)
- [ ] 安装 olahol/melody
- [ ] 实现 WebSocket Manager 封装
- [ ] 创建 WebSocket 控制器
- [ ] 创建独立启动脚本 `cmd/websocket/main.go`
- [ ] 添加配置项 `websocket.enable`
- [ ] 框架集成启动模式测试
- [ ] 独立脚本启动模式测试
- [ ] 前端连接测试

### RabbitMQ 消费者
- [ ] 创建独立启动脚本 `cmd/consumer/main.go`
- [ ] 添加配置项 `rabbitmq.enable`
- [ ] 框架集成启动模式测试
- [ ] 独立脚本启动模式测试

### 配置管理
- [ ] 新增 `config/cron.go`
- [ ] 新增 `config/websocket.go`
- [ ] 更新 `config/rabbitmq.go` 添加 enable 字段
- [ ] 更新 `config/config.go` 添加新配置结构
- [ ] 更新 `config.yaml` 添加所有启动开关

---

### 4. RabbitMQ 消费者启动优化

- [ ] **任务完成**

**目标**: 为 RabbitMQ 消费者添加独立启动模式和配置控制

#### 4.1 启动方式一：跟随框架启动 (已实现)

当前 main.go 中已有实现，需添加配置开关控制：

```go
// main.go
if global.App.Config.RabbitMQ.Enable {
    if cm := bootstrap.InitRabbitmq(); cm != nil {
        defer cm.Stop()
        global.App.Log.Info("RabbitMQ consumer manager started")
    }
}
```

#### 4.2 启动方式二：独立脚本启动

**新建文件**: `cmd/consumer/main.go`

```go
package main

import (
    "gin-web/bootstrap"
    "gin-web/global"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 初始化配置和日志
    bootstrap.InitializeConfig()
    global.App.Log = bootstrap.InitializeLog()
    global.App.DB = bootstrap.InitializeDB()
    global.App.Redis = bootstrap.InitializeRedis()

    // 启动 RabbitMQ 消费者
    cm := bootstrap.InitRabbitmq()
    if cm == nil {
        global.App.Log.Fatal("Failed to start RabbitMQ consumer manager")
    }
    global.App.Log.Info("RabbitMQ consumer service started")

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    cm.Stop()
    global.App.Log.Info("RabbitMQ consumer service stopped")
}
```

**运行独立脚本**:

```bash
# 独立启动 RabbitMQ 消费者服务
go run cmd/consumer/main.go
```

---

## 统一配置管理

所有服务的启动开关统一在 `config.yaml` 中配置：

```yaml
# config.yaml

app:
  env: local
  port: 8080
  app_name: gin-web

# RabbitMQ 配置
rabbitmq:
  enable: true                    # 框架启动时是否启用消费者
  host: localhost
  port: 5672
  username: guest
  password: guest
  vhost: /

# 定时任务配置
cron:
  enable: true                    # 框架启动时是否启用定时任务

# WebSocket 配置
websocket:
  enable: true                    # 框架启动时是否启用 WebSocket
  port: 8081                      # 独立启动时的端口
  max_connections: 10000
```

### 配置结构体更新

**文件**: `config/app.go`

```go
type App struct {
    Env     string `mapstructure:"env" json:"env" yaml:"env"`
    Port    string `mapstructure:"port" json:"port" yaml:"port"`
    AppName string `mapstructure:"app_name" json:"app_name" yaml:"app_name"`
}
```

**文件**: `config/rabbitmq.go` (更新)

```go
type RabbitMQ struct {
    Enable   bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
    Host     string `mapstructure:"host" json:"host" yaml:"host"`
    Port     int    `mapstructure:"port" json:"port" yaml:"port"`
    Username string `mapstructure:"username" json:"username" yaml:"username"`
    Password string `mapstructure:"password" json:"password" yaml:"password"`
    Vhost    string `mapstructure:"vhost" json:"vhost" yaml:"vhost"`
}
```

**新建文件**: `config/cron.go`

```go
package config

type Cron struct {
    Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`
}
```

**新建文件**: `config/websocket.go`

```go
package config

type WebSocket struct {
    Enable         bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
    Port           string `mapstructure:"port" json:"port" yaml:"port"`
    MaxConnections int    `mapstructure:"max_connections" json:"max_connections" yaml:"max_connections"`
}
```

**文件**: `config/config.go` (更新)

```go
type Configuration struct {
    App       App       `mapstructure:"app" json:"app" yaml:"app"`
    Log       Log       `mapstructure:"log" json:"log" yaml:"log"`
    Database  Database  `mapstructure:"database" json:"database" yaml:"database"`
    Jwt       Jwt       `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
    Redis     Redis     `mapstructure:"redis" json:"redis" yaml:"redis"`
    RabbitMQ  RabbitMQ  `mapstructure:"rabbitmq" json:"rabbitMQ" yaml:"rabbitMQ"`
    Cron      Cron      `mapstructure:"cron" json:"cron" yaml:"cron"`
    WebSocket WebSocket `mapstructure:"websocket" json:"websocket" yaml:"websocket"`
    ApiUrls   ApiUrls   `mapstructure:"api_url" json:"api_url" yaml:"api_url"`
}
```

### main.go 统一启动逻辑

```go
func main() {
    // ... 初始化代码 ...

    // 使用 Wire 初始化应用
    app, err := container.InitializeApp()
    if err != nil {
        global.App.Log.Fatal("Failed to initialize app: " + err.Error())
    }

    // 启动 RabbitMQ 消费者 (根据配置)
    var consumerManager *bootstrap.ConsumerManager
    if global.App.Config.RabbitMQ.Enable {
        consumerManager = bootstrap.InitRabbitmq()
        if consumerManager != nil {
            global.App.Log.Info("RabbitMQ consumer started")
        }
    }

    // 启动定时任务 (根据配置)
    var cronManager *cron.Manager
    if global.App.Config.Cron.Enable {
        cronManager = cron.NewManager(global.App.Log)
        cronManager.Register(&appCron.CleanupJob{})
        cronManager.Start()
        global.App.Log.Info("Cron manager started")
    }

    // 启动 WebSocket (根据配置)
    var wsHub *websocket.Hub
    var wsController *controllers.WebSocketController
    if global.App.Config.WebSocket.Enable {
        wsHub = websocket.NewHub(global.App.Log)
        go wsHub.Run()
        wsController = controllers.NewWebSocketController(wsHub)
        global.App.Log.Info("WebSocket hub started")
    }

    // 组装控制器列表
    allControllers := app.GetControllers()
    if wsController != nil {
        allControllers = append(allControllers, wsController)
    }

    // 启动 HTTP 服务器
    bootstrap.RunServer(allControllers...)

    // 清理资源 (defer 或信号处理中)
    if consumerManager != nil {
        consumerManager.Stop()
    }
    if cronManager != nil {
        cronManager.Stop()
    }
}
```

---

## 启动模式总结

| 功能 | 配置项 | 框架集成启动 | 独立脚本启动 |
|------|--------|-------------|-------------|
| **HTTP API** | - | `go run main.go` | - |
| **RabbitMQ 消费者** | `rabbitmq.enable` | `main.go` | `go run cmd/consumer/main.go` |
| **定时任务** | `cron.enable` | `main.go` | `go run cmd/cron/main.go` |
| **WebSocket** | `websocket.enable` | `main.go` | `go run cmd/websocket/main.go` |

**配置示例**:

```yaml
# 开发环境 - 全部启用
rabbitmq:
  enable: true
cron:
  enable: true
websocket:
  enable: true

# 生产环境 - 独立部署，关闭框架集成
rabbitmq:
  enable: false   # 使用 cmd/consumer/main.go 独立启动
cron:
  enable: false   # 使用 cmd/cron/main.go 独立启动
websocket:
  enable: false   # 使用 cmd/websocket/main.go 独立启动
```

**推荐使用场景**:
- **开发环境**: 全部 enable: true，一个命令启动所有服务
- **生产环境**: 全部 enable: false，各服务独立启动，便于扩展和部署

---

## 依赖更新

```bash
# Swagger
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
go install github.com/swaggo/swag/cmd/swag@latest

# 定时任务
go get github.com/robfig/cron/v3

# WebSocket (Melody - 基于 gorilla/websocket 的高层封装)
go get github.com/olahol/melody
```

---

## 参考资源

- [Melody - Minimalist WebSocket Framework](https://github.com/olahol/melody)
- [robfig/cron - Go Cron Library](https://github.com/robfig/cron)
- [Swaggo - Swagger for Go](https://github.com/swaggo/swag)
- [Centrifuge - 大规模实时应用备选](https://github.com/centrifugal/centrifuge)
