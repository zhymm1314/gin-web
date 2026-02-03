# 🟢 P3 - 功能增强 (Enhancement)

> 优先级：优化
> 预计工时：3-5 天
> 影响范围：功能完善 & 开发体验

---

## 概述

这些优化将进一步完善项目功能，提升开发体验和生产可用性。

---

## TODO 列表

### 1. ✅ Docker 多阶段构建

- [ ] **任务完成**

**背景**:
当前 Dockerfile 使用完整的 golang 镜像，构建出的镜像约 800MB，生产环境部署效率低。

**当前 Dockerfile**:
```dockerfile
FROM golang:1.22.3-bookworm  # ~800MB
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main main.go
EXPOSE 8080
CMD ["./main"]
```

**优化后 Dockerfile**:

```dockerfile
# ==================== 构建阶段 ====================
FROM golang:1.22-alpine AS builder

# 设置必要的环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 设置工作目录
WORKDIR /build

# 复制依赖文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译应用（使用 ldflags 减小二进制大小）
RUN go build -ldflags="-s -w" -o main .

# ==================== 运行阶段 ====================
FROM alpine:3.19

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/main .

# 复制配置文件目录
COPY --from=builder /build/config ./config

# 创建日志目录
RUN mkdir -p storage/logs && chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/ping || exit 1

# 启动应用
CMD ["./main"]
```

**新增 .dockerignore**:

```
# .dockerignore
.git
.gitignore
.idea
.vscode
*.md
!README.md
Dockerfile
docker-compose*.yml
.env*
*.log
storage/logs/*
todo/
docs/
*.test
*_test.go
```

**新增 docker-compose.yml**:

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: gin-web
    ports:
      - "8080:8080"
    volumes:
      - ./config/config.yaml:/app/config/config.yaml:ro
      - ./storage/logs:/app/storage/logs
    environment:
      - GIN_MODE=release
    depends_on:
      - mysql
      - redis
      - rabbitmq
    networks:
      - gin-web-network
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    container_name: gin-web-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: gin-web
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
    networks:
      - gin-web-network
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: gin-web-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - gin-web-network
    restart: unless-stopped

  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: gin-web-rabbitmq
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin123
      RABBITMQ_DEFAULT_VHOST: /gin-web
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq-data:/var/lib/rabbitmq
    networks:
      - gin-web-network
    restart: unless-stopped

volumes:
  mysql-data:
  redis-data:
  rabbitmq-data:

networks:
  gin-web-network:
    driver: bridge
```

**效果对比**:
| 项目 | 优化前 | 优化后 |
|------|--------|--------|
| 镜像大小 | ~800MB | ~20MB |
| 构建时间 | 约2分钟 | 约1分钟 |
| 安全性 | root 用户 | 非 root 用户 |
| 健康检查 | 无 | 有 |

---

### 2. ✅ 添加更多中间件

- [ ] **任务完成**

**待添加的中间件**:

#### 2.1 请求日志中间件

**新建文件**: `app/middleware/logger.go`
```go
package middleware

import (
    "gin-web/global"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "time"
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 开始时间
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        
        // 处理请求
        c.Next()
        
        // 计算耗时
        latency := time.Since(start)
        
        // 记录日志
        global.App.Log.Info("HTTP Request",
            zap.Int("status", c.Writer.Status()),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.String("user-agent", c.Request.UserAgent()),
            zap.Duration("latency", latency),
            zap.Int("body_size", c.Writer.Size()),
        )
    }
}
```

#### 2.2 限流中间件

**新建文件**: `app/middleware/ratelimit.go`
```go
package middleware

import (
    "gin-web/app/common/response"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    "sync"
)

// IPRateLimiter IP 限流器
type IPRateLimiter struct {
    ips    map[string]*rate.Limiter
    mu     *sync.RWMutex
    rate   rate.Limit
    burst  int
}

// NewIPRateLimiter 创建 IP 限流器
func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
    return &IPRateLimiter{
        ips:   make(map[string]*rate.Limiter),
        mu:    &sync.RWMutex{},
        rate:  r,
        burst: burst,
    }
}

// GetLimiter 获取限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()
    
    limiter, exists := i.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(i.rate, i.burst)
        i.ips[ip] = limiter
    }
    
    return limiter
}

