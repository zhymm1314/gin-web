---
name: add-cron
description: 在 gin-web 脚手架中新增定时任务（app/cron 实现 JobHandler，internal/fx/cron.go 注册，6段表达式，enable 开关）
---

# 新增定时任务

基于 robfig/cron v3（`WithSeconds()`，**6 段含秒**）。可作独立进程 `cmd/cron` 启动，也可在主进程按 `cron.enable` 开关集成。
范例参照 `app/cron/health_check_job.go`、`app/cron/cleanup_job.go`。

## Job 接口（`pkg/cron/manager.go:10-15`）
```go
type JobHandler interface {
    Name() string // 任务名称，必须唯一
    Spec() string // cron 表达式（6 段含秒）
    Run()         // 执行方法，无参无返回
}
```

## 步骤

### 1. 新建 Job - `app/cron/xxx_job.go`
```go
package cron

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type XxxJob struct {
	db    *gorm.DB
	redis *redis.Client
	log   *zap.Logger
}

func NewXxxJob(db *gorm.DB, redis *redis.Client, log *zap.Logger) *XxxJob {
	return &XxxJob{db: db, redis: redis, log: log}
}

func (j *XxxJob) Name() string { return "xxx_job" }

// 6 段表达式：秒 分 时 日 月 周
func (j *XxxJob) Spec() string { return "*/30 * * * * *" }

func (j *XxxJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 业务逻辑，用 j.log 记日志
	_ = ctx
}
```
要点：
- 构造函数**显式列依赖**（`db/redis/log`），由 fx 注入，不走全局变量。
- `Run()` 无参无返回；关闭信号不会传入，长任务需自建 `context.WithTimeout` 控制超时。

### 2. fx 注册 - `internal/fx/cron.go` `ProvideCronManager`
在现有 `manager.Register` 后加一行（`internal/fx/cron.go:39-40` 之后）：
```go
manager := cron.NewManager(log)
manager.Register(appCron.NewCleanupJob(db, redis, log))
manager.Register(appCron.NewHealthCheckJob(db, redis, log))
manager.Register(appCron.NewXxxJob(db, redis, log)) // 新增
```
注意：
- job **不是** `fx.Provide` 暴露的，而是**手动 new + `manager.Register`**。
- `ProvideCronManager` 入参已有 `db *gorm.DB`、`redis *redis.Client`、`log *zap.Logger`、`cfg *config.Configuration`（`cron.go:29-35`），新 job 需要哪些就传哪些；需要新配置项时先扩 `config/cron.go` 的 `Cron` 结构体再传 `cfg.Cron.Xxx`。
- `Name()` **必须唯一**，否则后者覆盖前者、旧 entry 无法 `Remove`。

### 3. 主进程集成开关
`config.yaml` 的 `cron.enable` 默认 `false`，主进程不加载 cron（仅 `cmd/cron` 独立进程运行）。要在主进程跑定时任务，改为 `true`：
```yaml
cron:
  enable: true
```
独立进程 `cmd/cron` 不受影响（`CronModule(true)` 硬编码强制启用）。

无需改 `pkg/cron/manager.go`、`config/cron.go`、`cmd/cron/main.go`、`internal/fx/modules.go`（除非加新配置项）。

## 入口与 enable 开关
- 主进程集成（`NewApp`，`modules.go:36`）：`CronModule(cfg.Cron.Enable)` - false 时返回空模块（`cron.go:18-20`）
- 独立 cron 进程（`cmd/cron/main.go:16` -> `NewCronApp`，`modules.go:69-86`）：`CronModule(true)` 强制启用，仍走 `InfrastructureModule` 加载 config/db/redis/log
- 模块本身：`CronModule` = `fx.Provide(ProvideCronManager)` + `fx.Invoke(StartCron)`；真正启动靠 `ProvideCronManager` 里挂的 `fx.Hook`（OnStart->`manager.Start()`，OnStop->`manager.Stop()`，`cron.go:42-52`）

## cron 表达式：6 段（含秒）
`pkg/cron/manager.go:29` 用 `cron.New(cron.WithSeconds())`，6 字段解析器。
字段顺序：**秒 分 时 日 月 周**。现有范例：
- `health_check_job.go:35` -> `*/30 * * * * *`（每 30 秒）
- `cleanup_job.go:35` -> `0 0 2 * * *`（每天 02:00:00）

**写 5 段会解析报错**。

## 要点 / 坑
1. **`config.yaml` 的 `cron.enable` 默认 false** - 主进程不加载 cron（仅 `cmd/cron` 独立进程运行）；要在主进程跑定时任务设为 `true`。
2. **panic 恢复已实现**：`manager.go:45-51` 每个 job 包 `recover()`，panic 仅记 `m.log.Error` 不崩进程，单次失败不影响后续调度。
3. **优雅关闭已实现**：`Manager.Stop()`(`manager.go:75-79`) 调 `m.cron.Stop()` 返回 ctx，`<-ctx.Done()` 等待运行中的 job 完成；挂 fx `OnStop`。
4. **时区未配置**：`NewManager` 无 `cron.WithLocation(...)`，用进程本地时区（`time.Local`）。容器跨时区部署时需注意，必要时加 `cron.WithLocation(time.LoadLocation("Asia/Shanghai"))`。
5. **job 无 context 传入**：`Run()` 签名无 ctx，关闭信号不传入；长任务需在 `Run()` 内自建 `context.WithTimeout`。
6. **logger 注入**：job 构造函数收 `*zap.Logger`，不依赖全局；Manager 自身也收 logger（`manager.go:27`）。

## 一句话配方
> `app/cron/` 新建实现 `Name()/Spec()/Run()` 的结构体+构造函数 -> `internal/fx/cron.go` 的 `ProvideCronManager` 加一行 `manager.Register(...)` -> 表达式写 6 段含秒 -> 主进程跑就 `config.yaml` 的 `cron.enable: true`，独立跑就 `go run cmd/cron/main.go`。
