# Swagger API 文档指南

本指南介绍如何使用 Swagger 自动生成 API 文档。

## 目录

- [快速开始](#快速开始)
- [访问文档](#访问文档)
- [注释语法](#注释语法)
- [常用注释示例](#常用注释示例)
- [生成文档](#生成文档)
- [最佳实践](#最佳实践)

---

## 快速开始

### 1. 安装 swag 工具

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 生成文档

```bash
swag init
```

### 3. 启动服务

```bash
go run main.go
```

### 4. 访问文档

浏览器打开: `http://localhost:8889/swagger/index.html`

---

## 访问文档

启动服务后，Swagger UI 可通过以下地址访问：

- **Swagger UI**: `http://localhost:{port}/swagger/index.html`
- **JSON 格式**: `http://localhost:{port}/swagger/doc.json`

> **注意**: 生产环境 (`app.env: production`) 下 Swagger 路由会被禁用。

---

## 注释语法

### 主入口注释 (main.go)

```go
// @title           Gin-Web API
// @version         1.6.0
// @description     Gin-Web 脚手架 API 文档

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

### 控制器方法注释

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

---

## 常用注释示例

### GET 请求 - 查询参数

```go
// Search 搜索 Mod
// @Summary      搜索 Mod
// @Description  根据关键词、游戏、分类等条件搜索 Mod
// @Tags         Mod
// @Accept       json
// @Produce      json
// @Param        keyword query string false "搜索关键词"
// @Param        game_id query int false "游戏ID"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /mods/search [get]
func (mc *ModController) Search(c *gin.Context) {
    // ...
}
```

### GET 请求 - 路径参数

```go
// Detail 获取 Mod 详情
// @Summary      获取 Mod 详情
// @Description  根据 ID 获取 Mod 详细信息
// @Tags         Mod
// @Accept       json
// @Produce      json
// @Param        id path int true "Mod ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /mods/{id} [get]
func (mc *ModController) Detail(c *gin.Context) {
    // ...
}
```

### POST 请求 - JSON Body

```go
// Login 用户登录
// @Summary      用户登录
// @Description  使用手机号和密码登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body request.Login true "登录信息"
// @Success      200 {object} response.Response{data=response.TokenResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
    // ...
}
```

### 需要认证的接口

```go
// Info 获取当前用户信息
// @Summary      获取当前用户信息
// @Description  获取已登录用户的详细信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} response.Response{data=models.User}
// @Failure      401 {object} response.Response
// @Router       /auth/info [get]
func (c *AuthController) Info(ctx *gin.Context) {
    // ...
}
```

---

## 生成文档

### 基本命令

```bash
# 在项目根目录执行
swag init
```

### 指定目录

```bash
# 指定 main.go 路径和输出目录
swag init -g main.go -o docs/
```

### 常用参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-g` | 指定入口文件 | `-g main.go` |
| `-o` | 指定输出目录 | `-o docs/` |
| `-d` | 指定搜索目录 | `-d ./` |
| `--parseDependency` | 解析依赖 | `--parseDependency` |

### Makefile 集成

```makefile
# Makefile
swagger:
	swag init
	@echo "Swagger docs generated"
	@echo "Visit: http://localhost:8889/swagger/index.html"
```

---

## 最佳实践

### 1. 统一 Tags 分类

按功能模块划分 Tags：

```go
// @Tags         认证      // 认证相关接口
// @Tags         用户      // 用户管理接口
// @Tags         Mod       // Mod 相关接口
// @Tags         WebSocket // WebSocket 接口
```

### 2. 详细描述

提供清晰的 Summary 和 Description：

```go
// @Summary      用户注册
// @Description  创建新用户账号，需要提供手机号、密码和用户名
```

### 3. 完整的响应定义

定义所有可能的响应状态：

```go
// @Success      200 {object} response.Response{data=models.User}
// @Failure      400 {object} response.Response "参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      404 {object} response.Response "未找到"
// @Failure      500 {object} response.Response "服务器错误"
```

### 4. 安全定义

统一使用 Bearer Token 认证：

```go
// main.go
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入 Bearer {token}

// 需要认证的接口
// @Security     Bearer
```

### 5. 代码修改后重新生成

每次修改 API 注释后，需要重新运行：

```bash
swag init
```

---

## 常见问题

### Q: 文档不显示最新修改？

A: 运行 `swag init` 重新生成文档，然后重启服务。

### Q: 生产环境如何禁用 Swagger？

A: 在 `config.yaml` 设置 `app.env: production`，框架会自动禁用 Swagger 路由。

### Q: 如何添加请求示例？

A: 使用 `example` 标签：

```go
// @Param request body request.Login true "登录信息" example({"mobile":"13800138000","password":"123456"})
```

---

## 参考资源

- [Swaggo 官方文档](https://github.com/swaggo/swag)
- [Swagger 注释语法](https://github.com/swaggo/swag#declarative-comments-format)
- [gin-swagger](https://github.com/swaggo/gin-swagger)
