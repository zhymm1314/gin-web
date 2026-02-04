# ✅ P2 - 代码规范 (Code Style) - 已完成

> 优先级：一般
> 完成日期：2026-02-04
> 影响范围：代码可读性 & 一致性

---

## 概述

这些优化将提升代码的规范性和一致性，使项目更易于维护和协作。

**状态**: ✅ 全部完成

---

## TODO 列表

### 1. ✅ 统一返回值顺序

- [x] **任务完成**

**背景**:
Go 语言惯例是将 error 放在返回值最后，当前项目中存在不一致的情况。

**当前代码**:
```go
// app/services/user.go
func (userService *userService) Register(params request.Register) (err error, user models.User)
func (userService *userService) Login(params request.Login) (err error, user *models.User)
func (userService *userService) GetUserInfo(id string) (err error, user models.User)

// app/services/jwt.go  
func (jwtService *jwtService) GetUserInfo(GuardName string, id string) (err error, user JwtUser)
```

**目标代码**:
```go
// 统一为：结果在前，error 在后
func (s *UserService) Register(params request.Register) (*models.User, error)
func (s *UserService) Login(params request.Login) (*models.User, error)
func (s *UserService) GetUserInfo(id string) (*models.User, error)

func (s *JwtService) GetUserInfo(guardName string, id string) (JwtUser, error)
```

**修改文件清单**:

| 文件 | 函数 | 修改内容 |
|------|------|----------|
| `app/services/user.go` | Register | 返回值顺序 |
| `app/services/user.go` | Login | 返回值顺序 |
| `app/services/user.go` | GetUserInfo | 返回值顺序 |
| `app/services/jwt.go` | GetUserInfo | 返回值顺序 |
| `app/services/jwt.go` | CreateToken | 检查返回值顺序 |
| `app/controllers/user.go` | 所有调用处 | 适配新的返回值顺序 |
| `app/controllers/auth.go` | 所有调用处 | 适配新的返回值顺序 |

**修改示例**:

```go
// Before
if err, user := services.UserService.Register(form); err != nil {
    response.BusinessFail(c, err.Error())
} else {
    response.Success(c, user)
}

// After
user, err := services.UserService.Register(form)
if err != nil {
    response.BusinessFail(c, err.Error())
    return
}
response.Success(c, user)
```

---

### 2. ✅ 修复目录拼写错误

- [x] **任务完成**

**问题**:
```
❌ app/ampq/   (拼写错误)
✅ app/amqp/   (正确拼写: Advanced Message Queuing Protocol)
```

**修改步骤**:

```bash
# Step 1: 重命名目录
mv app/ampq app/amqp

# Step 2: 更新所有 import 语句
# 搜索所有包含 "gin-web/app/ampq" 的文件并替换为 "gin-web/app/amqp"
```

**需要更新的文件**:
- `bootstrap/rabbitmq.go`
- `app/amqp/consumer/abstract.go`
- `app/amqp/consumer/log_consumer.go`
- `app/amqp/producer/abstract.go`
- `app/amqp/producer/log_producer.go`

---

### 3. ✅ 规范 Model 定义

- [x] **任务完成**

**当前问题**:

```go
// app/models/user.go
type User struct {
    ID
    Name     string `json:"name1" gorm:"not null;comment:用户名称"`   // ❌ json tag 命名奇怪
    Mobile   string `json:"mobile2" gorm:"not null;index;comment:用户手机号"` // ❌ json tag 命名奇怪
    Password string `json:"-" gorm:"not null;default:'';comment:用户密码"`
    Timestamps
    SoftDeletes
}
```

**优化后**:

