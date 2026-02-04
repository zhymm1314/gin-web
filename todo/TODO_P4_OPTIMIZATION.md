# P4 - 工程化优化 (Engineering Optimization)

> 优先级：中
> 预计工时：2-3 天
> 影响范围：开发效率 & 部署运维

---

## 概述

P4 阶段专注于工程化优化，包括 Docker 容器化、中间件增强、单元测试和构建脚本等，提升开发效率和生产部署能力。

---

## TODO 列表

### 1. Docker 多阶段构建

- [ ] **任务完成**

**背景**:
当前 Dockerfile 使用完整的 golang 镜像，构建出的镜像约 800MB，生产环境部署效率低。

**优化后 Dockerfile**:

```dockerfile
# ==================== 构建阶段 ====================
FROM golang:1.22-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o main .

# ==================== 运行阶段 ====================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Shanghai

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /build/main .
COPY --from=builder /build/config ./config

RUN mkdir -p storage/logs && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/ping || exit 1

CMD ["./main"]
```

**新增 .dockerignore**:

```
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

### 2. 添加更多中间件

- [ ] **任务完成**

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

func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)

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

type IPRateLimiter struct {
    ips   map[string]*rate.Limiter
    mu    *sync.RWMutex
    rate  rate.Limit
    burst int
}

func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
    return &IPRateLimiter{
        ips:   make(map[string]*rate.Limiter),
        mu:    &sync.RWMutex{},
        rate:  r,
        burst: burst,
    }
}

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

func Trace() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.GetHeader(TraceIDHeader)
        if traceID == "" {
            traceID = uuid.New().String()
        }

        c.Set(TraceIDKey, traceID)
        c.Header(TraceIDHeader, traceID)

        c.Next()
    }
}

func GetTraceID(c *gin.Context) string {
    if traceID, exists := c.Get(TraceIDKey); exists {
        return traceID.(string)
    }
    return ""
}
```

---

### 3. 添加单元测试框架

- [ ] **任务完成**

#### 3.1 测试目录结构

```
gin-web/
├── tests/
│   ├── setup_test.go       # 测试初始化
│   ├── user_test.go        # 用户测试
│   └── mocks/
│       └── user_repository_mock.go
```

#### 3.2 测试初始化

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
    os.Setenv("GIN_MODE", "test")

    bootstrap.InitializeConfig()
    global.App.Log = bootstrap.InitializeLog()

    code := m.Run()

    os.Exit(code)
}
```

#### 3.3 Mock 工具

```bash
go install github.com/golang/mock/mockgen@latest

# 生成 mock
mockgen -source=internal/repository/repository.go -destination=tests/mocks/user_repository_mock.go -package=mocks
```

#### 3.4 Service 测试示例

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

### 4. Makefile 构建脚本

- [ ] **任务完成**

**新建文件**: `Makefile`

```makefile
.PHONY: build run test clean docker swagger lint help

APP_NAME := gin-web
BUILD_DIR := ./build
GO := go
GOFLAGS := -ldflags="-s -w"

all: lint test build

help:
	@echo "Usage:"
	@echo "  make build     - 编译应用"
	@echo "  make run       - 运行应用"
	@echo "  make test      - 运行测试"
	@echo "  make lint      - 代码检查"
	@echo "  make swagger   - 生成 Swagger 文档"
	@echo "  make docker    - 构建 Docker 镜像"
	@echo "  make clean     - 清理构建产物"

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run:
	$(GO) run main.go

dev:
	air

test:
	$(GO) test -v -cover ./...

test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

swagger:
	swag init
	@echo "Swagger docs generated"

wire:
	cd internal/container && wire
	@echo "Wire generated"

docker:
	docker build -t $(APP_NAME):latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

migrate:
	$(GO) run main.go migrate

tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"
```

---

### 5. 热重载开发 (Air)

- [ ] **任务完成**

#### 5.1 安装 Air

```bash
go install github.com/cosmtrek/air@latest
```

#### 5.2 配置文件

**新建文件**: `.air.toml`

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ."
bin = "./tmp/main"
full_bin = "./tmp/main"
include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
exclude_dir = ["assets", "tmp", "vendor", "storage", "docs", "todo"]
include_dir = []
exclude_file = []
delay = 1000
stop_on_error = true
log = "air.log"

[log]
time = true

[color]
main = "yellow"
watcher = "cyan"
build = "green"
runner = "magenta"

[misc]
clean_on_exit = true
```

#### 5.3 使用

```bash
# 启动热重载开发
air

# 或使用 Makefile
make dev
```

---

## 完成检查清单

### Docker
- [ ] Dockerfile 多阶段构建
- [ ] .dockerignore 文件
- [ ] docker-compose.yml
- [ ] 镜像构建测试
- [ ] 容器运行测试

### 中间件
- [ ] 请求日志中间件
- [ ] 限流中间件
- [ ] 超时中间件
- [ ] 链路追踪中间件
- [ ] 中间件集成测试

### 单元测试
- [ ] 测试目录结构
- [ ] Mock 生成
- [ ] Service 测试示例
- [ ] 测试覆盖率报告

### 构建脚本
- [ ] Makefile 创建
- [ ] Air 热重载配置
- [ ] 常用命令测试

---

## 依赖更新

```bash
# 限流
go get golang.org/x/time/rate

# UUID
go get github.com/google/uuid

# 测试
go get github.com/stretchr/testify
go get github.com/golang/mock/gomock

# 热重载
go install github.com/cosmtrek/air@latest
```
