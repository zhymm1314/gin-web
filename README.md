# Gin Web API 脚手架

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.10.0-00ADD8?style=flat)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-1.5.0-brightgreen.svg)](CHANGELOG.md)

一个基于 Gin 框架的企业级 Go 语言后端 API 脚手架，采用标准的 MVC 架构模式，为 PHP 开发者提供友好的 Go 语言开发体验。

📋 **[查看更新日志 (CHANGELOG)](CHANGELOG.md)** | 🚀 **当前版本: v1.5.0**

---

## 📚 文档导航

> **新手入门？** 请按以下顺序阅读文档：

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [API 接口开发指南](docs/API_DEVELOPMENT.md) | 从零开始开发一个完整的 API 接口 | 开发新功能、新接口 |
| [中间件使用指南](docs/MIDDLEWARE_GUIDE.md) | 内置中间件使用与自定义中间件开发 | 认证、限流、日志、权限控制 |
| [RabbitMQ 消息队列指南](docs/RABBITMQ_GUIDE.md) | 消息队列的生产者和消费者开发 | 异步任务、解耦服务 |

---

## 🎯 项目初衷

本项目旨在实现 Hyperf 框架到 Gin 框架的无缝切换，为 PHP 开发者（特别是 Hyperf 用户）提供熟悉的开发体验。通过标准化的项目结构和开箱即用的功能模块，让开发者能够快速上手 Go 语言后端开发。