// RateLimiter 创建限流中间件
// r: 每秒允许的请求数
// burst: 突发请求数
func RateLimiter(r rate.Limit, burst int) gin.HandlerFunc {
    limiter := NewIPRateLimiter(r, burst)
    
    return func(c *gin.Context) {
        ip := c.ClientIP()
        if !limiter.GetLimiter(ip).Allow() {
            response.Fail(c, 429, "请求过于频繁，请稍后再试")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### 2.3 超时中间件

**新建文件**: `app/middleware/timeout.go`
```go
package middleware

import (
    "context"
    "gin-web/app/common/response"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()
        
        c.Request = c.Request.WithContext(ctx)
        
        finished := make(chan struct{}, 1)
        panicChan := make(chan interface{}, 1)
        
        go func() {
            defer func() {
                if p := recover(); p != nil {
                    panicChan <- p
                }
            }()
            c.Next()
            finished <- struct{}{}
        }()
        
        select {
        case <-panicChan:
            response.ServerError(c, "Internal Server Error")
        case <-finished:
            // 正常完成
        case <-ctx.Done():
            c.Writer.WriteHeader(http.StatusGatewayTimeout)
            response.Fail(c, http.StatusGatewayTimeout, "请求超时")
            c.Abort()
        }
    }
}
```

#### 2.4 链路追踪中间件

**新建文件**: `app/middleware/trace.go`
```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

const (
    TraceIDHeader = "X-Trace-ID"
    TraceIDKey    = "trace_id"
)

// Trace 链路追踪中间件
func Trace() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 优先从请求头获取 trace_id
        traceID := c.GetHeader(TraceIDHeader)
        if traceID == "" {
            traceID = uuid.New().String()
        }
        
        // 设置到上下文
        c.Set(TraceIDKey, traceID)
        
        // 设置响应头
        c.Header(TraceIDHeader, traceID)
        
        c.Next()
    }
}

// GetTraceID 获取 trace_id
func GetTraceID(c *gin.Context) string {
    if traceID, exists := c.Get(TraceIDKey); exists {
        return traceID.(string)
    }
    return ""
}
```

#### 2.5 更新路由使用中间件

```go
// bootstrap/router.go
func setupRouter() *gin.Engine {
    // ...
    router := gin.New()
    
    // 全局中间件
    router.Use(
        middleware.Trace(),                    // 链路追踪
        middleware.RequestLogger(),            // 请求日志
        middleware.CustomRecovery(),           // 异常恢复
        middleware.Cors(),                     // 跨域
        middleware.RateLimiter(100, 200),      // 限流：每秒100请求，突发200
    )
    
    // 注册路由
    apiGroup := router.Group("/api")
    apiGroup.Use(middleware.Timeout(30 * time.Second))  // API 超时控制
    routes.SetApiGroupRoutes(apiGroup)
    
    return router
}
```

---

### 3. ✅ 实现 WebSocket 封装

- [ ] **任务完成**

**背景**:
README 中提到 WebSocket 客户端封装是待完成功能。

#### 3.1 安装依赖

```bash
go get github.com/gorilla/websocket
```

#### 3.2 实现 WebSocket Hub

**新建文件**: `pkg/websocket/hub.go`
```go
package websocket

import (
    "sync"
)

// Hub 维护活跃客户端集合
type Hub struct {
    // 注册的客户端
    clients map[*Client]bool
    
    // 按用户ID索引的客户端
    userClients map[string]map[*Client]bool
    
    // 广播消息通道
    broadcast chan *Message
    
    // 注册请求通道
    register chan *Client
    
    // 注销请求通道
    unregister chan *Client
    
    mu sync.RWMutex
}

// Message WebSocket 消息
type Message struct {
    Type    string      `json:"type"`
    To      string      `json:"to,omitempty"`      // 目标用户ID，空表示广播
    From    string      `json:"from,omitempty"`
    Content interface{} `json:"content"`
}

// NewHub 创建 Hub
func NewHub() *Hub {
    return &Hub{
        clients:     make(map[*Client]bool),
        userClients: make(map[string]map[*Client]bool),
        broadcast:   make(chan *Message, 256),
        register:    make(chan *Client),
        unregister:  make(chan *Client),
    }
}

// Run 启动 Hub
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            if client.UserID != "" {
                if h.userClients[client.UserID] == nil {
                    h.userClients[client.UserID] = make(map[*Client]bool)
                }
                h.userClients[client.UserID][client] = true
            }
            h.mu.Unlock()
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                if client.UserID != "" {
                    delete(h.userClients[client.UserID], client)
                }
                close(client.send)
            }
            h.mu.Unlock()
            
        case message := <-h.broadcast:
            h.mu.RLock()
            if message.To != "" {
                // 发送给指定用户
                if clients, ok := h.userClients[message.To]; ok {
                    for client := range clients {
                        select {
                        case client.send <- message:
                        default:
                            close(client.send)
                            delete(h.clients, client)
                        }
                    }
                }
            } else {
                // 广播给所有客户端
                for client := range h.clients {
                    select {
                    case client.send <- message:
                    default:
                        close(client.send)
                        delete(h.clients, client)
                    }
                }
            }
            h.mu.RUnlock()
        }
    }
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID string, message *Message) {
    message.To = userID
    h.broadcast <- message
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message *Message) {
    h.broadcast <- message
}

// OnlineCount 获取在线人数
func (h *Hub) OnlineCount() int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.clients)
}
```

#### 3.3 实现 Client

**新建文件**: `pkg/websocket/client.go`
```go
package websocket

import (
    "encoding/json"
    "github.com/gorilla/websocket"
    "log"
    "time"
)

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 512 * 1024  // 512KB
)

// Client WebSocket 客户端
type Client struct {
    Hub    *Hub
    Conn   *websocket.Conn
    UserID string
    send   chan *Message
}

// NewClient 创建客户端
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
    return &Client{
        Hub:    hub,
        Conn:   conn,
        UserID: userID,
        send:   make(chan *Message, 256),
    }
}

// ReadPump 读取消息
func (c *Client) ReadPump() {
    defer func() {
        c.Hub.unregister <- c
        c.Conn.Close()
    }()
    
    c.Conn.SetReadLimit(maxMessageSize)
    c.Conn.SetReadDeadline(time.Now().Add(pongWait))
    c.Conn.SetPongHandler(func(string) error {
        c.Conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })
    
    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("websocket error: %v", err)
            }
            break
        }
        
        var msg Message
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("message unmarshal error: %v", err)
            continue
        }
        msg.From = c.UserID
        c.Hub.broadcast <- &msg
    }
}

// WritePump 写入消息
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            data, err := json.Marshal(message)
            if err != nil {
                log.Printf("message marshal error: %v", err)
                continue
            }
            
            if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
                return
            }
            
        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

#### 3.4 WebSocket 控制器

**新建文件**: `app/controllers/websocket.go`
```go
package controllers

import (
    "gin-web/pkg/websocket"
    "github.com/gin-gonic/gin"
    ws "github.com/gorilla/websocket"
    "net/http"
)

var upgrader = ws.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true  // 生产环境需要严格检查
    },
}

type WebSocketController struct {
    hub *websocket.Hub
}

func NewWebSocketController(hub *websocket.Hub) *WebSocketController {
    return &WebSocketController{hub: hub}
}

func (c *WebSocketController) Prefix() string {
    return "/ws"
}

func (c *WebSocketController) Routes() []Route {
    return []Route{
        {Method: "GET", Path: "/connect", Handler: c.Connect},
    }
}

func (c *WebSocketController) Connect(ctx *gin.Context) {
    userID := ctx.GetString("id")  // 从 JWT 中间件获取
    
    conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
    if err != nil {
        return
    }
    
    client := websocket.NewClient(c.hub, conn, userID)
    c.hub.register <- client
    
    go client.WritePump()
    go client.ReadPump()
}
```

---

### 4. ✅ 添加 Swagger 文档

- [ ] **任务完成**

#### 4.1 安装 swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

#### 4.2 添加注释

**更新 main.go**:
```go
// @title           Gin-Web API
// @version         1.0
// @description     Gin-Web 脚手架 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8889
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入 Bearer {token}

func main() {
    // ...
}
```

**添加 API 注释示例**:
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
func (c *UserController) Register(ctx *gin.Context) {
    // ...
}
```

#### 4.3 生成文档

```bash
swag init
```

#### 4.4 注册 Swagger 路由

```go
import (
    _ "gin-web/docs"  // swagger docs
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRouter() *gin.Engine {
    router := gin.New()
    // ...
    
    // Swagger 文档（仅非生产环境）
    if global.App.Config.App.Env != "production" {
        router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }
    
    return router
}
```

---

### 5. ✅ 添加单元测试框架

- [ ] **任务完成**

#### 5.1 测试目录结构

```
gin-web/
├── tests/
│   ├── setup_test.go       # 测试初始化
│   ├── user_test.go        # 用户测试
│   └── mocks/
│       └── user_repository_mock.go
```

#### 5.2 测试初始化

**新建文件**: `tests/setup_test.go`
```go
package tests

import (
    "gin-web/bootstrap"
    "gin-web/global"
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    // 设置测试环境
    os.Setenv("GIN_MODE", "test")
    
    // 初始化配置
    bootstrap.InitializeConfig()
    global.App.Log = bootstrap.InitializeLog()
    
    // 运行测试
    code := m.Run()
    
    os.Exit(code)
}
```

#### 5.3 Mock 工具

```bash
go install github.com/golang/mock/mockgen@latest

# 生成 mock
mockgen -source=internal/repository/repository.go -destination=tests/mocks/user_repository_mock.go -package=mocks
```

#### 5.4 Service 测试示例

**新建文件**: `tests/user_test.go`
```go
package tests

import (
    "gin-web/app/common/request"
    "gin-web/app/models"
    "gin-web/app/services"
    "gin-web/tests/mocks"
    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"
    "testing"
)

func TestUserService_Register(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockUserRepository(ctrl)
    logger := zap.NewNop()
    userService := services.NewUserService(mockRepo, logger)
    
    t.Run("注册成功", func(t *testing.T) {
        params := request.Register{
            Name:     "test",
            Mobile:   "13800138000",
            Password: "123456",
        }
        
        mockRepo.EXPECT().
            FindByMobile(params.Mobile).
            Return(nil, nil)
        
        mockRepo.EXPECT().
            Create(gomock.Any()).
            Return(nil)
        
        user, err := userService.Register(params)
        
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, params.Name, user.Name)
    })
    
    t.Run("手机号已存在", func(t *testing.T) {
        params := request.Register{
            Mobile: "13800138000",
        }
        
        mockRepo.EXPECT().
            FindByMobile(params.Mobile).
            Return(&models.User{}, nil)
        
        _, err := userService.Register(params)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "已存在")
    })
}
```

---

### 6. ✅ Makefile 构建脚本

- [ ] **任务完成**

**新建文件**: `Makefile`
```makefile
.PHONY: build run test clean docker swagger lint help

# 变量定义
APP_NAME := gin-web
BUILD_DIR := ./build
GO := go
GOFLAGS := -ldflags="-s -w"

# 默认目标
all: lint test build

# 帮助信息
help:
	@echo "Usage:"
	@echo "  make build     - 编译应用"
	@echo "  make run       - 运行应用"
	@echo "  make test      - 运行测试"
	@echo "  make lint      - 代码检查"
	@echo "  make swagger   - 生成 Swagger 文档"
	@echo "  make docker    - 构建 Docker 镜像"
	@echo "  make clean     - 清理构建产物"

# 编译
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run:
	$(GO) run main.go

# 热重载运行 (需要安装 air: go install github.com/cosmtrek/air@latest)
dev:
	air

# 测试
test:
	$(GO) test -v -cover ./...

# 测试覆盖率
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 代码检查
lint:
	golangci-lint run ./...

# 格式化
fmt:
	gofmt -s -w .
	goimports -w .

# 生成 Swagger 文档
swagger:
	swag init
	@echo "Swagger docs generated"

# 生成 Wire 依赖注入代码
wire:
	cd internal/container && wire
	@echo "Wire generated"

# 构建 Docker 镜像
docker:
	docker build -t $(APP_NAME):latest .

# Docker Compose 启动
docker-up:
	docker-compose up -d

# Docker Compose 停止
docker-down:
	docker-compose down

# 清理
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

# 数据库迁移
migrate:
	$(GO) run main.go migrate

# 安装开发工具
tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"
```

---

## 完成检查清单

- [ ] Docker 多阶段构建已实现
- [ ] docker-compose.yml 已创建
- [ ] 请求日志中间件已添加
- [ ] 限流中间件已添加
- [ ] 超时中间件已添加
- [ ] 链路追踪中间件已添加
- [ ] WebSocket Hub 已实现
- [ ] WebSocket Client 已实现
- [ ] Swagger 文档已集成
- [ ] 单元测试框架已搭建
- [ ] Makefile 已创建
- [ ] 所有新功能已测试
- [ ] 文档已更新

---

## 依赖更新

需要添加到 `go.mod` 的依赖：

```bash
# WebSocket
go get github.com/gorilla/websocket

# 限流
go get golang.org/x/time/rate

# UUID
go get github.com/google/uuid

# Swagger
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
go get github.com/swaggo/swag

# 测试
go get github.com/stretchr/testify
go get github.com/golang/mock/gomock
```

---

## 参考资源

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Swaggo](https://github.com/swaggo/swag)
- [Go Mock](https://github.com/golang/mock)
- [Docker 多阶段构建](https://docs.docker.com/develop/develop-images/multistage-build/)