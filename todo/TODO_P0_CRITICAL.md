# 🔴 P0 - 紧急修复 (Critical)

> 优先级：最高
> 预计工时：2-4 小时
> 影响范围：系统稳定性 & 安全性

---

## 概述

这些问题必须立即修复，它们直接影响系统的正常运行和安全性。

---

## TODO 列表

### 1. ✅ 修复 Router 重复启动 Bug

- [ ] **任务完成**

**文件位置**: `bootstrap/router.go`

**问题描述**:
`RunServer()` 函数中 `r.Run()` 会阻塞程序，导致后续的优雅关闭逻辑永远不会执行。

**当前代码**:
```go
func RunServer() {
    r := setupRouter()
    r.Run(":" + global.App.Config.App.Port)  // ❌ 阻塞在这里

    srv := &http.Server{
        Addr:    ":" + global.App.Config.App.Port,
        Handler: r,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // 等待中断信号以优雅地关闭服务器
    quit := make(chan os.Signal)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    // ... 优雅关闭逻辑
}
```

**修复方案**:
```go
func RunServer() {
    r := setupRouter()
    
    srv := &http.Server{
        Addr:    ":" + global.App.Config.App.Port,
        Handler: r,
    }

    // 在 goroutine 中启动服务器
    go func() {
        global.App.Log.Info("Server starting on port " + global.App.Config.App.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
    quit := make(chan os.Signal, 1)  // 注意：添加 buffer size
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutdown Server ...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server Shutdown:", err)
    }
    log.Println("Server exiting")
}
```

**验证方法**:
1. 启动服务器
2. 发送 SIGTERM 信号 (`kill -15 <pid>`)
3. 确认服务器能够优雅关闭

---

### 2. ✅ 升级 JWT 库

- [ ] **任务完成**

**问题描述**:
`github.com/dgrijalva/jwt-go` 已被废弃，存在安全漏洞 CVE-2020-26160。

**当前依赖**:
```go
// go.mod
github.com/dgrijalva/jwt-go v3.2.0+incompatible
```

**修复步骤**:

#### Step 1: 更新 go.mod
```bash
go get github.com/golang-jwt/jwt/v5
go mod tidy
```

#### Step 2: 更新 import 语句

**文件**: `app/middleware/jwt.go`
```go
// Before
import "github.com/dgrijalva/jwt-go"

// After
import "github.com/golang-jwt/jwt/v5"
```

**文件**: `app/services/jwt.go`
```go
// Before
import "github.com/dgrijalva/jwt-go"

// After
import "github.com/golang-jwt/jwt/v5"
```

#### Step 3: 更新代码适配新 API

```go
// CustomClaims 结构体更新
type CustomClaims struct {
    jwt.RegisteredClaims  // v5 使用 RegisteredClaims 替代 StandardClaims
    // 自定义字段...
}

// Token 解析更新
token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
    return []byte(global.App.Config.Jwt.Secret), nil
})

// 时间处理更新 (v5 使用 *jwt.NumericDate)
claims := &CustomClaims{
    RegisteredClaims: jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        Issuer:    "your-app",
    },
}
```

**验证方法**:
1. 运行所有 JWT 相关测试
2. 测试登录、Token 验证、Token 续签功能

---

### 3. ✅ 统一 RabbitMQ 库

- [ ] **任务完成**

**问题描述**:
项目中同时使用了两个 RabbitMQ 库，且 `streadway/amqp` 已被废弃。

**当前状态**:
```go
// go.mod
github.com/rabbitmq/amqp091-go  // producer 使用
github.com/streadway/amqp       // consumer 使用 (已废弃)
```

**修复步骤**:

#### Step 1: 移除废弃库
```bash
go mod edit -droprequire github.com/streadway/amqp
go mod tidy
```

#### Step 2: 更新 Consumer 代码

**文件**: `bootstrap/rabbitmq.go`
```go
// Before
import "github.com/streadway/amqp"

// After
import amqp "github.com/rabbitmq/amqp091-go"
```

**文件**: `app/ampq/consumer/abstract.go`
```go
// Before
import "github.com/streadway/amqp"

// After
import amqp "github.com/rabbitmq/amqp091-go"
```

**文件**: `app/ampq/consumer/log_consumer.go`
```go
// Before
import "github.com/streadway/amqp"

// After
import amqp "github.com/rabbitmq/amqp091-go"
```

#### Step 3: API 差异处理

`amqp091-go` 与 `streadway/amqp` 的 API 基本兼容，但需注意：

```go
// PublishWithContext 需要 context 参数
err := ch.PublishWithContext(
    context.Background(),  // 新增 context 参数
    "",                    // exchange
    queueName,             // routing key
    false,                 // mandatory
    false,                 // immediate
    amqp.Publishing{...},
)
```

**验证方法**:
1. 启动 RabbitMQ 服务
2. 测试消息发送功能
3. 测试消息消费功能
4. 检查日志确认无错误

---

## 完成检查清单

- [ ] Router 重复启动 Bug 已修复
- [ ] 优雅关闭功能正常工作
- [ ] JWT 库已升级到 v5
- [ ] JWT 登录/验证/续签功能正常
- [ ] RabbitMQ 库已统一
- [ ] 消息发送/消费功能正常
- [ ] 所有修改已通过代码审查
- [ ] 相关文档已更新

---

## 风险提示

⚠️ **JWT 升级注意事项**:
- 升级后生成的 Token 格式可能与旧版本不兼容
- 建议在升级时同时使旧 Token 失效，要求用户重新登录
- 或者设置过渡期，同时支持新旧 Token

⚠️ **RabbitMQ 升级注意事项**:
- 确保队列中没有积压的重要消息
- 建议在低峰期进行升级
- 保持消费者的幂等性

---

## 参考链接

- [golang-jwt/jwt 官方文档](https://github.com/golang-jwt/jwt)
- [jwt-go 迁移指南](https://github.com/golang-jwt/jwt/blob/main/MIGRATION_GUIDE.md)
- [amqp091-go 官方仓库](https://github.com/rabbitmq/amqp091-go)