# Gin Web API 脚手架

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.10.0-00ADD8?style=flat)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-2.0.0-brightgreen.svg)](CHANGELOG.md)

**「不重复造轮子，专注写业务」**

每个新项目都要从零搭建目录结构、配置日志、接入数据库、写一遍 JWT 认证？太累了。这个脚手架把这些重复劳动都替你做好了，clone 下来改改配置就能直接写业务代码。

基于 Gin 框架，采用 MVC + Repository 分层架构，集成 Uber fx 依赖注入，开箱即用。对 PHP/Hyperf、Java/Spring Boot 开发者友好，上手无压力。

> 参考文章：[从 PHP 到 Go：Hyperf 开发者的 Gin 框架指南](https://juejin.cn/post/7016742808560074783)

## 核心特性

- **MVC + Repository 架构** - Controller → Service → Repository → Model 四层分层
- **模块化设计** - 基于 fx.Module 的插件化架构，配置驱动启停
- **依赖注入** - 基于 Uber fx 的运行时 DI 容器（类似 Spring Boot）
- **JWT 认证** - 完整认证体系，支持令牌黑名单
- **消息队列** - RabbitMQ 生产者/消费者模式
- **定时任务** - 基于 robfig/cron 的任务调度
- **WebSocket** - 基于 Melody 的实时通信
- **生命周期管理** - fx.Lifecycle 自动管理组件启动/关闭
- **统一规范** - 标准化错误码、响应格式、参数验证

## 文档导航

| 文档 | 说明 |
|------|------|
| [API 接口开发指南](docs/API_DEVELOPMENT.md) | 从零开始开发完整 API 接口 |
| [中间件使用指南](docs/MIDDLEWARE_GUIDE.md) | 内置中间件与自定义开发 |
| [RabbitMQ 指南](docs/RABBITMQ_GUIDE.md) | 消息队列生产者/消费者开发 |
| [定时任务指南](docs/CRON_GUIDE.md) | 定时任务开发与管理 |
| [WebSocket 指南](docs/WEBSOCKET_GUIDE.md) | WebSocket 实时通信开发 |
| [Swagger 指南](docs/SWAGGER_GUIDE.md) | API 文档自动生成 |
| [更新日志](CHANGELOG.md) | 版本更新记录 |

## 快速开始

### 环境要求

- Go 1.19+、MySQL 8.0+、Redis 6.0+
- RabbitMQ 3.8+（可选）

### 安装运行

```bash
# 克隆项目
git clone <repository-url>
cd gin-web

# 安装依赖
go mod tidy

# 配置环境
cp example-config.yaml config.yaml
# 编辑 config.yaml 修改数据库、Redis 连接信息

# 启动服务（fx 自动管理生命周期）
go run main.go

# 启动时会看到 fx 的依赖注入日志
# [Fx] PROVIDE    *config.Configuration
# [Fx] PROVIDE    *zap.Logger
# [Fx] PROVIDE    *gorm.DB
# [Fx] INVOKE     fx.RegisterRoutes()
# [Fx] RUNNING

# 访问 Swagger 文档
# http://localhost:8889/swagger/index.html
```

### Docker 部署

```bash
docker build -t gin-web-api .
docker run -d -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml gin-web-api
```

## 项目结构

```
gin-web/
├── app/                    # 应用核心代码
│   ├── controllers/        # 控制器层（实现 Controller 接口）
│   ├── services/           # 服务层（业务逻辑）
│   ├── models/             # 数据模型（GORM）
│   ├── dto/                # 数据传输对象（请求/响应/错误码）
│   ├── middleware/         # 中间件（JWT、Recovery、Cors）
│   ├── api/                # HTTP 客户端（外部 API 调用）
│   ├── cron/               # 定时任务实现
│   └── amqp/               # 消息队列
│       ├── producer/       # 生产者
│       └── consumer/       # 消费者
├── test/                   # 单元测试
│   └── services/           # Service 层测试（testify + mock）
├── internal/               # 内部包（不对外暴露）
│   ├── fx/                 # fx 依赖注入模块
│   │   ├── infrastructure.go  # 基础设施 Provider（DB/Redis/Logger）
│   │   ├── repository.go      # 仓储 Provider
│   │   ├── service.go         # 服务 Provider
│   │   ├── controller.go      # 控制器 Provider
│   │   ├── middleware.go      # 中间件 Provider
│   │   ├── router.go          # 路由 Provider
│   │   ├── rabbitmq.go        # RabbitMQ 模块
│   │   ├── cron.go            # Cron 模块
│   │   ├── websocket.go       # WebSocket 模块
│   │   ├── banner.go          # 启动 Banner
│   │   ├── types.go           # 类型定义
│   │   └── modules.go         # 应用组装入口
│   └── repository/         # 数据仓储层（封装数据库操作）
├── pkg/                    # 可复用公共包
│   ├── app/                # 模块化应用管理
│   ├── cron/               # 定时任务管理器
│   ├── rabbitmq/           # RabbitMQ 管理器
│   ├── websocket/          # WebSocket 管理器
│   └── errors/             # 统一错误定义
├── bootstrap/              # 引导初始化（数据库、Redis、验证器）
├── config/                 # 配置结构体定义
├── routes/                 # 路由定义
├── utils/                  # 工具函数
├── cmd/                    # 独立服务启动入口
│   ├── consumer/           # RabbitMQ 消费者服务
│   ├── cron/               # 定时任务服务
│   └── websocket/          # WebSocket 服务
├── docs/                   # 文档与 Swagger 生成文件
├── storage/                # 存储目录
│   └── logs/               # 日志文件
├── config.yaml             # 配置文件
└── main.go                 # 主入口（fx.New）
```

## 独立服务部署

框架支持将各模块作为独立进程部署，适用于生产环境水平扩展：

```bash
go run cmd/consumer/main.go   # RabbitMQ 消费者
go run cmd/cron/main.go       # 定时任务
go run cmd/websocket/main.go  # WebSocket
```

通过 `config.yaml` 中的 `enable` 开关控制主进程是否集成启动这些模块。

## 技术栈

| 组件 | 技术 | 组件 | 技术 |
|------|------|------|------|
| Web 框架 | Gin | ORM | GORM |
| 缓存 | Redis | 日志 | Zap |
| 配置 | Viper | 认证 | golang-jwt |
| 消息队列 | RabbitMQ | 定时任务 | robfig/cron |
| WebSocket | Melody | **依赖注入** | **Uber fx** |
| 参数验证 | validator | API 文档 | Swaggo |
| **单元测试** | **testify** | Mock | testify/mock |

## fx 依赖注入

本项目使用 [Uber fx](https://github.com/uber-go/fx) 进行依赖注入，类似于 Java Spring Boot / PHP Hyperf 的 DI 容器：

```go
// main.go - 应用启动
fxmodule.NewApp().Run()

// 添加新服务只需在对应模块添加 Provider
var ServiceModule = fx.Module("service",
    fx.Provide(
        ProvideUserService,
        ProvideJwtService,
        ProvideModService,
        ProvideNewService,  // 新增服务
    ),
)
```

**优势：**
- 无需手动运行代码生成命令
- 自动解析依赖关系
- 自动管理组件生命周期
- 启动时检测循环依赖

## 单元测试

项目使用 [testify](https://github.com/stretchr/testify) 框架进行单元测试：

```bash
# 运行所有测试
go test ./...

# 运行带详细输出
go test ./... -v

# 运行特定包的测试
go test ./test/services/... -v

# 查看测试覆盖率
go test ./... -cover
```

测试示例（使用 Mock）：

```go
func TestUserService_Register_Success(t *testing.T) {
    mockRepo := new(MockUserRepository)
    logger, _ := zap.NewDevelopment()
    service := NewUserService(mockRepo, logger)

    mockRepo.On("FindByMobile", "13800138000").Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

    user, err := service.Register(dto.RegisterRequest{
        Name:     "张三",
        Mobile:   "13800138000",
        Password: "password123",
    })

    assert.NoError(t, err)
    assert.NotNil(t, user)
    mockRepo.AssertExpectations(t)
}
```

## 贡献

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/xxx`)
3. 提交更改 (`git commit -m 'Add xxx'`)
4. 推送并创建 Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE)
