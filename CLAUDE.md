# CLAUDE.md — gin-web 脚手架导航地图

> 本文件是给 AI 的**项目索引**，不是教程。目标：精确定位、少读文件、省 token。
> 详细操作流程按需调用 skill：`/add-api`、`/add-mq-consumer`、`/add-cron`。
> 人类向文档见 `README.md` 与 `docs/*.md`，本文件不重复其内容。

## 一句话定位
Gin + Uber fx 依赖注入的企业级 Go API 脚手架。MVC + Repository 四层分层，配置开关驱动模块启停。Go 1.25+。

## 分层与请求链路
```
请求: /api -> Controller -> Service -> Repository -> Model(GORM)
装配: fx.Module 顺序 Repository -> Service -> Middleware -> Controller -> Router
      (internal/fx/modules.go)
路由: Controller 按 group:"controllers" 自动收集，按 Prefix()+Routes() 自动注册，无需手写路由表。
```

## 目录 -> 职责 路由表（找 X 去 Y）
| 要做什么 | 去哪 |
|---|---|
| HTTP 入口/自动路由 | `routes/api.go`, `app/controllers/controller.go`(RegisterController) |
| Controller 层 | `app/controllers/`（须实现 `Controller` 接口：`Prefix()`+`Routes()`） |
| Service 业务逻辑 | `app/services/`（具体结构体，注入 repo+logger） |
| Repository 数据访问 | `internal/repository/`（接口+私有实现，Service 不直接碰 `*gorm.DB`） |
| Model | `app/models/`（可嵌 `BaseModel`/`Timestamps`/`SoftDeletes`，见 `common.go`） |
| DTO/请求响应/错误码/校验 | `app/dto/`（`response.go`,`errors.go`,`validator.go`） |
| fx 装配/Provider | `internal/fx/`（`repository.go`,`service.go`,`controller.go`,`router.go`,`mq.go`,`cron.go`,`websocket.go`） |
| 中间件 | `app/middleware/`（`jwt.go`,`cors.go`,`recovery.go`） |
| 统一业务错误 | `pkg/errors/errors.go`（`BizError`/`Wrap`/`New`） |
| MQ 抽象 | `pkg/mq/`；Kafka 实现 `pkg/kafka/`；RabbitMQ `pkg/rabbitmq/` |
| MQ 消费者/生产者 | `app/amqp/{consumer,producer}/` |
| MQ 入口/装配 | `cmd/consumer/main.go`, `internal/fx/mq.go` |
| 消费者清单配置 | `config/yaml/consumer.yaml` |
| 定时任务 | `app/cron/`, `pkg/cron/manager.go`, `internal/fx/cron.go`, `cmd/cron/main.go` |
| WebSocket | `pkg/websocket/`, `app/controllers/websocket_controller.go`, `cmd/websocket/main.go` |
| 配置结构体 | `config/*.go`；运行配置 `config.yaml` |
| 引导初始化 | `bootstrap/`（db,redis,log,validator,rabbitmq） |
| 工具函数 | `utils/` |
| 单元测试 | `test/services/`（testify+mock） |
| 专题指南 | `docs/*.md`（API/MQ/CRON/WS/中间件/Swagger） |

## 常见任务定位
| 任务 | 范例文件 | skill |
|---|---|---|
| 新增 API 接口 | `app/controllers/auth_controller.go`、`mod.go` | `/add-api` |
| 新增 MQ 消费者 | `app/amqp/consumer/log_consumer.go` | `/add-mq-consumer` |
| 新增定时任务 | `app/cron/health_check_job.go` | `/add-cron` |

## 关键约定（避免踩坑）
- **HTTP 状态码恒 200**（除 500），错误用 `dto.Response.error_code` 区分。Controller 用 `dto.Success`/`BusinessFail`/`ValidateFail`，**不要** `c.JSON(http.StatusBadRequest,...)`。
- **Controller fx provider 必须返回 `controllers.Controller` 接口**且带 `fx.ResultTags(\`group:"controllers"\`)`，否则路由不收集。
- **Repository 用接口**，provider 在 `db==nil` 时返回 `nil`；Service/Controller 用具体结构体。
- **Service 不直接操作 `*gorm.DB`**，查询条件封 `XxxSearchCriteria` 传 Repository。
- **校验**：自定义 tag `mobile`/`email` 须在 `internal/fx/validator.go` 注册（不注册会 panic）；**不要**注册 `RegisterTagNameFunc`（会破坏 `GetMessages` key 匹配）。
- **MQ 选型**：`kafka.enable` 优先，否则 `rabbitmq.enable`（二选一，同时启用优先 Kafka）。这两个开关控制主进程是否加载 MQ；独立进程 `cmd/consumer` 强制启用不受影响。
- **cron 表达式 6 段含秒**（`pkg/cron` 用 `WithSeconds()`，`秒 分 时 日 月 周`）。主进程跑 cron 须 `config.yaml` 的 `cron.enable: true`（默认 false，仅 `cmd/cron` 独立进程运行）；独立进程强制启用不受影响。
- **模块开关**：`config.yaml` 里 `kafka.enable`/`rabbitmq.enable`/`cron.enable`/`websocket.enable` 控制主进程是否集成；独立进程 `cmd/*` 强制启用。

## 技术栈
Gin · GORM · Redis · Zap · Viper · golang-jwt · Kafka(sarama)/RabbitMQ · robfig/cron v3 · Melody(WS) · Uber fx · testify

## 常用命令
- 运行主进程：`go run main.go`
- 独立服务：`go run cmd/{consumer,cron,websocket}/main.go`
- 测试：`go test ./...`
- 本地基础设施：`docker compose up -d`（mysql/redis/postgresql/kafka）
