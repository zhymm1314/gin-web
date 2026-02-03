# 🟠 P1 - 架构优化 (Architecture)

> 优先级：重要
> 预计工时：2-3 天
> 影响范围：代码可维护性 & 可测试性

---

## 概述

这些优化将显著提升项目的架构质量，使其更接近 Hyperf 的设计理念，便于 PHPer 无缝切换。

---

## TODO 列表

### 1. ✅ 实现依赖注入 (DI)

- [ ] **任务完成**

**背景**:
Hyperf 核心特性之一就是强大的 DI 容器，当前项目使用全局变量管理依赖，不利于测试和解耦。

**当前方式**:
```go
// 全局变量方式 - 紧耦合
var App = new(Application)
var UserService = new(userService)

func (s *userService) Register(params request.Register) (err error, user models.User) {
    global.App.DB.Where(...)  // 直接依赖全局变量
}
```

**目标方式**:
```go
// 依赖注入方式 - 松耦合
type UserService struct {
    db   *gorm.DB
    log  *zap.Logger
    repo repository.UserRepository
}

func NewUserService(db *gorm.DB, log *zap.Logger, repo repository.UserRepository) *UserService {
    return &UserService{db: db, log: log, repo: repo}
}
```

**实施步骤**:

#### Step 1: 安装 Wire (Google 的 DI 工具)
```bash
go install github.com/google/wire/cmd/wire@latest
go get github.com/google/wire
```

#### Step 2: 创建 Provider 定义

**新建文件**: `internal/container/providers.go`
```go
package container

import (
    "gin-web/app/controllers"
    "gin-web/app/services"
    "gin-web/internal/repository"
    "gin-web/bootstrap"
)

// ProvideDB 提供数据库连接
func ProvideDB() *gorm.DB {
    return bootstrap.InitializeDB()
}

// ProvideLogger 提供日志实例
func ProvideLogger() *zap.Logger {
    return bootstrap.InitializeLog()
}

// ProvideUserRepository 提供用户仓储
func ProvideUserRepository(db *gorm.DB) repository.UserRepository {
    return repository.NewUserRepository(db)
}

// ProvideUserService 提供用户服务
func ProvideUserService(repo repository.UserRepository, log *zap.Logger) *services.UserService {
    return services.NewUserService(repo, log)
}

// ProvideUserController 提供用户控制器
func ProvideUserController(svc *services.UserService) *controllers.UserController {
    return controllers.NewUserController(svc)
}
```

#### Step 3: 创建 Wire 配置

**新建文件**: `internal/container/wire.go`
```go
//go:build wireinject
// +build wireinject

package container

import (
    "gin-web/app/controllers"
    "github.com/google/wire"
)

func InitializeApp() (*App, error) {
    wire.Build(
        ProvideDB,
        ProvideLogger,
        ProvideRedis,
        ProvideUserRepository,
        ProvideUserService,
        ProvideUserController,
        NewApp,
    )
    return nil, nil
}
```

#### Step 4: 生成 Wire 代码
```bash
cd internal/container
wire
```

#### Step 5: 更新 main.go
```go
func main() {
    bootstrap.InitializeConfig()
    
    app, err := container.InitializeApp()
    if err != nil {
        log.Fatal(err)
    }
    defer app.Cleanup()
    
    app.Run()
}
```

**验证方法**:
1. 运行 `wire` 生成代码无错误
2. 应用正常启动
3. 所有 API 功能正常

---

### 2. ✅ 添加 Repository 层

- [ ] **任务完成**

**背景**:
当前 Service 直接操作数据库，违反单一职责原则。添加 Repository 层实现数据访问抽象。

**目录结构**:
```
internal/
└── repository/
    ├── repository.go      # 接口定义
    ├── user_repository.go # 用户仓储实现
    └── base_repository.go # 基础仓储
```

#### Step 1: 定义 Repository 接口

**新建文件**: `internal/repository/repository.go`
```go
package repository

import "gin-web/app/models"

// UserRepository 用户仓储接口
type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uint) (*models.User, error)
    FindByMobile(mobile string) (*models.User, error)
    Update(user *models.User) error
    Delete(id uint) error
}
```

