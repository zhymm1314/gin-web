---
name: add-api
description: 在 gin-web 脚手架中按四层+fx规范新增一个完整 API 接口（Model/Repository/Service/Controller/DTO/路由/fx注册）
---

# 新增 API 接口

按本脚手架 `Controller -> Service -> Repository -> Model` 四层 + Uber fx 装配新增一个接口。
范例参照 `app/controllers/auth_controller.go`、`app/controllers/mod.go`、`app/services/user.go`、`internal/repository/user_repository.go`。

## 链路与命名约定
| 层 | 目录 | 命名 | 形态 |
|---|---|---|---|
| Controller | `app/controllers/` | `XxxController` | 导出结构体，**必须实现 `Controller` 接口**（`Prefix() string` + `Routes() []Route`，见 `controller.go:13-19`） |
| Service | `app/services/` | `XxxService` | 导出结构体（非接口），构造返回 `*XxxService` |
| Repository | `internal/repository/` | `XxxRepository` | 导出**接口** + 小写私有实现，构造返回接口类型 |
| Model | `app/models/` | `Xxx` | GORM 结构体，可嵌 `BaseModel`(`common.go:8-29`) |

构造函数统一 `NewXxx(...)`。Service/Repository 注入 `repo` + `*zap.Logger`；Controller 注入 Service（可选 `*middleware.JwtMiddleware`）。

## 步骤（以 `Article` 为例）

### 1. Model — `app/models/article.go`
```go
package models

type Article struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Title   string `json:"title" gorm:"size:255;not null" binding:"required"`
	Content string `json:"content" gorm:"type:text"`
	Timestamps  // 或嵌入 BaseModel，见 common.go
}
```

### 2. Repository — `internal/repository/article_repository.go`
```go
package repository

import (
	"gin-web/app/models"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(article *models.Article) error
	FindByID(id uint) (*models.Article, error)
}

type articleRepository struct{ db *gorm.DB }

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}
func (r *articleRepository) FindByID(id uint) (*models.Article, error) {
	var a models.Article
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}
```
复杂查询条件封成 `XxxSearchCriteria` 结构体放同文件（见 `mod_repository.go:9-27`），**Service 不直接碰 `*gorm.DB`**。

### 3. DTO — `app/dto/article.go`
```go
package dto

type CreateArticleRequest struct {
	Title   string `form:"title" json:"title" binding:"required" example:"标题"`
	Content string `form:"content" json:"content" binding:"required" example:"正文"`
}

// 自定义校验文案（实现 Validator 接口，key = Go字段名 + "." + tag）
func (r CreateArticleRequest) GetMessages() ValidatorMessages {
	return ValidatorMessages{
		"Title.required":   "标题不能为空",
		"Content.required": "正文不能为空",
	}
}

type ArticleResponse struct {
	ID      uint   `json:"id" example:"1"`
	Title   string `json:"title" example:"标题"`
	Content string `json:"content" example:"正文"`
}
```
Request 同时带 `form`/`json`/`binding`/`example` 四 tag（见 `auth.go:12-15`）。

### 4. Service — `app/services/article.go`
```go
package services

import (
	"go.uber.org/zap"

	"gin-web/app/dto"
	"gin-web/app/models"
	"gin-web/internal/repository"
	bizErr "gin-web/pkg/errors"
)

type ArticleService struct {
	repo repository.ArticleRepository
	log  *zap.Logger
}

func NewArticleService(repo repository.ArticleRepository, log *zap.Logger) *ArticleService {
	return &ArticleService{repo: repo, log: log}
}

func (s *ArticleService) Create(req dto.CreateArticleRequest) (*models.Article, error) {
	article := &models.Article{Title: req.Title, Content: req.Content}
	if err := s.repo.Create(article); err != nil {
		s.log.Error("create article failed", zap.Error(err))
		return nil, bizErr.Wrap(err, bizErr.CodeInternalError, "创建文章失败")
	}
	return article, nil
}
```
业务错误用 `pkg/errors` 的 `bizErr.Wrap`/`bizErr.New`（`errors.go:24-31,50-60`），新增错误码在此添加。

### 5. Controller — `app/controllers/article_controller.go`
```go
package controllers

import (
	"github.com/gin-gonic/gin"

	"gin-web/app/dto"
	"gin-web/app/services"
)

type ArticleController struct {
	articleService *services.ArticleService
}

func NewArticleController(articleService *services.ArticleService) *ArticleController {
	return &ArticleController{articleService: articleService}
}

func (c *ArticleController) Prefix() string { return "/articles" }

func (c *ArticleController) Routes() []Route {
	return []Route{
		{Method: "POST", Path: "", Handler: c.Create},
		// 需登录的路由按路由粒度挂 JWT：
		// {Method: "GET", Path: "/:id", Handler: c.Detail,
		//  Middlewares: []gin.HandlerFunc{c.jwtMiddleware.JWTAuth(services.AppGuardName)}},
	}
}

// Create 创建文章
// @Summary  创建文章
// @Tags     文章
// @Accept   json
// @Produce  json
// @Param    request body dto.CreateArticleRequest true "文章信息"
// @Success  200 {object} dto.Response
// @Router   /articles [post]
func (c *ArticleController) Create(ctx *gin.Context) {
	var form dto.CreateArticleRequest
	if err := ctx.ShouldBindJSON(&form); err != nil {
		dto.ValidateFail(ctx, dto.GetErrorMsg(form, err))
		return
	}
	article, err := c.articleService.Create(form)
	if err != nil {
		dto.BusinessFail(ctx, err.Error())
		return
	}
	dto.Success(ctx, article)
}
```
Controller 三段式：`ShouldBind` -> 调 Service -> `dto.Success`（见 `auth_controller.go:53-66`）。