> 参考文章：[从 PHP 到 Go：Hyperf 开发者的 Gin 框架指南](https://juejin.cn/post/7016742808560074783)

## ✨ 核心特性

### 🏗️ 架构设计
- **MVC + Repository 架构模式**：清晰的分层设计，Controller-Service-Repository-Model 四层架构
- **依赖注入容器**：基于 Google Wire 的 DI 容器，编译时依赖注入
- **控制器自动注册**：类似 Hyperf 的控制器路由自动注册机制
- **统一错误处理**：标准化的业务错误码和错误包装
- **中间件支持**：JWT 认证、CORS 跨域、异常恢复等中间件
- **配置管理**：基于 Viper 的 YAML 配置文件，支持热重载

### 🔐 认证授权
- **JWT 令牌认证**：完整的用户认证体系
- **令牌黑名单**：支持令牌撤销和黑名单管理
- **密码加密**：使用 bcrypt 进行密码安全存储
- **权限中间件**：路由级别的权限控制

### 💾 数据存储
- **MySQL 数据库**：基于 GORM 的 ORM 操作
- **Redis 缓存**：支持缓存和会话存储
- **数据库迁移**：自动表结构创建和数据初始化
- **连接池管理**：数据库连接池优化配置

### 📝 日志系统
- **结构化日志**：基于 Zap 的高性能日志系统
- **日志分类**：应用日志、数据库日志分类存储
- **日志轮转**：自动日志文件轮转和压缩
- **统一存储**：所有日志统一存储在 `storage/logs/` 目录

### 🔄 消息队列
- **RabbitMQ 集成**：完整的生产者-消费者模式
- **多队列支持**：支持多个队列的并发处理
- **消费者管理**：自动重连和错误处理机制
- **配置化管理**：通过配置文件管理队列和消费者

### 🛠️ 开发工具
- **参数验证**：基于 validator 的请求参数验证
- **统一响应**：标准化的 API 响应格式
- **错误处理**：全局异常捕获和处理
- **热重载**：开发环境支持配置文件热重载

## 🚀 技术栈

| 组件 | 技术选型 | 版本 | 说明 |
|------|----------|------|------|
| **Web 框架** | Gin | v1.10.0 | 高性能 HTTP Web 框架 |
| **ORM** | GORM | v1.25.12 | Go 语言 ORM 库 |
| **数据库** | MySQL | 8.0+ | 关系型数据库 |
| **缓存** | Redis | v8.11.5 | 内存数据库 |
| **日志** | Zap | v1.27.0 | 高性能结构化日志 |
| **配置** | Viper | v1.19.0 | 配置文件管理 |
| **认证** | JWT | v5.2.1 | JSON Web Token (golang-jwt) |
| **消息队列** | RabbitMQ | v1.10.0 | 消息中间件 (amqp091-go) |
| **密码加密** | bcrypt | - | 密码哈希算法 |
| **参数验证** | validator | v10.23.0 | 结构体验证 |
| **依赖注入** | Wire | v0.6.0 | Google 编译时依赖注入 |

## 📁 项目结构

```
gin-web/
├── app/                          # 应用核心代码
│   ├── controllers/              # 控制器层
│   │   ├── controller.go        # 控制器接口定义
│   │   ├── auth_controller.go   # 认证控制器
│   │   └── mod.go               # Mod 控制器
│   ├── services/                # 服务层
│   │   ├── user.go              # 用户服务
│   │   ├── jwt.go               # JWT 服务
│   │   └── mod.go               # Mod 服务
│   ├── models/                  # 模型层
│   │   ├── user.go              # 用户模型
│   │   └── common.go            # 公共模型
│   ├── middleware/              # 中间件
│   │   ├── jwt.go               # JWT 中间件
│   │   ├── cors.go              # CORS 中间件
│   │   └── recovery.go          # 异常恢复中间件
│   ├── common/                  # 公共组件
│   │   ├── request/             # 请求结构体
│   │   └── response/            # 响应处理
│   ├── amqp/                    # 消息队列 (AMQP)
│   │   ├── consumer/            # 消费者
│   │   └── producer/            # 生产者
│   └── api/                     # API 客户端
├── internal/                    # 内部包 (不对外暴露)
│   ├── container/               # 依赖注入容器 (Wire)
│   │   ├── provider.go          # Provider 函数定义
│   │   ├── wire.go              # Wire 配置文件
│   │   └── wire_gen.go          # Wire 生成的代码
│   └── repository/              # 仓储层
│       ├── repository.go        # 仓储接口定义
│       └── user_repository.go   # 用户仓储实现
├── pkg/                         # 可复用的公共包
│   └── errors/                  # 统一错误处理
│       └── errors.go            # 业务错误定义
├── bootstrap/                   # 引导程序
│   ├── config.go                # 配置初始化
│   ├── db.go                    # 数据库初始化
│   ├── log.go                   # 日志初始化
│   ├── redis.go                 # Redis 初始化
│   ├── router.go                # 路由初始化
│   ├── rabbitmq_manager.go      # RabbitMQ 管理器
│   └── validator.go             # 验证器初始化
├── config/                      # 配置结构体
│   ├── config.go                # 配置汇总
│   ├── app.go                   # 应用配置
│   ├── database.go              # 数据库配置
│   ├── log.go                   # 日志配置
│   ├── jwt.go                   # JWT 配置
│   ├── redis.go                 # Redis 配置
│   └── queue.go                 # 队列配置
├── routes/                      # 路由定义
│   └── api.go                   # API 路由
├── storage/                     # 存储目录
│   └── logs/                    # 日志文件
│       ├── app.log              # 应用日志
│       └── sql.log              # 数据库日志
├── global/                      # 全局变量
│   ├── app.go                   # 应用实例
│   ├── error.go                 # 错误定义
│   └── lock.go                  # 分布式锁
├── utils/                       # 工具函数
├── docs/                        # 文档目录
│   ├── API_DEVELOPMENT.md       # API 开发指南
│   ├── MIDDLEWARE_GUIDE.md      # 中间件指南
│   └── RABBITMQ_GUIDE.md        # RabbitMQ 指南
├── config.yaml                  # 配置文件
├── example-config.yaml          # 配置文件模板
├── main.go                      # 程序入口
├── go.mod                       # Go 模块文件
├── go.sum                       # 依赖校验文件
├── Dockerfile                   # Docker 构建文件
└── README.md                    # 项目说明
```

## 🚀 快速开始

### 环境要求

- **Go**: 1.19 或更高版本
- **MySQL**: 8.0 或更高版本
- **Redis**: 6.0 或更高版本
- **RabbitMQ**: 3.8 或更高版本（可选）

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd backend
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **配置环境**
   ```bash
   # 复制配置文件模板
   cp example-config.yaml config.yaml
   
   # 编辑配置文件，修改数据库、Redis 等连接信息
   vim config.yaml
   ```

4. **初始化数据库**
   ```bash
   # 创建数据库
   mysql -u root -p -e "CREATE DATABASE gin_web CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
   
   # 导入测试数据（可选）
   mysql -u root -p gin_web < test_data.sql
   ```

5. **启动服务**
   ```bash
   go run main.go
   ```

6. **验证服务**
   ```bash
   # 健康检查
   curl http://localhost:8080/api/ping
   
   # 用户注册
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"name":"test","mobile":"13800138000","password":"123456"}'
   ```

### Docker 部署

```bash
# 构建镜像
docker build -t gin-web-api .

# 运行容器
docker run -d \
  --name gin-web-api \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/storage:/app/storage \
  gin-web-api
```

## 📖 开发指南

> **详细文档请参阅**: [API 接口开发指南](docs/API_DEVELOPMENT.md)

### 快速开发速查

#### 开发一个新接口的步骤

```
1. app/common/request/   → 定义请求结构体
2. app/models/           → 定义数据模型
3. internal/repository/  → 实现数据访问层 (可选但推荐)
4. app/services/         → 实现业务逻辑
5. app/controllers/      → 实现控制器
6. routes/api.go         → 注册路由
```

#### 代码模板

**请求结构体** (`app/common/request/xxx.go`):
```go
type CreateXxx struct {
    Name string `json:"name" binding:"required"`
}

func (req CreateXxx) GetMessages() ValidatorMessages {
    return ValidatorMessages{
        "Name.required": "名称不能为空",
    }
}
```

**控制器方法**:
```go
func CreateXxx(c *gin.Context) {
    var req request.CreateXxx
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidateFail(c, request.GetErrorMsg(req, err))
        return
    }
    // 调用 Service
    result, err := services.XxxService.Create(req)
    if err != nil {
        response.BusinessFail(c, err.Error())
        return
    }
    response.Success(c, result)
}
```

### API 接口

#### 认证接口
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出
- `GET /api/auth/info` - 获取用户信息

#### 用户接口
- `GET /api/user` - 获取用户列表（需要认证）
- `GET /api/user/:id` - 获取用户详情（需要认证）

### 分层架构

| 层级 | 目录 | 职责 |
|------|------|------|
| Controller | `app/controllers/` | 处理 HTTP 请求/响应 |
| Service | `app/services/` | 业务逻辑处理 |
| Repository | `internal/repository/` | 数据访问抽象 |
| Model | `app/models/` | 数据模型定义 |

### 依赖注入模式 (Wire)

```go
// 1. 实现 Controller 接口
type MyController struct {
    myService *services.MyService
}

func (c *MyController) Prefix() string {
    return "/my"
}

func (c *MyController) Routes() []controllers.Route {
    return []controllers.Route{
        {Method: "GET", Path: "/list", Handler: c.List},
        {Method: "POST", Path: "/create", Handler: c.Create},
    }
}

// 2. 在 internal/container/provider.go 添加 Provider
func ProvideMyController(service *services.MyService) *controllers.MyController {
    return controllers.NewMyController(service)
}

// 3. 运行 wire 命令重新生成
// $ wire ./internal/container/
```

### 中间件使用

> **详细文档请参阅**: [中间件使用指南](docs/MIDDLEWARE_GUIDE.md)

#### 内置中间件

| 中间件 | 说明 | 使用方式 |
|--------|------|----------|
| `middleware.JWTAuth()` | JWT 认证 | 路由组/单路由 |
| `middleware.Cors()` | 跨域处理 | 全局 |
| `middleware.CustomRecovery()` | 异常恢复 | 全局 |

#### 快速使用

```go
// 路由组中间件
authRouter := router.Group("/api").Use(middleware.JWTAuth(services.AppGuardName))
{
    authRouter.GET("/user/info", controllers.UserInfo)
}

// 控制器中定义中间件
func (ctrl *MyController) Routes() []Route {
    return []Route{
        {
            Method:      "POST",
            Path:        "/create",
            Handler:     ctrl.Create,
            Middlewares: []gin.HandlerFunc{middleware.JWTAuth(services.AppGuardName)},
        },
    }
}
```

---

### 消息队列使用

> **详细文档请参阅**: [RabbitMQ 消息队列指南](docs/RABBITMQ_GUIDE.md)

#### 快速开发消费者

```
1. app/amqp/consumer/  → 实现 ConsumerHandler 接口
2. main.go             → 注册处理器到 handlers map
3. config.yaml         → 配置消费者队列
```

**消费者模板** (`app/amqp/consumer/xxx_consumer.go`):
```go
type XxxConsumer struct{}

func (c *XxxConsumer) HandleMessage(msg amqp.Delivery) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered: %v", r)
        }
    }()

    // 解析消息
    var data YourStruct
    json.Unmarshal(msg.Body, &data)

    // 处理业务逻辑
    // ...

    return nil // 返回 nil 确认消费
}
```

**配置队列** (`config.yaml`):
```yaml
consumers:
  - queue: "xxx_queue"
    handler: "xxx_consumer"
    concurrency: 2
```

---

### 配置说明

主要配置项说明：

```yaml
app:
  env: local          # 环境：local/dev/prod
  port: 8080         # 服务端口
  app_name: gin-web  # 应用名称

log:
  level: info                    # 日志级别
  root_dir: ./storage/logs      # 日志目录
  max_size: 500                 # 单文件最大大小(MB)
  max_age: 28                   # 保留天数

database:
  driver: mysql               # 数据库驱动
  host: 127.0.0.1            # 数据库地址
  port: 3306                  # 数据库端口
  database: gin_web           # 数据库名
  username: root              # 用户名
  password: password          # 密码

jwt:
  secret: your-secret-key     # JWT 密钥
  jwt_ttl: 43200             # Token 有效期(秒)

redis:
  host: 127.0.0.1            # Redis 地址
  port: 6379                 # Redis 端口
  db: 0                      # 数据库编号
```

## 🔧 功能特性

### 已支持功能

| 功能模块 | 说明 |
|----------|------|
| **MVC + Repository 架构** | 清晰的四层分层设计 |
| **依赖注入容器 (Wire)** | 基于 Google Wire 的编译时依赖注入 |
| **JWT 用户认证** | 完整的认证体系，支持令牌刷新和黑名单 |
| **MySQL + GORM** | ORM 操作，自动迁移 |
| **Redis 缓存** | 缓存和会话存储 |
| **RabbitMQ 消息队列** | 生产者-消费者模式 |
| **Zap 日志系统** | 高性能结构化日志 |
| **参数验证** | 基于 validator 的请求验证 |
| **统一响应格式** | 标准化 API 响应 |
| **中间件系统** | JWT、CORS、异常恢复等 |
| **优雅关闭** | 支持 Graceful Shutdown |
| **Docker 支持** | 容器化部署 |

### 规划中功能

- API 文档自动生成 (Swagger)
- 限流中间件
- 监控指标收集 (Prometheus)
- 链路追踪 (OpenTelemetry)

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Wire](https://github.com/google/wire) - 依赖注入
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 日志库
- [Hyperf](https://hyperf.io/) - 设计灵感来源

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发起 Pull Request

---

**让 PHP 开发者轻松上手 Go 语言后端开发！** 🚀

