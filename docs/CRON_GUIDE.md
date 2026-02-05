# 定时任务指南

本指南介绍如何使用框架的定时任务功能。

> **v2.0.0 更新**: 定时任务现在通过 fx 依赖注入管理，使用构造函数注入依赖。

## 目录

- [快速开始](#快速开始)
- [启动方式](#启动方式)
- [创建定时任务](#创建定时任务)
- [Cron 表达式](#cron-表达式)
- [配置说明](#配置说明)
- [最佳实践](#最佳实践)

---

## 快速开始

### 1. 创建定时任务

在 `app/cron/` 目录下创建任务文件，使用构造函数注入依赖：

```go
// app/cron/my_job.go
package cron

import (
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type MyJob struct {
    db    *gorm.DB
    redis *redis.Client
    log   *zap.Logger
}

// NewMyJob 创建任务（通过 fx 注入依赖）
func NewMyJob(db *gorm.DB, redis *redis.Client, log *zap.Logger) *MyJob {
    return &MyJob{db: db, redis: redis, log: log}
}

func (j *MyJob) Name() string {
    return "my_job"
}

func (j *MyJob) Spec() string {
    return "*/5 * * * * *"  // 每5秒执行
}

func (j *MyJob) Run() {
    j.log.Info("my job running")
    // 使用注入的 db 和 redis
}
```

### 2. 注册任务

在 `internal/fx/cron.go` 中注册：

```go
func ProvideCronManager(
    lc fx.Lifecycle,
    cfg *config.Configuration,
    db *gorm.DB,
    redis *redis.Client,
    log *zap.Logger,
) *cron.Manager {
    manager := cron.NewManager(log)

    // 注册任务（通过构造函数注入依赖）
    manager.Register(appCron.NewCleanupJob(db, redis, log))
    manager.Register(appCron.NewHealthCheckJob(db, redis, log))
    manager.Register(appCron.NewMyJob(db, redis, log))  // 新增

    // ... lifecycle hooks
    return manager
}
```

### 3. 启动

确保配置 `cron.enable: true`，然后启动框架即可。

---

## 启动方式

### 方式一：跟随框架启动

修改 `config.yaml`：

```yaml
cron:
  enable: true
```

启动框架：

```bash
go run main.go
```

### 方式二：独立脚本启动

独立启动定时任务服务：

```bash
go run cmd/cron/main.go
```

**独立启动适用场景**：
- 生产环境分布式部署
- 定时任务需要独立扩展
- 避免定时任务影响 API 服务

---

## 创建定时任务

### 任务接口

所有定时任务必须实现 `JobHandler` 接口：

```go
type JobHandler interface {
    Name() string   // 任务名称（唯一标识）
    Spec() string   // Cron 表达式
    Run()           // 任务执行逻辑
}
```

### 完整示例（依赖注入版本）

```go
// app/cron/cleanup_job.go
package cron

import (
    "context"
    "time"

    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// CleanupJob 清理过期数据任务
type CleanupJob struct {
    db    *gorm.DB
    redis *redis.Client
    log   *zap.Logger
}

// NewCleanupJob 创建清理任务（通过 fx 注入依赖）
func NewCleanupJob(db *gorm.DB, redis *redis.Client, log *zap.Logger) *CleanupJob {
    return &CleanupJob{
        db:    db,
        redis: redis,
        log:   log,
    }
}

func (j *CleanupJob) Name() string {
    return "cleanup_expired_data"
}

func (j *CleanupJob) Spec() string {
    return "0 0 2 * * *"  // 每天凌晨 2 点执行
}

func (j *CleanupJob) Run() {
    startTime := time.Now()
    j.log.Info("cleanup job started")

    // 使用注入的 db
    if j.db != nil {
        result := j.db.Exec("DELETE FROM jwt_blacklist WHERE expired_at < ?", time.Now())
        j.log.Info("cleanup job completed",
            zap.Int64("deleted_rows", result.RowsAffected),
            zap.Duration("duration", time.Since(startTime)))
    }
}
```

### 健康检查任务示例

```go
// app/cron/health_check_job.go
package cron

import (
    "context"
    "time"

    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// HealthCheckJob 健康检查任务
type HealthCheckJob struct {
    db    *gorm.DB
    redis *redis.Client
    log   *zap.Logger
}

// NewHealthCheckJob 创建健康检查任务
func NewHealthCheckJob(db *gorm.DB, redis *redis.Client, log *zap.Logger) *HealthCheckJob {
    return &HealthCheckJob{
        db:    db,
        redis: redis,
        log:   log,
    }
}

func (j *HealthCheckJob) Name() string {
    return "health_check"
}

func (j *HealthCheckJob) Spec() string {
    return "*/30 * * * * *"  // 每 30 秒执行
}

func (j *HealthCheckJob) Run() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 检查数据库连接
    if j.db != nil {
        sqlDB, err := j.db.DB()
        if err == nil {
            if err := sqlDB.PingContext(ctx); err != nil {
                j.log.Warn("database health check failed", zap.Error(err))
            }
        }
    }

    // 检查 Redis 连接
    if j.redis != nil {
        if err := j.redis.Ping(ctx).Err(); err != nil {
            j.log.Warn("redis health check failed", zap.Error(err))
        }
    }
}
```

---

## Cron 表达式

框架使用 6 位 cron 表达式（支持秒级调度）：

```
秒 分 时 日 月 周
*  *  *  *  *  *
```

### 表达式说明

| 字段 | 范围 | 特殊字符 |
|------|------|----------|
| 秒 | 0-59 | * , - / |
| 分 | 0-59 | * , - / |
| 时 | 0-23 | * , - / |
| 日 | 1-31 | * , - / ? |
| 月 | 1-12 | * , - / |
| 周 | 0-6 (0=周日) | * , - / ? |

### 常用表达式

| 表达式 | 说明 |
|--------|------|
| `*/5 * * * * *` | 每 5 秒 |
| `0 */1 * * * *` | 每分钟 |
| `0 0 * * * *` | 每小时整点 |
| `0 30 * * * *` | 每小时第 30 分钟 |
| `0 0 2 * * *` | 每天凌晨 2 点 |
| `0 0 8,12,18 * * *` | 每天 8:00、12:00、18:00 |
| `0 0 0 * * 1` | 每周一 00:00 |
| `0 0 0 1 * *` | 每月 1 日 00:00 |
| `0 0 0 1 1 *` | 每年 1 月 1 日 00:00 |

### 预定义表达式

也可以使用预定义表达式：

| 表达式 | 等价于 | 说明 |
|--------|--------|------|
| `@yearly` | `0 0 0 1 1 *` | 每年 |
| `@monthly` | `0 0 0 1 * *` | 每月 |
| `@weekly` | `0 0 0 * * 0` | 每周 |
| `@daily` | `0 0 0 * * *` | 每天 |
| `@hourly` | `0 0 * * * *` | 每小时 |
| `@every 5m` | - | 每 5 分钟 |
| `@every 30s` | - | 每 30 秒 |

---

## 配置说明

### config.yaml

```yaml
cron:
  enable: true  # 框架启动时是否启用定时任务
```

### 配置结构体

```go
// config/cron.go
type Cron struct {
    Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`
}
```

---

## 最佳实践

### 1. 任务幂等性

确保任务可以重复执行不会产生副作用：

```go
func (j *CleanupJob) Run() {
    // 使用事务保证原子性
    tx := j.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 使用唯一键或条件判断避免重复处理
    tx.Exec("DELETE FROM expired_tokens WHERE expired_at < ? AND deleted = 0", time.Now())
    tx.Commit()
}
```

### 2. 分布式锁

多实例部署时，使用 Redis 分布式锁避免重复执行：

```go
func (j *CleanupJob) Run() {
    ctx := context.Background()

    // 获取分布式锁
    lockKey := "cron:cleanup_job"
    locked, err := j.redis.SetNX(ctx, lockKey, "1", time.Minute*5).Result()

    if err != nil || !locked {
        return  // 其他实例正在执行
    }
    defer j.redis.Del(ctx, lockKey)

    // 执行任务逻辑
}
```

### 3. 错误处理和告警

```go
func (j *CleanupJob) Run() {
    defer func() {
        if r := recover(); r != nil {
            j.log.Error("cron job panic",
                zap.String("job", j.Name()),
                zap.Any("error", r))
            // 发送告警通知
        }
    }()

    if err := j.execute(); err != nil {
        j.log.Error("cron job failed",
            zap.String("job", j.Name()),
            zap.Error(err))
        // 发送告警通知
    }
}
```

### 4. 任务超时控制

```go
func (j *CleanupJob) Run() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
    defer cancel()

    done := make(chan error, 1)
    go func() {
        done <- j.execute(ctx)
    }()

    select {
    case <-ctx.Done():
        j.log.Error("cron job timeout", zap.String("job", j.Name()))
    case err := <-done:
        if err != nil {
            j.log.Error("cron job failed", zap.Error(err))
        }
    }
}
```

### 5. 日志记录

记录任务执行的开始、结束和耗时：

```go
func (j *CleanupJob) Run() {
    startTime := time.Now()
    j.log.Info("cron job started", zap.String("job", j.Name()))

    // 执行任务
    result, err := j.execute()

    j.log.Info("cron job completed",
        zap.String("job", j.Name()),
        zap.Duration("duration", time.Since(startTime)),
        zap.Any("result", result),
        zap.Error(err))
}
```

---

## 常见问题

### Q: 如何动态添加/删除任务？

A: 可以扩展 Manager 添加 `AddJob` 和 `RemoveJob` 方法：

```go
func (m *Manager) AddJob(handler JobHandler) error {
    entryID, err := m.cron.AddFunc(handler.Spec(), handler.Run)
    if err != nil {
        return err
    }
    m.mu.Lock()
    m.jobs[handler.Name()] = entryID
    m.mu.Unlock()
    return nil
}

func (m *Manager) RemoveJob(name string) {
    m.mu.Lock()
    if entryID, ok := m.jobs[name]; ok {
        m.cron.Remove(entryID)
        delete(m.jobs, name)
    }
    m.mu.Unlock()
}
```

### Q: 如何查看当前运行的任务？

A: 使用 `m.cron.Entries()` 获取所有任务信息：

```go
func (m *Manager) ListJobs() []string {
    entries := m.cron.Entries()
    jobs := make([]string, 0, len(entries))
    for _, entry := range entries {
        jobs = append(jobs, fmt.Sprintf("ID: %d, Next: %s", entry.ID, entry.Next))
    }
    return jobs
}
```

---

## 参考资源

- [robfig/cron 官方文档](https://github.com/robfig/cron)
- [Cron 表达式生成器](https://crontab.guru/)