```go
// app/models/user.go
package models

import "strconv"

// User 用户模型
type User struct {
    ID
    Name     string `json:"name" gorm:"type:varchar(100);not null;comment:用户名称"`
    Mobile   string `json:"mobile" gorm:"type:varchar(20);not null;uniqueIndex;comment:用户手机号"`
    Password string `json:"-" gorm:"type:varchar(255);not null;comment:用户密码"`
    Timestamps
    SoftDeletes
}

// TableName 指定表名
func (User) TableName() string {
    return "users"
}

// GetUid 获取用户ID字符串
func (u User) GetUid() string {
    return strconv.FormatUint(uint64(u.ID.ID), 10)
}

// MaskMobile 获取脱敏手机号
func (u User) MaskMobile() string {
    if len(u.Mobile) < 11 {
        return u.Mobile
    }
    return u.Mobile[:3] + "****" + u.Mobile[7:]
}
```

**建议增加 common.go 的完善**:

```go
// app/models/common.go
package models

import (
    "gorm.io/gorm"
    "time"
)

// ID 主键
type ID struct {
    ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
}

// Timestamps 时间戳
type Timestamps struct {
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SoftDeletes 软删除
type SoftDeletes struct {
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BaseModel 基础模型（组合使用）
type BaseModel struct {
    ID
    Timestamps
    SoftDeletes
}
```

---

### 4. ✅ 类型安全的配置

- [x] **任务完成**

**当前问题**:

```go
// config/config.go
type Configuration struct {
    // ...
    ApiUrls  map[string]any `yaml:"api_url"`  // ❌ any 类型不安全
}
```

**优化方案**:

```go
// config/api_url.go
package config

// ApiUrls API 地址配置
type ApiUrls struct {
    UserService   ServiceEndpoint `yaml:"user_service" mapstructure:"user_service"`
    OrderService  ServiceEndpoint `yaml:"order_service" mapstructure:"order_service"`
    PayService    ServiceEndpoint `yaml:"pay_service" mapstructure:"pay_service"`
}

// ServiceEndpoint 服务端点配置
type ServiceEndpoint struct {
    Local      string `yaml:"local" mapstructure:"local"`
    Dev        string `yaml:"dev" mapstructure:"dev"`
    Test       string `yaml:"test" mapstructure:"test"`
    Production string `yaml:"production" mapstructure:"production"`
}

// GetURL 根据环境获取 URL
func (s ServiceEndpoint) GetURL(env string) string {
    switch env {
    case "local":
        return s.Local
    case "dev":
        return s.Dev
    case "test":
        return s.Test
    case "production":
        return s.Production
    default:
        return s.Local
    }
}
```

**更新 config.go**:

```go
// config/config.go
type Configuration struct {
    App      App      `mapstructure:"app" yaml:"app"`
    Log      Log      `mapstructure:"log" yaml:"log"`
    Database Database `mapstructure:"database" yaml:"database"`
    Jwt      Jwt      `mapstructure:"jwt" yaml:"jwt"`
    Redis    Redis    `mapstructure:"redis" yaml:"redis"`
    RabbitMQ RabbitMQ `mapstructure:"rabbitmq" yaml:"rabbitmq"`
    ApiUrls  ApiUrls  `mapstructure:"api_url" yaml:"api_url"`  // ✅ 类型安全
}
```

**更新 api/abstract.go**:

```go
// Before
func GetApiUrl(serviceName string) string {
    env := global.App.Config.App.Env
    urlMap := global.App.Config.ApiUrls
    rawInnerMap, ok := urlMap[serviceName].(map[string]any)
    if !ok {
        panic("配置格式错误：非 map[string]string 类型")
    }
    url := rawInnerMap[env].(string)
    // ...
}

// After
func GetApiUrl(serviceName string) string {
    env := global.App.Config.App.Env
    apiUrls := global.App.Config.ApiUrls
    
    switch serviceName {
    case "user_service":
        return apiUrls.UserService.GetURL(env)
    case "order_service":
        return apiUrls.OrderService.GetURL(env)
    case "pay_service":
        return apiUrls.PayService.GetURL(env)
    default:
        panic("未配置的服务: " + serviceName)
    }
}
```

---

### 5. ✅ 代码注释完善

- [x] **任务完成**

**规范要求**:

1. **包注释**: 每个包必须有包注释
2. **导出函数/类型注释**: 所有导出的函数、类型必须有注释
3. **注释格式**: 以被注释的名称开头

**示例**:

