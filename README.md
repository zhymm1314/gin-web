# Gin Web API 脚手架

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.10.0-00ADD8?style=flat)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-1.6.0-brightgreen.svg)](CHANGELOG.md)

基于 Gin 框架的 Go 语言后端 API 脚手架，采用 MVC + Repository 架构模式，为 PHP/Hyperf 开发者提供熟悉的开发体验。

> 参考文章：[从 PHP 到 Go：Hyperf 开发者的 Gin 框架指南](https://juejin.cn/post/7016742808560074783)

## 核心特性

- **MVC + Repository 架构** - Controller → Service → Repository → Model 四层分层
- **模块化设计** - 基于 Module 接口的插件化架构，配置驱动启停
- **依赖注入** - 基于 Google Wire 的编译时 DI 容器
- **JWT 认证** - 完整认证体系，支持令牌黑名单
- **消息队列** - RabbitMQ 生产者/消费者模式
- **定时任务** - 基于 robfig/cron 的任务调度
- **WebSocket** - 基于 Melody 的实时通信
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

# 启动服务
go run main.go

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
│   ├── controllers/        # 控制器层
│   ├── services/           # 服务层
│   ├── models/             # 数据模型
│   ├── middleware/         # 中间件
│   ├── common/             # 请求/响应结构体
│   ├── cron/               # 定时任务实现
│   └── amqp/               # 消息队列生产者/消费者
├── internal/               # 内部包
│   ├── container/          # Wire 依赖注入容器
│   └── repository/         # 数据仓储层
├── pkg/                    # 可复用公共包
│   ├── app/                # 模块化应用管理器
│   ├── cron/               # 定时任务管理器
│   ├── rabbitmq/           # RabbitMQ 管理器
│   ├── websocket/          # WebSocket 管理器
│   └── errors/             # 统一错误定义
├── bootstrap/              # 引导初始化（配置/数据库/日志/路由等）
├── config/                 # 配置结构体定义
├── routes/                 # 路由定义
├── cmd/                    # 独立服务启动入口
│   ├── consumer/           # RabbitMQ 消费者服务
│   ├── cron/               # 定时任务服务
│   └── websocket/          # WebSocket 服务
├── storage/logs/           # 日志文件
├── docs/                   # 文档与 Swagger 生成文件
├── config.yaml             # 配置文件
└── main.go                 # 主入口
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
| WebSocket | Melody | 依赖注入 | Wire |
| 参数验证 | validator | API 文档 | Swaggo |

## 贡献

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/xxx`)
3. 提交更改 (`git commit -m 'Add xxx'`)
4. 推送并创建 Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE)