#### Step 2: 实现 Repository

**新建文件**: `internal/repository/user_repository.go`
```go
package repository

import (
    "gin-web/app/models"
    "gorm.io/gorm"
)

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByMobile(mobile string) (*models.User, error) {
    var user models.User
    err := r.db.Where("mobile = ?", mobile).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}
```

#### Step 3: 重构 Service 使用 Repository

**更新文件**: `app/services/user.go`
```go
package services

import (
    "errors"
    "gin-web/app/common/request"
    "gin-web/app/models"
    "gin-web/internal/repository"
    "gin-web/utils"
    "go.uber.org/zap"
)

type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

func NewUserService(repo repository.UserRepository, log *zap.Logger) *UserService {
    return &UserService{repo: repo, log: log}
}

func (s *UserService) Register(params request.Register) (*models.User, error) {
    // 检查手机号是否存在
    existUser, _ := s.repo.FindByMobile(params.Mobile)
    if existUser != nil {
        return nil, errors.New("手机号已存在")
    }
    
    user := &models.User{
        Name:     params.Name,
        Mobile:   params.Mobile,
        Password: utils.BcryptMake([]byte(params.Password)),
    }
    
    if err := s.repo.Create(user); err != nil {
        s.log.Error("create user failed", zap.Error(err))
        return nil, err
    }
    
    return user, nil
}
```

**验证方法**:
1. 编写 Repository 单元测试
2. 使用 mock 测试 Service
3. 集成测试确保功能正常

---

### 3. ✅ 重构路由注册方式

- [ ] **任务完成**

**背景**:
向 Hyperf 的控制器注解路由靠拢，实现更优雅的路由注册。

**当前方式**:
```go
// routes/api.go - 手动注册每个路由
router.POST("/auth/register", app.Register)
router.POST("/auth/login", app.Login)
```

**目标方式**:
```go
// 控制器自动注册
type UserController struct {
    userService *services.UserService
}

func (c *UserController) Routes() []Route {
    return []Route{
        {Method: "POST", Path: "/register", Handler: c.Register},
        {Method: "POST", Path: "/login", Handler: c.Login},
    }
}
```

#### Step 1: 定义路由接口

**新建文件**: `app/controllers/controller.go`
```go
package controllers

import "github.com/gin-gonic/gin"

// Route 路由定义
type Route struct {
    Method      string
    Path        string
    Handler     gin.HandlerFunc
    Middlewares []gin.HandlerFunc
}

// Controller 控制器接口
type Controller interface {
    // Prefix 返回路由前缀
    Prefix() string
    // Routes 返回路由列表
    Routes() []Route
}

// RegisterController 注册控制器路由
func RegisterController(router *gin.RouterGroup, controller Controller) {
    group := router.Group(controller.Prefix())
    for _, route := range controller.Routes() {
        handlers := append(route.Middlewares, route.Handler)
        switch route.Method {
        case "GET":
            group.GET(route.Path, handlers...)
        case "POST":
            group.POST(route.Path, handlers...)
        case "PUT":
            group.PUT(route.Path, handlers...)
        case "DELETE":
            group.DELETE(route.Path, handlers...)
        }
    }
}
```

#### Step 2: 重构 UserController

**更新文件**: `app/controllers/user.go`
```go
package controllers

import (
    "gin-web/app/common/request"
    "gin-web/app/common/response"
    "gin-web/app/services"
    "github.com/gin-gonic/gin"
)

type UserController struct {
    userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
    return &UserController{userService: userService}
}

func (c *UserController) Prefix() string {
    return "/auth"
}

func (c *UserController) Routes() []Route {
    return []Route{
        {Method: "POST", Path: "/register", Handler: c.Register},
        {Method: "POST", Path: "/login", Handler: c.Login},
    }
}

func (c *UserController) Register(ctx *gin.Context) {
    var form request.Register
    if err := ctx.ShouldBindJSON(&form); err != nil {
        response.ValidateFail(ctx, request.GetErrorMsg(form, err))
        return
    }

    user, err := c.userService.Register(form)
    if err != nil {
        response.BusinessFail(ctx, err.Error())
        return
    }
    response.Success(ctx, user)
}
```