### 6. fx 注册（改 3 个文件，各加 2 行）

**`internal/fx/repository.go`** — `RepositoryModule` 的 `fx.Provide` 加一行 + 新增 provider：
```go
fx.Provide(
	ProvideUserRepository,
	ProvideModRepository,
	ProvideArticleRepository, // 新增
),
// db==nil 返回 nil 是项目惯例
func ProvideArticleRepository(db *gorm.DB) repository.ArticleRepository {
	if db == nil {
		return nil
	}
	return repository.NewArticleRepository(db)
}
```

**`internal/fx/service.go`** — `ServiceModule` 的 `fx.Provide` 加一行 + 新增 provider：
```go
fx.Provide(
	ProvideUserService,
	ProvideJwtService,
	ProvideModService,
	ProvideArticleService, // 新增
),
func ProvideArticleService(repo repository.ArticleRepository, log *zap.Logger) *services.ArticleService {
	return services.NewArticleService(repo, log)
}
```

**`internal/fx/controller.go`** — `ControllerModule` 加一个 `fx.Annotate` 块 + 包装构造函数：
```go
fx.Provide(
	fx.Annotate(NewAuthController, fx.ResultTags(`group:"controllers"`)),
	fx.Annotate(NewModController, fx.ResultTags(`group:"controllers"`)),
	fx.Annotate(NewArticleController, fx.ResultTags(`group:"controllers"`)), // 新增
),
// fx 层包装器，返回接口类型 controllers.Controller（分组注入硬性要求）
func NewArticleController(articleSvc *services.ArticleService) controllers.Controller {
	return controllers.NewArticleController(articleSvc)
}
```

## 路由注册
**无需改 `routes/api.go`**。路由全自动注册：`RegisterRoutes`(`router.go:89-92`) 建 `/api` group -> `SetApiGroupRoutes`(`routes/api.go:12-27`) 遍历 Controller -> `RegisterController`(`controller.go:22-39`) 按 `Prefix()` 建 sub-group 再按 `Routes()` 挂载。最终路径 = `/api` + `Prefix()` + `Path`。

## 统一响应（`app/dto/response.go`）
- `dto.Success(c, data)` → error_code=0
- `dto.BusinessFail(c, msg)` → error_code=40000
- `dto.ValidateFail(c, msg)` → error_code=42200
- `dto.TokenFail(c)` → error_code=40100
- `dto.Fail(c, code, msg)` / `dto.FailByError(c, err)` → 自定义

HTTP 状态码恒 200（除 500）。**不要** `c.JSON(http.StatusBadRequest,...)`。

## 错误码
- HTTP 层：`app/dto/errors.go`（`CodeSuccess=0`、`CodeBusinessError=40000`、`CodeTokenError=40100`、`CodeValidateError=42200`、`CodeServerError=50000`）
- Service 层：`pkg/errors/errors.go`（`CodeUserNotFound=20001` 等，`BizError` + `Wrap`/`New`）

## 绑定方式
- JSON body：`ctx.ShouldBindJSON(&form)`
- Query：`ctx.ShouldBindQuery(&req)`，DTO 用 `form` tag
- URI 参数：`ctx.ShouldBindUri(&req)`，DTO 用 `uri:"id"` tag

## 要点 / 坑
1. Controller fx provider **必须**返回 `controllers.Controller` 接口且带 `fx.ResultTags(\`group:"controllers"\`)`，少一个都不收集。
2. Repository provider 返回接口类型，`db==nil` 返回 `nil`；Service/Controller 用具体结构体。
3. 自定义校验 tag `mobile`/`email` 必须在 `internal/fx/validator.go` 注册（不注册会 panic）；**不要**注册 `RegisterTagNameFunc`（会破坏 `GetMessages` key）。
4. JWT 按路由挂（`Route.Middlewares` 加 `c.jwtMiddleware.JWTAuth(services.AppGuardName)`，`AppGuardName="app"`），登录用户 ID 取 `ctx.Keys["id"].(string)`。
5. 新模块只需加进对应 module 的 `fx.Provide`，**无需动 `modules.go`**（加载顺序已固定）。
6. Swagger 注释写在 Controller 方法上方，`@BasePath /api` 已在 `main.go` 声明，`@Router` 路径不含 `/api` 前缀。