```go
// Package services 提供业务逻辑服务
package services

import (
    // ...
)

// UserService 用户服务
// 提供用户注册、登录、信息获取等功能
type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

// NewUserService 创建用户服务实例
// 
// 参数:
//   - repo: 用户仓储接口
//   - log: 日志记录器
//
// 返回:
//   - *UserService: 用户服务实例
func NewUserService(repo repository.UserRepository, log *zap.Logger) *UserService {
    return &UserService{repo: repo, log: log}
}

// Register 用户注册
//
// 参数:
//   - params: 注册请求参数
//
// 返回:
//   - *models.User: 创建的用户对象
//   - error: 错误信息，手机号已存在时返回错误
func (s *UserService) Register(params request.Register) (*models.User, error) {
    // ...
}
```

**需要添加注释的文件清单**:

| 文件 | 优先级 | 说明 |
|------|--------|------|
| `app/services/*.go` | 高 | 核心业务逻辑 |
| `app/controllers/*.go` | 高 | API 接口 |
| `app/middleware/*.go` | 中 | 中间件 |
| `bootstrap/*.go` | 中 | 初始化逻辑 |
| `utils/*.go` | 低 | 工具函数 |
| `config/*.go` | 低 | 配置定义 |

---

### 6. ✅ 变量命名规范

- [x] **任务完成**

**命名规则**:

| 类型 | 规则 | 示例 |
|------|------|------|
| 包名 | 小写，简短 | `services`, `models` |
| 接口 | 以 -er 结尾 | `Reader`, `UserRepository` |
| 结构体 | 驼峰命名 | `UserService`, `CustomClaims` |
| 函数 | 驼峰命名 | `GetUserInfo`, `CreateToken` |
| 常量 | 驼峰命名 | `MaxRetryCount`, `DefaultTimeout` |
| 变量 | 驼峰命名 | `userService`, `tokenStr` |
| 参数 | 驼峰命名 | `guardName` (不是 `GuardName`) |

**需要修改的命名问题**:

```go
// app/middleware/jwt.go
func JWTAuth(GuardName string) gin.HandlerFunc  // ❌ 参数不应大写开头

// 应改为
func JWTAuth(guardName string) gin.HandlerFunc  // ✅
```

---

### 7. ✅ 统一使用指针还是值类型

- [x] **任务完成**

**规则**:
- **大型结构体**: 使用指针，避免拷贝开销
- **需要修改**: 使用指针
- **小型不可变结构体**: 可以使用值类型

**当前不一致**:
```go
// 有时返回值类型
func (s *userService) GetUserInfo(id string) (err error, user models.User)

// 有时返回指针
func (s *userService) Login(params request.Login) (err error, user *models.User)
```

**统一为指针**:
```go
func (s *UserService) Register(params request.Register) (*models.User, error)
func (s *UserService) Login(params request.Login) (*models.User, error)
func (s *UserService) GetUserInfo(id uint) (*models.User, error)
```

---

## 完成检查清单

- [x] 所有返回值顺序已统一 (error 在最后)
- [x] 目录拼写错误已修复 (ampq → amqp)
- [x] Model JSON tag 已规范化
- [x] 配置类型已改为类型安全
- [x] 核心文件注释已完善
- [x] 变量命名已统一
- [x] 值类型/指针使用已统一
- [x] 运行 `go vet` 无警告
- [ ] 运行 `golangci-lint` 无错误（需安装 golangci-lint）

---

## 工具配置

### golangci-lint 配置

**新建文件**: `.golangci.yml`
```yaml
run:
  timeout: 5m
  go: '1.22'

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - misspell
    - revive
    - exportloopref

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: gin-web
  revive:
    rules:
      - name: exported
        severity: warning
      - name: var-naming
        severity: warning

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

### EditorConfig 配置

**新建文件**: `.editorconfig`
```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 4

[*.go]
indent_style = tab

[*.{yaml,yml}]
indent_size = 2

[*.md]
trim_trailing_whitespace = false
```

---

## 参考资源

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [golangci-lint](https://golangci-lint.run/)