#### Step 3: 更新路由注册

**更新文件**: `routes/api.go`
```go
package routes

import (
    "gin-web/app/controllers"
    "github.com/gin-gonic/gin"
)

func SetApiGroupRoutes(router *gin.RouterGroup, ctrls ...controllers.Controller) {
    // 自动注册所有控制器
    for _, ctrl := range ctrls {
        controllers.RegisterController(router, ctrl)
    }
}
```

---

### 4. ✅ 统一错误处理

- [ ] **任务完成**

**背景**:
建立统一的错误处理机制，包括业务错误码、错误包装等。

#### Step 1: 定义业务错误

**新建文件**: `pkg/errors/errors.go`
```go
package errors

import "fmt"

// BizError 业务错误
type BizError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *BizError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *BizError) Unwrap() error {
    return e.Err
}

// New 创建业务错误
func New(code int, message string) *BizError {
    return &BizError{Code: code, Message: message}
}

// Wrap 包装错误
func Wrap(err error, code int, message string) *BizError {
    return &BizError{Code: code, Message: message, Err: err}
}

// 预定义错误码
const (
    CodeSuccess          = 0
    CodeValidationError  = 10001
    CodeUnauthorized     = 10002
    CodeForbidden        = 10003
    CodeNotFound         = 10004
    CodeInternalError    = 50000
    
    // 用户相关
    CodeUserNotFound     = 20001
    CodeUserExists       = 20002
    CodePasswordError    = 20003
)

// 预定义错误
var (
    ErrValidation    = New(CodeValidationError, "参数验证失败")
    ErrUnauthorized  = New(CodeUnauthorized, "未授权")
    ErrForbidden     = New(CodeForbidden, "禁止访问")
    ErrNotFound      = New(CodeNotFound, "资源不存在")
    ErrInternal      = New(CodeInternalError, "服务器内部错误")
    ErrUserNotFound  = New(CodeUserNotFound, "用户不存在")
    ErrUserExists    = New(CodeUserExists, "用户已存在")
    ErrPassword      = New(CodePasswordError, "密码错误")
)
```

#### Step 2: 更新 Response 处理

**更新文件**: `app/common/response/response.go`
```go
package response

import (
    "gin-web/pkg/errors"
    "github.com/gin-gonic/gin"
    "net/http"
)

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        ErrorCode: errors.CodeSuccess,
        Data:      data,
        Message:   "ok",
    })
}

func Error(c *gin.Context, err error) {
    if bizErr, ok := err.(*errors.BizError); ok {
        c.JSON(http.StatusOK, Response{
            ErrorCode: bizErr.Code,
            Data:      nil,
            Message:   bizErr.Message,
        })
        return
    }
    
    // 未知错误
    c.JSON(http.StatusInternalServerError, Response{
        ErrorCode: errors.CodeInternalError,
        Data:      nil,
        Message:   "服务器内部错误",
    })
}
```

---

## 完成检查清单

- [ ] Wire 依赖注入已实现
- [ ] Repository 层已添加
- [ ] Controller 路由自动注册已实现
- [ ] 统一错误处理已完成
- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 文档已更新

---

## 目录结构变更

```
gin-web/
├── cmd/
│   └── server/
│       └── main.go           # 新入口
├── internal/
│   ├── container/
│   │   ├── providers.go      # 新增
│   │   ├── wire.go           # 新增
│   │   └── wire_gen.go       # 自动生成
│   └── repository/
│       ├── repository.go     # 新增
│       ├── user_repository.go # 新增
│       └── base_repository.go # 新增
├── pkg/
│   └── errors/
│       └── errors.go         # 新增
├── app/
│   ├── controllers/
│   │   ├── controller.go     # 新增
│   │   └── user.go           # 重构
│   └── services/
│       └── user.go           # 重构
└── ...
```

---

## 参考资源

- [Google Wire 文档](https://github.com/google/wire)
- [Uber Fx 文档](https://github.com/uber-go/fx)
- [Clean Architecture in Go](https://github.com/bxcodec/go-clean-arch)
- [Hyperf DI 文档](https://hyperf.wiki/3.0/#/zh-cn/di)