# 中间件使用指南

本文档详细说明如何在 Gin-Web 项目中使用和开发中间件。

---

## 目录

- [概述](#概述)
- [内置中间件](#内置中间件)
  - [JWT 认证中间件](#jwt-认证中间件)
  - [CORS 跨域中间件](#cors-跨域中间件)
  - [Recovery 恢复中间件](#recovery-恢复中间件)
- [中间件使用方式](#中间件使用方式)
  - [全局中间件](#全局中间件)
  - [路由组中间件](#路由组中间件)
  - [单路由中间件](#单路由中间件)
  - [控制器中间件](#控制器中间件)
- [开发自定义中间件](#开发自定义中间件)
  - [基本结构](#基本结构)
  - [完整示例](#完整示例)
- [常用中间件模板](#常用中间件模板)
  - [请求日志中间件](#请求日志中间件)
  - [请求限流中间件](#请求限流中间件)
  - [权限验证中间件](#权限验证中间件)
  - [请求签名验证中间件](#请求签名验证中间件)
- [中间件执行顺序](#中间件执行顺序)
- [最佳实践](#最佳实践)
- [注意事项](#注意事项)

---

## 概述

中间件是在请求处理前后执行的函数，常用于：

- **认证授权**：JWT 验证、权限检查
- **日志记录**：请求日志、审计日志
- **安全防护**：CORS、限流、签名验证
- **错误处理**：panic 恢复、统一错误处理
- **数据处理**：请求解密、响应压缩

**Gin 中间件签名**:

```go
type HandlerFunc func(*gin.Context)
```

---

## 内置中间件

项目内置了三个核心中间件，位于 `app/middleware/` 目录：

### JWT 认证中间件

**文件位置**: `app/middleware/jwt.go`

**功能**:
- Token 解析与验证
- Token 黑名单检查
- Token 自动续签
- 用户信息注入 Context

**使用方式**:

```go
import (
    "gin-web/app/middleware"
    "gin-web/app/services"
)

// 在路由组中使用
authRouter := router.Group("/api").Use(middleware.JWTAuth(services.AppGuardName))
{
    authRouter.GET("/user/info", controllers.UserInfo)
}

// 在单个路由中使用
router.POST("/api/order", middleware.JWTAuth(services.AppGuardName), controllers.CreateOrder)
```

**获取用户信息**:

```go
func UserInfo(c *gin.Context) {
    // 获取用户 ID
    userID := c.GetString("id")

    // 获取完整 Token 对象
    token := c.MustGet("token").(*jwt.Token)
    claims := token.Claims.(*services.CustomClaims)

    response.Success(c, gin.H{
        "user_id": userID,
        "issuer":  claims.Issuer,
    })
}
```

### CORS 跨域中间件

**文件位置**: `app/middleware/cors.go`

**功能**:
- 允许跨域请求
- 配置允许的请求头
- 暴露自定义响应头（如 Token 续签）

**使用方式**:

```go
// 全局使用（已在 bootstrap/router.go 中配置）
router.Use(middleware.Cors())
```

**自定义配置**:

```go
func Cors() gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     []string{"https://example.com", "https://api.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
        ExposeHeaders:    []string{"New-Token", "New-Expires-In"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    return cors.New(config)
}
```

### Recovery 恢复中间件

**文件位置**: `app/middleware/recovery.go`

**功能**:
- 捕获 panic 防止服务崩溃
- 记录错误日志到文件
- 返回统一错误响应

**使用方式**:

```go
// 全局使用（已在 bootstrap/router.go 中配置）
router.Use(middleware.CustomRecovery())
```

---

## 中间件使用方式

### 全局中间件

在 `bootstrap/router.go` 中注册，对所有请求生效：

```go
func setupRouter() *gin.Engine {
    router := gin.New()

    // 全局中间件
    router.Use(gin.Logger())              // 日志
    router.Use(middleware.CustomRecovery()) // 错误恢复
    router.Use(middleware.Cors())          // 跨域

    return router
}
```

### 路由组中间件

对特定路由组生效：

```go
// 方式一：创建时指定
authRouter := router.Group("/api/admin").Use(middleware.JWTAuth(services.AppGuardName))
{
    authRouter.GET("/users", controllers.ListUsers)
    authRouter.POST("/users", controllers.CreateUser)
}

// 方式二：链式调用
adminRouter := router.Group("/api/admin")
adminRouter.Use(middleware.JWTAuth(services.AppGuardName))
adminRouter.Use(middleware.AdminOnly())  // 自定义中间件
{
    adminRouter.GET("/dashboard", controllers.Dashboard)
}
```

### 单路由中间件

只对单个路由生效：

```go
// 单个中间件
router.POST("/api/upload", middleware.JWTAuth(services.AppGuardName), controllers.Upload)

// 多个中间件
router.POST("/api/admin/config",
    middleware.JWTAuth(services.AppGuardName),
    middleware.AdminOnly(),
    middleware.RateLimit(10), // 每秒10次
    controllers.UpdateConfig,
)
```

### 控制器中间件

使用依赖注入模式时，可在 Controller 中定义中间件：

```go
// app/controllers/admin_controller.go

type AdminController struct {
    adminService *services.AdminService
}

func (ctrl *AdminController) Prefix() string {
    return "/admin"
}

func (ctrl *AdminController) Routes() []Route {
    return []Route{
        // 公开接口
        {Method: "GET", Path: "/health", Handler: ctrl.Health},

        // 需要认证的接口
        {
            Method:      "GET",
            Path:        "/users",
            Handler:     ctrl.ListUsers,
            Middlewares: []gin.HandlerFunc{middleware.JWTAuth(services.AppGuardName)},
        },

        // 需要认证 + 管理员权限
        {
            Method:  "DELETE",
            Path:    "/users/:id",
            Handler: ctrl.DeleteUser,
            Middlewares: []gin.HandlerFunc{
                middleware.JWTAuth(services.AppGuardName),
                middleware.AdminOnly(),
            },
        },
    }
}
```

---

## 开发自定义中间件

### 基本结构

```go
package middleware

import "github.com/gin-gonic/gin"

// MyMiddleware 自定义中间件
func MyMiddleware() gin.HandlerFunc {
    // 这里可以做初始化工作（只执行一次）

    return func(c *gin.Context) {
        // ========== 请求前处理 ==========
        // 在这里处理请求进入时的逻辑

        // 调用下一个处理器
        c.Next()

        // ========== 请求后处理 ==========
        // 在这里处理响应返回前的逻辑
    }
}
```

### 完整示例

**带参数的中间件**:

```go
// app/middleware/timeout.go
package middleware

import (
    "context"
    "gin-web/app/common/response"
    "github.com/gin-gonic/gin"
    "time"
)

// Timeout 请求超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 创建带超时的 context
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()

        // 替换请求的 context
        c.Request = c.Request.WithContext(ctx)

        // 使用 channel 等待处理完成
        done := make(chan struct{})

        go func() {
            c.Next()
            close(done)
        }()

        select {
        case <-done:
            // 正常完成
        case <-ctx.Done():
            // 超时
            c.Abort()
            response.Fail(c, 408, "请求超时")
        }
    }
}
```

**使用**:

```go
router.POST("/api/slow-task", middleware.Timeout(30*time.Second), controllers.SlowTask)
```

---

## 常用中间件模板

### 请求日志中间件

```go
// app/middleware/request_log.go
package middleware

import (
    "bytes"
    "gin-web/global"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "io"
    "time"
)

// RequestLog 请求日志中间件
func RequestLog() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }
        c.Set("request_id", requestID)

        // 读取请求体（需要时）
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }

        // 处理请求
        c.Next()

        // 记录日志
        duration := time.Since(startTime)
        global.App.Log.Info("HTTP Request",
            zap.String("request_id", requestID),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.String("query", c.Request.URL.RawQuery),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", duration),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        )
    }
}

func generateRequestID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

### 请求限流中间件

```go
// app/middleware/rate_limit.go
package middleware

import (
    "gin-web/app/common/response"
    "gin-web/global"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    "sync"
)

var (
    limiters = make(map[string]*rate.Limiter)
    mu       sync.RWMutex
)

// RateLimit 限流中间件
// rps: 每秒允许的请求数
func RateLimit(rps int) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.ClientIP() // 按 IP 限流

        limiter := getLimiter(key, rps)
        if !limiter.Allow() {
            response.Fail(c, 429, "请求过于频繁，请稍后再试")
            c.Abort()
            return
        }

        c.Next()
    }
}

// RateLimitByUser 按用户限流
func RateLimitByUser(rps int) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("id")
        if userID == "" {
            c.Next()
            return
        }

        key := "user:" + userID
        limiter := getLimiter(key, rps)
        if !limiter.Allow() {
            response.Fail(c, 429, "请求过于频繁，请稍后再试")
            c.Abort()
            return
        }

        c.Next()
    }
}

func getLimiter(key string, rps int) *rate.Limiter {
    mu.RLock()
    limiter, exists := limiters[key]
    mu.RUnlock()

    if exists {
        return limiter
    }

    mu.Lock()
    defer mu.Unlock()

    // 双重检查
    if limiter, exists = limiters[key]; exists {
        return limiter
    }

    limiter = rate.NewLimiter(rate.Limit(rps), rps*2) // 允许突发
    limiters[key] = limiter
    return limiter
}
```

### 权限验证中间件

```go
// app/middleware/permission.go
package middleware

import (
    "gin-web/app/common/response"
    "gin-web/app/services"
    "github.com/gin-gonic/gin"
)

// AdminOnly 仅管理员可访问
func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("id")
        if userID == "" {
            response.TokenFail(c)
            c.Abort()
            return
        }

        // 检查用户角色
        user, err := services.UserService.GetByID(userID)
        if err != nil || user.Role != "admin" {
            response.Fail(c, 403, "无权限访问")
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequirePermission 需要特定权限
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("id")
        if userID == "" {
            response.TokenFail(c)
            c.Abort()
            return
        }

        // 检查用户权限
        hasPermission := services.PermissionService.Check(userID, permission)
        if !hasPermission {
            response.Fail(c, 403, "缺少权限: "+permission)
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireRoles 需要特定角色（任一）
func RequireRoles(roles ...string) gin.HandlerFunc {
    roleSet := make(map[string]bool)
    for _, role := range roles {
        roleSet[role] = true
    }

    return func(c *gin.Context) {
        userID := c.GetString("id")
        if userID == "" {
            response.TokenFail(c)
            c.Abort()
            return
        }

        user, err := services.UserService.GetByID(userID)
        if err != nil {
            response.Fail(c, 403, "获取用户信息失败")
            c.Abort()
            return
        }

        if !roleSet[user.Role] {
            response.Fail(c, 403, "角色权限不足")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 请求签名验证中间件

```go
// app/middleware/signature.go
package middleware

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "gin-web/app/common/response"
    "gin-web/global"
    "github.com/gin-gonic/gin"
    "io"
    "strconv"
    "time"
)

// SignatureVerify 签名验证中间件
func SignatureVerify(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取签名相关头部
        timestamp := c.GetHeader("X-Timestamp")
        signature := c.GetHeader("X-Signature")
        nonce := c.GetHeader("X-Nonce")

        if timestamp == "" || signature == "" || nonce == "" {
            response.Fail(c, 401, "缺少签名参数")
            c.Abort()
            return
        }

        // 验证时间戳（防止重放攻击，5分钟内有效）
        ts, err := strconv.ParseInt(timestamp, 10, 64)
        if err != nil || time.Now().Unix()-ts > 300 {
            response.Fail(c, 401, "请求已过期")
            c.Abort()
            return
        }

        // 验证 nonce 是否已使用（使用 Redis）
        nonceKey := "nonce:" + nonce
        exists, _ := global.App.Redis.Exists(c, nonceKey).Result()
        if exists > 0 {
            response.Fail(c, 401, "请求重复")
            c.Abort()
            return
        }

        // 读取请求体
        body, _ := io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

        // 计算签名: HMAC-SHA256(timestamp + nonce + body)
        data := timestamp + nonce + string(body)
        expectedSig := computeHMAC(data, secret)

        if signature != expectedSig {
            response.Fail(c, 401, "签名验证失败")
            c.Abort()
            return
        }

        // 记录 nonce（5分钟过期）
        global.App.Redis.Set(c, nonceKey, "1", 5*time.Minute)

        c.Next()
    }
}

func computeHMAC(data, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
```

---

## 中间件执行顺序

```
请求进入
    │
    ▼
┌─────────────────────┐
│ 全局中间件1 (Before) │
├─────────────────────┤
│ 全局中间件2 (Before) │
├─────────────────────┤
│ 路由组中间件 (Before)│
├─────────────────────┤
│ 单路由中间件 (Before)│
├─────────────────────┤
│     Controller      │ ← 业务处理
├─────────────────────┤
│ 单路由中间件 (After) │
├─────────────────────┤
│ 路由组中间件 (After) │
├─────────────────────┤
│ 全局中间件2 (After)  │
├─────────────────────┤
│ 全局中间件1 (After)  │
└─────────────────────┘
    │
    ▼
响应返回
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `c.Next()` | 调用后续处理器，然后返回执行 After 逻辑 |
| `c.Abort()` | 中止后续处理器，但会执行当前中间件的 After 逻辑 |
| `c.AbortWithStatus(code)` | 中止并设置状态码 |
| `c.AbortWithStatusJSON(code, obj)` | 中止并返回 JSON |

---

## 最佳实践

### 1. 中间件职责单一

```go
// 好：每个中间件只做一件事
router.Use(middleware.RequestLog())
router.Use(middleware.RateLimit(100))
router.Use(middleware.JWTAuth(services.AppGuardName))

// 差：一个中间件做太多事情
router.Use(middleware.DoEverything()) // 不推荐
```

### 2. 合理使用 Context 传递数据

```go
// 设置数据
c.Set("user_id", userID)
c.Set("request_start_time", time.Now())

// 获取数据
userID := c.GetString("user_id")
startTime := c.MustGet("request_start_time").(time.Time)
```

### 3. 使用 defer 确保资源释放

```go
func DatabaseTransaction() gin.HandlerFunc {
    return func(c *gin.Context) {
        tx := global.App.DB.Begin()
        c.Set("tx", tx)

        defer func() {
            if r := recover(); r != nil {
                tx.Rollback()
                panic(r)
            }
        }()

        c.Next()

        if c.Writer.Status() >= 400 {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }
}
```

### 4. 错误处理统一

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 处理所有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            global.App.Log.Error("Request error",
                zap.Error(err),
                zap.String("path", c.Request.URL.Path),
            )
            // 错误已在业务中处理，这里只记录日志
        }
    }
}
```

---

## 注意事项

### 必须遵守

1. **使用 `c.Abort()` 终止请求**：不调用 `Abort()` 会继续执行后续处理器
2. **小心 `c.Next()` 的位置**：它决定了 Before 和 After 逻辑的分界
3. **避免阻塞操作**：中间件中的耗时操作会影响所有请求

### 建议遵守

1. **中间件顺序很重要**：Recovery 放最前面，日志放前面，认证放后面
2. **使用工厂函数**：返回 `gin.HandlerFunc`，支持参数配置
3. **记录关键日志**：特别是认证失败、权限拒绝等安全事件

### 避免

1. **避免在中间件中修改请求体后不重置**
2. **避免在 After 逻辑中再次写入响应**（响应已发送）
3. **避免中间件之间产生循环依赖**
