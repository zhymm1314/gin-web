# 更新日志

本文件记录项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [未发布]

### 计划中
- 单元测试覆盖
- Prometheus 监控集成
- Docker 多阶段构建优化

---

## [1.6.0] - 2026-02-04

### 重大变更
- **Swagger API 文档**: 集成 Swagger UI，启动后访问 `/swagger/index.html`
- **定时任务系统**: 基于 robfig/cron 的定时任务封装
- **WebSocket 支持**: 基于 Melody 的 WebSocket 封装
- **多启动模式**: 支持框架集成启动和独立脚本启动

### 新增
- `docs/` - Swagger 自动生成的 API 文档
- `pkg/cron/manager.go` - 定时任务管理器
- `pkg/websocket/manager.go` - WebSocket 管理器 (基于 Melody)
- `app/cron/cleanup_job.go` - 示例清理任务
- `app/cron/health_check_job.go` - 示例健康检查任务
- `app/controllers/websocket_controller.go` - WebSocket 控制器
- `cmd/consumer/main.go` - RabbitMQ 消费者独立启动脚本
- `cmd/cron/main.go` - 定时任务独立启动脚本
- `cmd/websocket/main.go` - WebSocket 独立启动脚本
- `config/cron.go` - 定时任务配置
- `config/websocket.go` - WebSocket 配置
- 控制器方法添加 Swagger 注释

### 变更
- `config/rabbitmq.go` - 添加 `Enable` 字段
- `config/config.go` - 添加 `Cron` 和 `WebSocket` 配置
- `config.yaml` - 添加 `cron.enable`、`websocket.enable`、`rabbitmq.enable` 配置项
- `bootstrap/router.go` - 添加 Swagger 路由注册
- `main.go` - 集成定时任务、WebSocket、配置开关

### 依赖更新
- 添加 `github.com/swaggo/gin-swagger` v1.6.1
- 添加 `github.com/swaggo/files` v1.0.1
- 添加 `github.com/swaggo/swag` v1.16.6
- 添加 `github.com/robfig/cron/v3` v3.0.1
- 添加 `github.com/olahol/melody` v1.4.0
- 升级 Go 版本要求至 1.23.0

---

## [1.5.0] - 2026-02-04

### 重大变更
- **Wire 依赖注入**: 使用 Google Wire 重构依赖注入系统

### 变更
- 手动 DI 容器改为 Wire 代码生成
- `main.go` 使用 `container.InitializeApp()` 替代 `container.NewContainer()`

### 新增
- `internal/container/provider.go` - Provider 函数定义
- `internal/container/wire.go` - Wire 配置文件
- `internal/container/wire_gen.go` - Wire 生成的代码

### 移除
- `internal/container/container.go` - 旧的手动 DI 容器

---

## [1.4.0] - 2026-02-04

### 重大变更
- **DI 系统全面激活**: 项目完全切换到依赖注入模式
- **删除 Legacy 代码**: 移除所有全局变量版本的 Service 和 Controller

### 变更
- `main.go` 使用 `container.NewContainer()` 创建 DI 容器
- `bootstrap.RunServerWithDI()` 替代 `RunServer()`
- `UserService` 统一为 DI 版本，删除 `UserServiceLegacy`
- `JwtService` 统一为 DI 版本，删除全局 `JwtService` 变量
- `JwtMiddleware` 改造为依赖注入，接收 `*services.JwtService`
- `ModService` 改造为 DI 版本
- `ModController` 实现 `Controller` 接口

### 新增
- `internal/repository/mod_repository.go` - Mod 仓储层
- `app/middleware/JwtMiddleware` 结构体
- Container 新增 `ModRepo`、`ModService`、`ModController`

### 移除
- `app/controllers/auth.go` - Legacy 认证控制器
- `app/controllers/user.go` - Legacy 用户控制器
- `services.UserServiceLegacy` 全局变量
- `services.JwtService` 全局变量
- `middleware.JWTAuth()` 函数（改为 `JwtMiddleware.JWTAuth()` 方法）

---

## [1.3.0] - 2026-02-04

### 变更
- **代码规范统一**: 函数返回值顺序统一为 `(result, error)`，符合 Go 惯例
- **目录命名修复**: `app/ampq/` 重命名为 `app/amqp/`（修复拼写错误）
- **Model 规范化**: 修复 JSON tag (`name1`→`name`, `mobile2`→`mobile`)，添加 GORM 类型约束
- **类型安全配置**: `ApiUrls` 配置从 `map[string]any` 改为强类型结构体
- **变量命名规范**: 函数参数命名统一使用小驼峰 (`GuardName`→`guardName`)
- **统一指针返回**: Service 层统一返回指针类型 `*models.User`

### 新增
- 新增 `.golangci.yml` 代码检查配置
- 新增 `.editorconfig` 编辑器配置
- 新增 `config/api_url.go` 类型安全的 API 配置
- Model 新增 `TableName()`、`MaskMobile()` 方法
- `common.go` 新增 `BaseModel` 组合类型

---

## [1.2.0] - 2026-02-04

### 新增
- 开发者文档体系（API 开发、中间件、消息队列指南）
- TODO 任务管理（P0-P3 优先级分类）
- README 文档导航和快速参考

---

## [1.1.0] - 2026-02-03

### 新增
- 依赖注入容器 (`internal/container/`)
- Repository 仓储模式 (`internal/repository/`)
- 统一错误处理 (`pkg/errors/`)
- Controller 接口自动路由注册

### 变更
- UserService/JWTService 支持 DI 模式

### 安全
- JWT 库升级至 v5，修复 CVE-2020-26160
- RabbitMQ 库统一为 amqp091-go

---

## [1.0.1] - 2025-05-16

### 新增
- Mod 相关路由和功能

### 变更
- 重构 RabbitMQ 消费者框架
- 优化日志消费者和连接管理

---

## [1.0.0] - 2024-12-30

### 新增
- **核心框架**: Gin + MVC 架构 + Viper 配置 + Zap 日志
- **认证授权**: JWT 认证、Token 黑名单、bcrypt 加密
- **数据存储**: MySQL (GORM) + Redis 缓存
- **消息队列**: RabbitMQ 生产者/消费者模式
- **中间件**: JWT 认证、CORS 跨域、异常恢复
- **部署支持**: Docker 构建、优雅关闭

---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.5.0 | 2026-02-04 | Wire 依赖注入重构 |
| 1.4.0 | 2026-02-04 | DI 系统全面激活，删除 Legacy 代码 |
| 1.3.0 | 2026-02-04 | 代码规范统一 (P2 完成) |
| 1.2.0 | 2026-02-04 | 开发文档体系 |
| 1.1.0 | 2026-02-03 | DI 容器、Repository 模式、安全修复 |
| 1.0.1 | 2025-05-16 | Mod 功能、RabbitMQ 重构 |
| 1.0.0 | 2024-12-30 | 初始版本 |

---

## 贡献指南

提交代码时请在 `[未发布]` 部分添加变更记录：

- `新增` - 新功能
- `变更` - 功能变更
- `修复` - Bug 修复
- `安全` - 安全修复
- `移除` - 移除功能
