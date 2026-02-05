# API 接口开发指南

本文档详细说明如何在 Gin-Web 项目中开发一个完整的 API 接口。

> **v2.0.0 更新**: 项目使用 Uber fx 进行依赖注入，类似 Spring Boot / Hyperf 的开发体验。

---

## 目录

- [概述](#概述)
- [为什么使用依赖注入](#为什么使用依赖注入)
- [开发流程总览](#开发流程总览)
- [详细步骤](#详细步骤)
  - [Step 1: 定义 DTO（数据传输对象）](#step-1-定义-dto数据传输对象)
  - [Step 2: 定义数据模型](#step-2-定义数据模型)
  - [Step 3: 实现 Repository 层](#step-3-实现-repository-层)
  - [Step 4: 实现 Service 层](#step-4-实现-service-层)
  - [Step 5: 实现 Controller 层](#step-5-实现-controller-层)
  - [Step 6: 注册到 fx 容器](#step-6-注册到-fx-容器)
  - [Step 7: 添加 Swagger 文档](#step-7-添加-swagger-文档)
- [Swagger 文档最佳实践](#swagger-文档最佳实践)
- [常用响应方法](#常用响应方法)
- [参数验证](#参数验证)
- [错误处理](#错误处理)
- [注意事项](#注意事项)

---

## 概述

本项目采用 **Controller → Service → Repository → Model** 四层架构模式，并通过 **Uber fx 依赖注入容器** 管理所有依赖关系：

```
┌─────────────────────────────────────┐
│         fx DI Container             │  ← Uber fx 依赖注入
├─────────────────────────────────────┤
│ Controller  │  ← 处理 HTTP 请求/响应
├─────────────┤
│  Service    │  ← 业务逻辑处理
├─────────────┤
│ Repository  │  ← 数据访问抽象
├─────────────┤
│   Model     │  ← 数据模型定义
└─────────────┘
```

---

## 为什么使用依赖注入

### 与全局变量对比

| 特性 | 全局变量 (global.App) | 依赖注入 (fx) |
|------|----------------------|---------------|
| **可测试性** | 难以 Mock，需要修改全局状态 | 轻松注入 Mock 依赖 |
| **依赖可见性** | 隐式依赖，不清楚组件需要什么 | 显式依赖，构造函数声明所需依赖 |
| **解耦程度** | 组件与全局变量强耦合 | 组件间完全解耦 |
| **生命周期管理** | 手动管理 | 自动管理（启动/关闭钩子） |
| **并发安全** | 需要额外考虑 | 容器自动处理 |

### 依赖注入的核心优势

#### 1. 单元测试 Mock

```go
// 使用依赖注入，可以轻松 Mock 依赖进行单元测试
func TestUserService_Register(t *testing.T) {
    // 创建 Mock Repository
    mockRepo := &MockUserRepository{
        FindByMobileFunc: func(mobile string) (*models.User, error) {
            return nil, gorm.ErrRecordNotFound // 模拟用户不存在
        },
        CreateFunc: func(user *models.User) error {
            user.ID = 1 // 模拟创建成功
            return nil
        },
    }

    // 注入 Mock 依赖
    service := services.NewUserService(mockRepo, zap.NewNop())

    // 执行测试
    user, err := service.Register(dto.RegisterRequest{
        Name:     "测试用户",
        Mobile:   "13800138000",
        Password: "123456",
    })

    assert.NoError(t, err)
    assert.Equal(t, uint(1), user.ID)
}
```

#### 2. 显式依赖声明

```go
// 通过构造函数，一眼就能看出组件需要哪些依赖
func NewUserService(
    repo repository.UserRepository,  // 需要用户仓储
    log *zap.Logger,                  // 需要日志
) *UserService {
    return &UserService{repo: repo, log: log}
}
```

#### 3. 生命周期管理

```go
// fx 自动管理组件的启动和关闭
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error {
        // 应用启动时执行
        return server.ListenAndServe()
    },
    OnStop: func(ctx context.Context) error {
        // 应用关闭时执行（优雅关闭）
        return server.Shutdown(ctx)
    },
})
```

#### 4. 自动装配

```go
// fx 自动解析依赖关系，无需手动 new 和传递
// 只需声明 Provider，fx 自动完成装配
fx.Provide(
    NewUserRepository,  // 提供 UserRepository
    NewUserService,     // 自动注入 UserRepository
    NewAuthController,  // 自动注入 UserService
)
```

---

## 开发流程总览

开发一个新的 API 接口需要以下步骤：

1. **定义 DTO** - `app/dto/` 目录
2. **定义数据模型** - `app/models/` 目录
3. **实现 Repository** - `internal/repository/` 目录
4. **实现 Service** - `app/services/` 目录
5. **实现 Controller** - `app/controllers/` 目录
6. **注册到 fx 容器** - `internal/fx/` 目录
7. **添加 Swagger 文档** - 在 Controller 方法上添加注释

---

## 详细步骤

### Step 1: 定义 DTO（数据传输对象）

在 `app/dto/` 目录下创建 DTO 文件。每个模块的 Request 和 Response 放在同一个文件中。

**文件位置**: `app/dto/article.go`

```go
package dto

// ================================
// 文章模块 DTO（Data Transfer Object）
// ================================

// -------------------- Request --------------------

// CreateArticleRequest 创建文章请求
// @Description 创建文章所需信息
type CreateArticleRequest struct {
    Title   string `form:"title" json:"title" binding:"required,min=1,max=200" example:"文章标题"`   // 文章标题
    Content string `form:"content" json:"content" binding:"required,min=10" example:"文章内容..."`   // 文章内容
    Status  int    `form:"status" json:"status" binding:"oneof=0 1" example:"1"`                   // 状态：0草稿 1发布
}

// GetMessages 自定义验证错误信息
func (req CreateArticleRequest) GetMessages() ValidatorMessages {
    return ValidatorMessages{
        "Title.required":   "文章标题不能为空",
        "Title.min":        "文章标题至少1个字符",
        "Title.max":        "文章标题最多200个字符",
        "Content.required": "文章内容不能为空",
        "Content.min":      "文章内容至少10个字符",
        "Status.oneof":     "状态值只能是0或1",
    }
}

// ArticleQueryRequest 文章查询请求
// @Description 文章分页查询参数
type ArticleQueryRequest struct {
    Page     int `form:"page" json:"page" binding:"min=1" example:"1"`                // 页码
    PageSize int `form:"page_size" json:"page_size" binding:"min=1,max=100" example:"20"` // 每页数量
}

// -------------------- Response --------------------

// ArticleResponse 文章响应
// @Description 文章详情
type ArticleResponse struct {
    ID        uint   `json:"id" example:"1"`                            // 文章ID
    Title     string `json:"title" example:"文章标题"`                    // 标题
    Content   string `json:"content" example:"文章内容..."`               // 内容
    Status    int    `json:"status" example:"1"`                        // 状态
    CreatedAt string `json:"created_at" example:"2024-01-01 12:00:00"`  // 创建时间
}

// ArticleListResponse 文章列表响应
// @Description 文章分页列表
type ArticleListResponse struct {
    List       []ArticleResponse `json:"list"`        // 文章列表
    Total      int64             `json:"total"`       // 总数
    Page       int               `json:"page"`        // 当前页
    PageSize   int               `json:"page_size"`   // 每页数量
    TotalPages int               `json:"total_pages"` // 总页数
}
```

**DTO 设计规范**:

| 规范 | 说明 |
|------|------|
| 文件命名 | 按模块命名，如 `auth.go`、`article.go` |
| Request 后缀 | 请求结构体使用 `XxxRequest` 命名 |
| Response 后缀 | 响应结构体使用 `XxxResponse` 命名 |
| `example` tag | 必须添加，用于 Swagger 文档示例 |
| 字段注释 | 必须添加，用于 Swagger 字段描述 |
| `GetMessages()` | Request 实现此方法以自定义验证错误信息 |

### Step 2: 定义数据模型

在 `app/models/` 目录下定义数据模型。

**文件位置**: `app/models/article.go`

```go
package models

import "time"

// Article 文章模型
type Article struct {
    ID        uint       `json:"id" gorm:"primaryKey"`
    Title     string     `json:"title" gorm:"type:varchar(200);not null"`
    Content   string     `json:"content" gorm:"type:text"`
    UserID    uint       `json:"user_id" gorm:"index"`
    Status    int        `json:"status" gorm:"default:0"` // 0:草稿 1:发布
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"-" gorm:"index"` // 软删除
}

// TableName 自定义表名
func (Article) TableName() string {
    return "gw_articles"
}
```

### Step 3: 实现 Repository 层

Repository 层负责数据访问抽象，与具体的数据库操作解耦。

**文件位置**: `internal/repository/article_repository.go`

```go
package repository

import (
    "gin-web/app/models"
    "gorm.io/gorm"
)

// ArticleRepository 文章仓储接口
type ArticleRepository interface {
    Create(article *models.Article) error
    FindByID(id uint) (*models.Article, error)
    FindByUserID(userID uint, page, pageSize int) ([]models.Article, int64, error)
    Update(article *models.Article) error
    Delete(id uint) error
}

// articleRepository 实现
type articleRepository struct {
    db *gorm.DB
}

// NewArticleRepository 创建文章仓储实例
func NewArticleRepository(db *gorm.DB) ArticleRepository {
    return &articleRepository{db: db}
}

func (r *articleRepository) Create(article *models.Article) error {
    return r.db.Create(article).Error
}

func (r *articleRepository) FindByID(id uint) (*models.Article, error) {
    var article models.Article
    err := r.db.First(&article, id).Error
    if err != nil {
        return nil, err
    }
    return &article, nil
}

func (r *articleRepository) FindByUserID(userID uint, page, pageSize int) ([]models.Article, int64, error) {
    var articles []models.Article
    var total int64

    query := r.db.Model(&models.Article{}).Where("user_id = ?", userID)
    query.Count(&total)

    offset := (page - 1) * pageSize
    err := query.Offset(offset).Limit(pageSize).Find(&articles).Error

    return articles, total, err
}

func (r *articleRepository) Update(article *models.Article) error {
    return r.db.Save(article).Error
}

func (r *articleRepository) Delete(id uint) error {
    return r.db.Delete(&models.Article{}, id).Error
}
```

### Step 4: 实现 Service 层

Service 层处理业务逻辑，通过构造函数注入依赖。

**文件位置**: `app/services/article.go`

```go
package services

import (
    "gin-web/app/dto"
    "gin-web/app/models"
    "gin-web/internal/repository"
    bizErr "gin-web/pkg/errors"
    "go.uber.org/zap"
)

// ArticleService 文章服务
type ArticleService struct {
    repo repository.ArticleRepository
    log  *zap.Logger
}

// NewArticleService 创建文章服务
func NewArticleService(repo repository.ArticleRepository, log *zap.Logger) *ArticleService {
    return &ArticleService{repo: repo, log: log}
}

// Create 创建文章
func (s *ArticleService) Create(userID uint, req dto.CreateArticleRequest) (*models.Article, error) {
    article := &models.Article{
        Title:   req.Title,
        Content: req.Content,
        UserID:  userID,
        Status:  req.Status,
    }

    if err := s.repo.Create(article); err != nil {
        s.log.Error("创建文章失败", zap.Error(err))
        return nil, bizErr.Wrap(err, bizErr.CodeInternalError, "创建文章失败")
    }

    return article, nil
}

// GetByID 获取文章详情
func (s *ArticleService) GetByID(id uint) (*models.Article, error) {
    article, err := s.repo.FindByID(id)
    if err != nil {
        return nil, bizErr.New(bizErr.CodeNotFound, "文章不存在")
    }
    return article, nil
}

// List 获取用户文章列表
func (s *ArticleService) List(userID uint, page, pageSize int) ([]models.Article, int64, error) {
    return s.repo.FindByUserID(userID, page, pageSize)
}
```

### Step 5: 实现 Controller 层

Controller 必须实现 `Controller` 接口，支持自动路由注册。

**文件位置**: `app/controllers/article_controller.go`

```go
package controllers

import (
    "strconv"

    "github.com/gin-gonic/gin"

    "gin-web/app/dto"
    "gin-web/app/middleware"
    "gin-web/app/services"
)

// ArticleController 文章控制器
type ArticleController struct {
    articleService *services.ArticleService
    jwtMiddleware  *middleware.JwtMiddleware
}

// NewArticleController 创建文章控制器
func NewArticleController(articleService *services.ArticleService, jwtMiddleware *middleware.JwtMiddleware) *ArticleController {
    return &ArticleController{
        articleService: articleService,
        jwtMiddleware:  jwtMiddleware,
    }
}

// Prefix 路由前缀
func (ctrl *ArticleController) Prefix() string {
    return "/article"
}

// Routes 路由列表
func (ctrl *ArticleController) Routes() []Route {
    return []Route{
        {Method: "POST", Path: "/create", Handler: ctrl.Create, Middlewares: []gin.HandlerFunc{ctrl.jwtMiddleware.JWTAuth(services.AppGuardName)}},
        {Method: "GET", Path: "/:id", Handler: ctrl.Detail},
        {Method: "GET", Path: "/list", Handler: ctrl.List, Middlewares: []gin.HandlerFunc{ctrl.jwtMiddleware.JWTAuth(services.AppGuardName)}},
    }
}

// Create 创建文章
// @Summary      创建文章
// @Description  创建新文章
// @Tags         文章
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body dto.CreateArticleRequest true "文章信息"
// @Success      200 {object} dto.Response{data=dto.ArticleResponse} "成功"
// @Failure      400 {object} dto.Response "参数错误"
// @Router       /article/create [post]
func (ctrl *ArticleController) Create(c *gin.Context) {
    var req dto.CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        dto.ValidateFail(c, dto.GetErrorMsg(req, err))
        return
    }

    userID := c.GetUint("user_id")
    article, err := ctrl.articleService.Create(userID, req)
    if err != nil {
        dto.BusinessFail(c, err.Error())
        return
    }

    dto.Success(c, article)
}

// Detail 文章详情
// @Summary      获取文章详情
// @Description  根据 ID 获取文章详情
// @Tags         文章
// @Produce      json
// @Param        id path int true "文章ID"
// @Success      200 {object} dto.Response{data=dto.ArticleResponse} "成功"
// @Failure      404 {object} dto.Response "文章不存在"
// @Router       /article/{id} [get]
func (ctrl *ArticleController) Detail(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))

    article, err := ctrl.articleService.GetByID(uint(id))
    if err != nil {
        dto.BusinessFail(c, err.Error())
        return
    }

    dto.Success(c, article)
}

// List 文章列表
// @Summary      获取文章列表
// @Description  获取当前用户的文章列表
// @Tags         文章
// @Produce      json
// @Security     Bearer
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} dto.Response{data=dto.ArticleListResponse} "成功"
// @Router       /article/list [get]
func (ctrl *ArticleController) List(c *gin.Context) {
    var req dto.ArticleQueryRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        dto.ValidateFail(c, dto.GetErrorMsg(req, err))
        return
    }

    userID := c.GetUint("user_id")
    articles, total, err := ctrl.articleService.List(userID, req.Page, req.PageSize)
    if err != nil {
        dto.BusinessFail(c, err.Error())
        return
    }

    dto.Success(c, gin.H{
        "list":  articles,
        "total": total,
    })
}
```

### Step 6: 注册到 fx 容器

项目使用 Uber fx 进行依赖注入，只需在对应模块文件中添加 Provider：

#### 6.1 添加 Repository Provider

在 `internal/fx/repository.go` 中添加：

```go
func ProvideArticleRepository(db *gorm.DB) repository.ArticleRepository {
    if db == nil {
        return nil
    }
    return repository.NewArticleRepository(db)
}

// 更新 RepositoryModule
var RepositoryModule = fx.Module("repository",
    fx.Provide(
        ProvideUserRepository,
        ProvideModRepository,
        ProvideArticleRepository, // 新增
    ),
)
```

#### 6.2 添加 Service Provider

在 `internal/fx/service.go` 中添加：

```go
func ProvideArticleService(
    repo repository.ArticleRepository,
    log *zap.Logger,
) *services.ArticleService {
    return services.NewArticleService(repo, log)
}

// 更新 ServiceModule
var ServiceModule = fx.Module("service",
    fx.Provide(
        ProvideUserService,
        ProvideJwtService,
        ProvideModService,
        ProvideArticleService, // 新增
    ),
)
```

#### 6.3 添加 Controller Provider（使用分组注入）

在 `internal/fx/controller.go` 中添加：

```go
func NewArticleController(
    articleSvc *services.ArticleService,
    jwtMw *middleware.JwtMiddleware,
) controllers.Controller {
    return controllers.NewArticleController(articleSvc, jwtMw)
}

// 更新 ControllerModule（使用 group 自动注册）
var ControllerModule = fx.Module("controller",
    fx.Provide(
        fx.Annotate(NewAuthController, fx.ResultTags(`group:"controllers"`)),
        fx.Annotate(NewModController, fx.ResultTags(`group:"controllers"`)),
        fx.Annotate(NewArticleController, fx.ResultTags(`group:"controllers"`)), // 新增
    ),
)
```

#### 6.4 无需额外步骤

- **无需运行代码生成命令** - fx 运行时自动解析依赖
- **无需修改 main.go** - 控制器通过 group 自动注册
- **无需手动注册路由** - RegisterRoutes 自动收集所有控制器

### Step 7: 添加 Swagger 文档

运行以下命令生成 Swagger 文档：

```bash
swag init
```

---

## Swagger 文档最佳实践

### DTO 自动生成文档

通过在 DTO 结构体上添加完整的 tag 和注释，Swagger 会自动生成字段说明：

```go
// LoginRequest 用户登录请求
// @Description 用户登录凭证
type LoginRequest struct {
    Mobile   string `json:"mobile" binding:"required,mobile" example:"13800138000"` // 手机号码
    Password string `json:"password" binding:"required" example:"123456"`           // 登录密码
}
```

生成的 Swagger 文档会包含：
- 字段名和类型
- 示例值（来自 `example` tag）
- 字段描述（来自行尾注释）
- 是否必填（来自 `binding:"required"`）

### Controller 注释规范

```go
// Login 用户登录
// @Summary      用户登录
// @Description  用户登录获取 JWT Token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "登录信息"
// @Success      200 {object} dto.Response{data=dto.LoginResponse} "成功"
// @Failure      400 {object} dto.Response "参数错误"
// @Failure      401 {object} dto.Response "认证失败"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
    // ...
}
```

### 响应结构嵌套

使用 `dto.Response{data=具体类型}` 格式声明嵌套响应：

```go
// @Success 200 {object} dto.Response{data=dto.ArticleResponse} "成功"
// @Success 200 {object} dto.Response{data=[]dto.ArticleResponse} "成功（数组）"
```

---

## 常用响应方法

项目提供了统一的响应封装，位于 `app/dto/response.go`：

```go
// 成功响应
dto.Success(c, data)

// 参数验证失败
dto.ValidateFail(c, "错误信息")

// 业务逻辑失败
dto.BusinessFail(c, "错误信息")

// 自定义状态码响应
dto.Fail(c, errorCode, "错误信息")
```

**响应格式示例**:

```json
// 成功响应
{
    "error_code": 0,
    "message": "ok",
    "data": { ... }
}

// 失败响应
{
    "error_code": 10001,
    "message": "参数验证失败",
    "data": null
}
```

---

## 参数验证

### 绑定方法选择

| 方法 | 使用场景 | Tag |
|------|----------|-----|
| `ShouldBindJSON` | JSON Body | `json` |
| `ShouldBindQuery` | URL Query | `form` |
| `ShouldBind` | Form Data | `form` |
| `Param()` | URL Path 参数 | - |

### 常用验证规则

| 规则 | 说明 | 示例 |
|------|------|------|
| `required` | 必填字段 | `binding:"required"` |
| `min` | 最小值/长度 | `binding:"min=1"` |
| `max` | 最大值/长度 | `binding:"max=100"` |
| `email` | 邮箱格式 | `binding:"email"` |
| `mobile` | 手机号格式 | `binding:"mobile"` (自定义) |
| `oneof` | 枚举值 | `binding:"oneof=0 1 2"` |
| `len` | 精确长度 | `binding:"len=11"` |

### 自定义验证规则

在 `bootstrap/validator.go` 中添加自定义规则：

```go
// 手机号验证
validate.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
    mobile := fl.Field().String()
    ok, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
    return ok
})
```

---

## 错误处理

使用 `pkg/errors` 包进行统一错误处理：

```go
import bizErr "gin-web/pkg/errors"

// 创建业务错误
err := bizErr.New(bizErr.CodeUserNotFound, "用户不存在")

// 包装错误
err := bizErr.Wrap(dbErr, bizErr.CodeInternalError, "数据库操作失败")

// 预定义错误
bizErr.ErrUserNotFound
bizErr.ErrUserExists
bizErr.ErrPassword
```

---

## 注意事项

1. **DTO 设计**: Request/Response 放在同一模块文件中（`app/dto/`），便于维护
2. **Repository** 接口定义与实现分离，便于单元测试 mock
3. **Service** 层不应直接依赖 `*gin.Context`，保持业务逻辑纯净
4. **Controller** 只负责请求处理和响应，不包含业务逻辑
5. **必须使用 fx 依赖注入模式**开发新功能
6. 所有对外 API 响应使用 `dto` 包统一格式
7. **中间件使用**: 通过注入的 `jwtMiddleware.JWTAuth()` 方法
8. **新增控制器**使用 `fx.Annotate` + `group:"controllers"` 自动注册
9. **Swagger 文档**: DTO 添加 `example` tag 和字段注释，自动生成文档
