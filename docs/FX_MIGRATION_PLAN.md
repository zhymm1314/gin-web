# Uber fx 依赖注入改造计划

> 将 gin-web 项目从 Wire + 全局变量 改造为 Uber fx 依赖注入

## 改造目标

1. **消除全局变量** - 移除 `global.App` 全局状态
2. **运行时依赖注入** - 使用 fx 自动解析依赖图
3. **生命周期管理** - 利用 fx.Lifecycle 管理启动/关闭
4. **模块化组织** - 每个功能域一个 fx.Module
5. **多进程统一** - main/consumer/cron/websocket 共享模块

---

## 改造概览

```
改造前                              改造后
├── global/app.go (全局变量)         ├── (删除)
├── internal/container/              ├── internal/fx/
│   ├── wire.go                     │   ├── modules.go        (模块注册)
│   ├── wire_gen.go                 │   ├── infrastructure.go (基础设施)
│   └── provider.go                 │   ├── repository.go     (仓储层)
│                                   │   ├── service.go        (服务层)
│                                   │   ├── middleware.go     (中间件)
│                                   │   ├── controller.go     (控制器)
│                                   │   ├── router.go         (路由)
│                                   │   ├── rabbitmq.go       (消息队列)
│                                   │   ├── cron.go           (定时任务)
│                                   │   └── websocket.go      (WebSocket)
└── main.go (手动初始化)             └── main.go (fx.New)
```

---

## Phase 1: 添加依赖并重构全局变量

### 1.1 添加 fx 依赖

```bash
go get go.uber.org/fx
```

### 1.2 创建配置容器结构

**新建文件: `internal/fx/types.go`**

```go
package fx

import (
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
    "gorm.io/gorm"

    "gin-web/config"
)

// Infrastructure 基础设施容器（替代 global.App）
type Infrastructure struct {
    Config *config.Configuration
    DB     *gorm.DB
    Redis  *redis.Client
    Log    *zap.Logger
}
```

### 1.3 改造思路

- `global.App.Config` → 通过 fx 注入 `*config.Configuration`
- `global.App.DB` → 通过 fx 注入 `*gorm.DB`
- `global.App.Redis` → 通过 fx 注入 `*redis.Client`
- `global.App.Log` → 通过 fx 注入 `*zap.Logger`

---

## Phase 2: 创建基础设施 Provider 模块

**新建文件: `internal/fx/infrastructure.go`**

```go
package fx

import (
    "context"
    "fmt"
    "os"

    "github.com/fsnotify/fsnotify"
    "github.com/redis/go-redis/v9"
    "github.com/spf13/viper"
    "go.uber.org/fx"
    "go.uber.org/zap"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"

    "gin-web/config"
)

// InfrastructureModule 基础设施模块
var InfrastructureModule = fx.Module("infrastructure",
    fx.Provide(
        ProvideConfig,
        ProvideLogger,
        ProvideDatabase,
        ProvideRedis,
    ),
)

// ProvideConfig 提供配置
func ProvideConfig() (*config.Configuration, error) {
    configPath := "config.yaml"
    if envPath := os.Getenv("VIPER_CONFIG"); envPath != "" {
        configPath = envPath
    }

    v := viper.New()
    v.SetConfigFile(configPath)
    v.SetConfigType("yaml")

    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取配置失败: %w", err)
    }

    var cfg config.Configuration
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }

    // 热重载（可选）
    v.WatchConfig()
    v.OnConfigChange(func(in fsnotify.Event) {
        v.Unmarshal(&cfg)
    })

    return &cfg, nil
}

// ProvideLogger 提供日志器
func ProvideLogger(cfg *config.Configuration) (*zap.Logger, error) {
    var level zap.AtomicLevel
    if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
        level = zap.NewAtomicLevelAt(zap.InfoLevel)
    }

    zapCfg := zap.Config{
        Level:            level,
        Development:      cfg.App.Env != "production",
        Encoding:         "json",
        OutputPaths:      []string{"stdout", cfg.Log.RootDir + "/app.log"},
        ErrorOutputPaths: []string{"stderr"},
        // ... 更多配置
    }

    return zapCfg.Build()
}

// ProvideDatabase 提供数据库连接
func ProvideDatabase(lc fx.Lifecycle, cfg *config.Configuration, log *zap.Logger) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
        cfg.Database.UserName,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Database,
        cfg.Database.Charset,
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("连接数据库失败: %w", err)
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
    sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)

    // 生命周期管理
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            log.Info("数据库连接已建立")
            return sqlDB.Ping()
        },
        OnStop: func(ctx context.Context) error {
            log.Info("关闭数据库连接")
            return sqlDB.Close()
        },
    })

    return db, nil
}

// ProvideRedis 提供 Redis 连接
func ProvideRedis(lc fx.Lifecycle, cfg *config.Configuration, log *zap.Logger) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            if err := client.Ping(ctx).Err(); err != nil {
                return fmt.Errorf("Redis 连接失败: %w", err)
            }
            log.Info("Redis 连接已建立")
            return nil
        },
        OnStop: func(ctx context.Context) error {
            log.Info("关闭 Redis 连接")
            return client.Close()
        },
    })

    return client, nil
}
```

---

## Phase 3: 改造 Repository 层

**新建文件: `internal/fx/repository.go`**

```go
package fx

import (
    "go.uber.org/fx"
    "gorm.io/gorm"

    "gin-web/internal/repository"
)

// RepositoryModule 仓储模块
var RepositoryModule = fx.Module("repository",
    fx.Provide(
        ProvideUserRepository,
        ProvideModRepository,
    ),
)

func ProvideUserRepository(db *gorm.DB) repository.UserRepository {
    return repository.NewUserRepository(db)
}

func ProvideModRepository(db *gorm.DB) repository.ModRepository {
    return repository.NewModRepository(db)
}
```

---

## Phase 4: 改造 Service 层

**新建文件: `internal/fx/service.go`**

```go
package fx

import (
    "github.com/redis/go-redis/v9"
    "go.uber.org/fx"
    "go.uber.org/zap"

    "gin-web/app/services"
    "gin-web/config"
    "gin-web/internal/repository"
)

// ServiceModule 服务模块
var ServiceModule = fx.Module("service",
    fx.Provide(
        ProvideUserService,
        ProvideJwtService,
        ProvideModService,
    ),
)

func ProvideUserService(
    repo repository.UserRepository,
    log *zap.Logger,
) *services.UserService {
    return services.NewUserService(repo, log)
}

// JwtService 需要适配器接口
func ProvideJwtService(
    cfg *config.Configuration,
    redis *redis.Client,
    userSvc *services.UserService,
) *services.JwtService {
    return services.NewJwtService(
        &jwtConfigAdapter{cfg},
        &redisAdapter{redis},
        &userGetterAdapter{userSvc},
    )
}

func ProvideModService(
    repo repository.ModRepository,
    log *zap.Logger,
) *services.ModService {
    return services.NewModService(repo, log)
}

// ========== 适配器 ==========

type jwtConfigAdapter struct {
    cfg *config.Configuration
}

func (a *jwtConfigAdapter) GetSecret() string           { return a.cfg.Jwt.Secret }
func (a *jwtConfigAdapter) GetJwtTtl() int64            { return a.cfg.Jwt.JwtTtl }
func (a *jwtConfigAdapter) GetJwtBlacklistGracePeriod() int64 {
    return a.cfg.Jwt.JwtBlacklistGracePeriod
}
func (a *jwtConfigAdapter) GetRefreshGracePeriod() int64 {
    return a.cfg.Jwt.RefreshGracePeriod
}

type redisAdapter struct {
    client *redis.Client
}

// ... 实现 RedisClient 接口方法

type userGetterAdapter struct {
    svc *services.UserService
}

func (a *userGetterAdapter) GetUserInfo(id string) (interface{}, error) {
    return a.svc.GetUserInfo(id)
}
```

---

## Phase 5: 改造 Middleware 层

**新建文件: `internal/fx/middleware.go`**

```go
package fx

import (
    "go.uber.org/fx"

    "gin-web/app/middleware"
    "gin-web/app/services"
)

// MiddlewareModule 中间件模块
var MiddlewareModule = fx.Module("middleware",
    fx.Provide(
        ProvideJwtMiddleware,
    ),
)

func ProvideJwtMiddleware(jwtSvc *services.JwtService) *middleware.JwtMiddleware {
    return middleware.NewJwtMiddleware(jwtSvc)
}
```

---

## Phase 6: 改造 Controller 层

**新建文件: `internal/fx/controller.go`**

```go
package fx

import (
    "go.uber.org/fx"

    "gin-web/app/controllers"
    "gin-web/app/middleware"
    "gin-web/app/services"
)

// ControllerModule 控制器模块
var ControllerModule = fx.Module("controller",
    fx.Provide(
        ProvideAuthController,
        ProvideModController,
        // 使用分组注入所有控制器
        fx.Annotate(
            ProvideAuthController,
            fx.ResultTags(`group:"controllers"`),
        ),
        fx.Annotate(
            ProvideModController,
            fx.ResultTags(`group:"controllers"`),
        ),
    ),
)

// ControllersResult 控制器分组结果
type ControllersResult struct {
    fx.In
    Controllers []controllers.Controller `group:"controllers"`
}

func ProvideAuthController(
    userSvc *services.UserService,
    jwtSvc *services.JwtService,
    jwtMw *middleware.JwtMiddleware,
) controllers.Controller {
    return controllers.NewAuthController(userSvc, jwtSvc, jwtMw)
}

func ProvideModController(
    modSvc *services.ModService,
    jwtMw *middleware.JwtMiddleware,
) controllers.Controller {
    return controllers.NewModController(modSvc, jwtMw)
}
```

---

## Phase 7: 创建 Router 模块

**新建文件: `internal/fx/router.go`**

```go
package fx

import (
    "context"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
    "go.uber.org/zap"

    "gin-web/app/controllers"
    "gin-web/app/middleware"
    "gin-web/config"
    "gin-web/routes"
)

// RouterModule 路由模块
var RouterModule = fx.Module("router",
    fx.Provide(ProvideGinEngine),
    fx.Provide(ProvideHTTPServer),
    fx.Invoke(RegisterRoutes),
)

// ControllerParams 控制器参数（分组注入）
type ControllerParams struct {
    fx.In
    Controllers []controllers.Controller `group:"controllers"`
}

func ProvideGinEngine(cfg *config.Configuration) *gin.Engine {
    if cfg.App.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.New()
    r.Use(middleware.Cors())
    r.Use(middleware.CustomRecovery())
    r.Use(gin.Logger())

    return r
}

func ProvideHTTPServer(
    lc fx.Lifecycle,
    cfg *config.Configuration,
    engine *gin.Engine,
    log *zap.Logger,
) *http.Server {
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.App.Port),
        Handler: engine,
    }

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            log.Info("启动 HTTP 服务器", zap.Int("port", cfg.App.Port))
            go func() {
                if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                    log.Fatal("HTTP 服务器错误", zap.Error(err))
                }
            }()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            log.Info("关闭 HTTP 服务器")
            return server.Shutdown(ctx)
        },
    })

    return server
}

// RegisterRoutes 注册所有路由
func RegisterRoutes(engine *gin.Engine, params ControllerParams) {
    api := engine.Group("/api")
    routes.SetApiGroupRoutes(api, params.Controllers...)
}
```

---

## Phase 8: 改造 RabbitMQ 模块

**新建文件: `internal/fx/rabbitmq.go`**

```go
package fx

import (
    "context"

    "go.uber.org/fx"
    "go.uber.org/zap"

    "gin-web/app/amqp/consumer"
    "gin-web/config"
    "gin-web/pkg/rabbitmq"
)

// RabbitMQModule RabbitMQ 模块（条件加载）
func RabbitMQModule(enabled bool) fx.Option {
    if !enabled {
        return fx.Options() // 空模块
    }

    return fx.Module("rabbitmq",
        fx.Provide(ProvideRabbitMQManager),
        fx.Invoke(StartRabbitMQ),
    )
}

func ProvideRabbitMQManager(
    lc fx.Lifecycle,
    cfg *config.Configuration,
    log *zap.Logger,
) (*rabbitmq.Manager, error) {
    // 加载消费者配置
    consumerCfg, err := config.LoadConfig("./config/yaml/consumer.yaml")
    if err != nil {
        return nil, err
    }

    // 注册消费者处理器
    handlers := map[string]consumer.ConsumerHandler{
        "LogConsumer": &consumer.LogConsumer{},
    }

    managerCfg := &rabbitmq.Config{
        Host:     cfg.RabbitMQ.Host,
        Port:     cfg.RabbitMQ.Port,
        User:     cfg.RabbitMQ.User,
        Password: cfg.RabbitMQ.Password,
        Vhost:    cfg.RabbitMQ.Vhost,
    }

    manager := rabbitmq.NewManager(managerCfg, consumerCfg.Consumers, handlers, log)

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            log.Info("启动 RabbitMQ 消费者")
            go manager.Start()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            log.Info("停止 RabbitMQ 消费者")
            manager.Stop()
            return nil
        },
    })

    return manager, nil
}

func StartRabbitMQ(manager *rabbitmq.Manager) {
    // 触发依赖注入，manager 会通过 lifecycle 启动
}
```

---

## Phase 9: 改造 Cron 模块

**新建文件: `internal/fx/cron.go`**

```go
package fx

import (
    "context"

    "go.uber.org/fx"
    "go.uber.org/zap"

    appCron "gin-web/app/cron"
    "gin-web/config"
    "gin-web/pkg/cron"
)

// CronModule 定时任务模块（条件加载）
func CronModule(enabled bool) fx.Option {
    if !enabled {
        return fx.Options()
    }

    return fx.Module("cron",
        fx.Provide(ProvideCronManager),
        fx.Invoke(StartCron),
    )
}

func ProvideCronManager(
    lc fx.Lifecycle,
    cfg *config.Configuration,
    log *zap.Logger,
) *cron.Manager {
    manager := cron.NewManager(log)

    // 注册定时任务
    manager.Register(&appCron.CleanupJob{})
    manager.Register(&appCron.HealthCheckJob{})

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            log.Info("启动定时任务管理器")
            manager.Start()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            log.Info("停止定时任务管理器")
            manager.Stop()
            return nil
        },
    })

    return manager
}

func StartCron(manager *cron.Manager) {
    // 触发依赖注入
}
```

---

## Phase 10: 改造 WebSocket 模块

**新建文件: `internal/fx/websocket.go`**

```go
package fx

import (
    "context"

    "go.uber.org/fx"
    "go.uber.org/zap"

    "gin-web/app/controllers"
    "gin-web/config"
    "gin-web/pkg/websocket"
)

// WebSocketModule WebSocket 模块（条件加载）
func WebSocketModule(enabled bool) fx.Option {
    if !enabled {
        return fx.Options()
    }

    return fx.Module("websocket",
        fx.Provide(ProvideWebSocketManager),
        fx.Provide(
            fx.Annotate(
                ProvideWebSocketController,
                fx.ResultTags(`group:"controllers"`),
            ),
        ),
    )
}

func ProvideWebSocketManager(
    lc fx.Lifecycle,
    cfg *config.Configuration,
    log *zap.Logger,
) *websocket.Manager {
    manager := websocket.NewManager(log)

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            log.Info("启动 WebSocket 管理器")
            go manager.Start()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            log.Info("停止 WebSocket 管理器")
            manager.Stop()
            return nil
        },
    })

    return manager
}

func ProvideWebSocketController(manager *websocket.Manager) controllers.Controller {
    return controllers.NewWebSocketController(manager)
}
```

---

## Phase 11: 重写 main.go

**重写文件: `main.go`**

```go
package main

import (
    "go.uber.org/fx"
    "go.uber.org/fx/fxevent"
    "go.uber.org/zap"

    fxmodule "gin-web/internal/fx"
)

func main() {
    app := fx.New(
        // 基础设施
        fxmodule.InfrastructureModule,

        // 业务层
        fxmodule.RepositoryModule,
        fxmodule.ServiceModule,
        fxmodule.MiddlewareModule,
        fxmodule.ControllerModule,

        // HTTP 路由
        fxmodule.RouterModule,

        // 可选模块（根据配置动态加载）
        fx.Invoke(func(cfg *config.Configuration) fx.Option {
            return fx.Options(
                fxmodule.RabbitMQModule(cfg.RabbitMQ.Enable),
                fxmodule.CronModule(cfg.Cron.Enable),
                fxmodule.WebSocketModule(cfg.WebSocket.Enable),
            )
        }),

        // 日志
        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    )

    app.Run()
}
```

**注意**: 上面的动态模块加载方式不太对，需要改成：

```go
package main

import (
    "go.uber.org/fx"
    "go.uber.org/fx/fxevent"
    "go.uber.org/zap"

    "gin-web/config"
    fxmodule "gin-web/internal/fx"
)

func main() {
    // 预加载配置以决定模块
    cfg, _ := fxmodule.ProvideConfig()

    app := fx.New(
        // 基础设施
        fxmodule.InfrastructureModule,

        // 业务层
        fxmodule.RepositoryModule,
        fxmodule.ServiceModule,
        fxmodule.MiddlewareModule,
        fxmodule.ControllerModule,

        // HTTP 路由
        fxmodule.RouterModule,

        // 可选模块（根据配置动态加载）
        fxmodule.RabbitMQModule(cfg.RabbitMQ.Enable),
        fxmodule.CronModule(cfg.Cron.Enable),
        fxmodule.WebSocketModule(cfg.WebSocket.Enable),

        // 日志
        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    )

    app.Run()
}
```

---

## Phase 12: 重写 cmd/consumer/main.go

```go
package main

import (
    "go.uber.org/fx"
    "go.uber.org/fx/fxevent"
    "go.uber.org/zap"

    fxmodule "gin-web/internal/fx"
)

func main() {
    app := fx.New(
        // 只加载需要的基础设施
        fxmodule.InfrastructureModule,

        // RabbitMQ 消费者
        fxmodule.RabbitMQModule(true),

        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    )

    app.Run()
}
```

---

## Phase 13: 重写 cmd/cron/main.go

```go
package main

import (
    "go.uber.org/fx"
    "go.uber.org/fx/fxevent"
    "go.uber.org/zap"

    fxmodule "gin-web/internal/fx"
)

func main() {
    app := fx.New(
        // 基础设施
        fxmodule.InfrastructureModule,

        // 可能需要 Repository 和 Service
        fxmodule.RepositoryModule,
        fxmodule.ServiceModule,

        // 定时任务
        fxmodule.CronModule(true),

        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    )

    app.Run()
}
```

---

## Phase 14: 重写 cmd/websocket/main.go

```go
package main

import (
    "go.uber.org/fx"
    "go.uber.org/fx/fxevent"
    "go.uber.org/zap"

    fxmodule "gin-web/internal/fx"
)

func main() {
    app := fx.New(
        fxmodule.InfrastructureModule,
        fxmodule.RepositoryModule,
        fxmodule.ServiceModule,
        fxmodule.MiddlewareModule,

        // 只加载 WebSocket 控制器
        fxmodule.WebSocketModule(true),

        // 需要路由
        fxmodule.RouterModule,

        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    )

    app.Run()
}
```

---

## Phase 15: 删除旧文件

### 需要删除的文件

```bash
# Wire 相关
rm internal/container/wire.go
rm internal/container/wire_gen.go
rm internal/container/provider.go

# 全局变量
rm global/app.go
rm global/error.go  # 如果只在全局变量中使用
rm global/lock.go   # 如果只在全局变量中使用

# 旧的 bootstrap 文件（功能已迁移到 fx 模块）
rm bootstrap/config.go
rm bootstrap/db.go
rm bootstrap/redis.go
rm bootstrap/log.go
rm bootstrap/rabbitmq.go

# 可以保留的 bootstrap 文件
# bootstrap/router.go   - 可能需要保留部分逻辑
# bootstrap/validator.go - 验证器初始化
```

### 需要修改的文件

所有使用 `global.App.XXX` 的地方需要改为依赖注入：

```bash
# 搜索需要修改的文件
grep -r "global.App" --include="*.go" .
```

---

## Phase 16: 测试和验证

### 16.1 编译测试

```bash
# 主服务
go build -o bin/api ./main.go

# 消费者
go build -o bin/consumer ./cmd/consumer/main.go

# 定时任务
go build -o bin/cron ./cmd/cron/main.go

# WebSocket
go build -o bin/websocket ./cmd/websocket/main.go
```

### 16.2 启动测试

```bash
# 测试主服务
./bin/api

# 测试 fx 依赖图可视化
go run main.go -fx.dig dot | dot -Tpng -o fx-graph.png
```

### 16.3 接口测试

```bash
# 健康检查
curl http://localhost:8080/api/ping

# 注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test","mobile":"13800138000","password":"123456"}'

# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"mobile":"13800138000","password":"123456"}'
```

---

## 最终目录结构

```
gin-web/
├── main.go                          # fx.New() 入口
├── cmd/
│   ├── consumer/main.go             # 消费者 fx 入口
│   ├── cron/main.go                 # 定时任务 fx 入口
│   └── websocket/main.go            # WebSocket fx 入口
│
├── internal/
│   ├── fx/                          # fx 模块（新增）
│   │   ├── types.go                 # 公共类型
│   │   ├── infrastructure.go        # 基础设施 Provider
│   │   ├── repository.go            # 仓储 Provider
│   │   ├── service.go               # 服务 Provider + 适配器
│   │   ├── middleware.go            # 中间件 Provider
│   │   ├── controller.go            # 控制器 Provider
│   │   ├── router.go                # 路由 Provider
│   │   ├── rabbitmq.go              # RabbitMQ 模块
│   │   ├── cron.go                  # Cron 模块
│   │   └── websocket.go             # WebSocket 模块
│   └── repository/                  # 保持不变
│
├── app/                             # 保持不变
│   ├── controllers/
│   ├── services/
│   ├── middleware/
│   ├── models/
│   └── ...
│
├── config/                          # 保持不变
├── routes/                          # 保持不变
└── pkg/                             # 保持不变
```

---

## 改造优势

| 改造前 | 改造后 |
|--------|--------|
| Wire 编译时生成代码 | fx 运行时自动解析 |
| 需要手动运行 `wire` | 无需额外构建步骤 |
| 全局变量 `global.App` | 依赖注入，无全局状态 |
| 手动管理生命周期 | fx.Lifecycle 自动管理 |
| 分散的模块初始化 | 模块化 fx.Module 组织 |
| 多进程代码重复 | 共享模块，按需组合 |

---

## Phase 17: 更新 README.md

README.md 需要更新以下内容：

### 17.1 更新核心特性

```markdown
## 核心特性

- **MVC + Repository 架构** - Controller → Service → Repository → Model 四层分层
- **模块化设计** - 基于 fx.Module 的插件化架构，配置驱动启停
- **依赖注入** - 基于 Uber fx 的运行时 DI 容器（类似 Spring Boot）  ← 更新
- **JWT 认证** - 完整认证体系，支持令牌黑名单
...
```

### 17.2 更新项目结构

```markdown
## 项目结构

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
│   ├── fx/                 # fx 依赖注入模块      ← 新增（替代 container）
│   │   ├── infrastructure.go  # 基础设施 Provider
│   │   ├── repository.go      # 仓储 Provider
│   │   ├── service.go         # 服务 Provider
│   │   ├── controller.go      # 控制器 Provider
│   │   └── ...
│   └── repository/         # 数据仓储层
├── pkg/                    # 可复用公共包
├── bootstrap/              # 引导初始化（验证器等）  ← 精简
├── config/                 # 配置结构体定义
├── routes/                 # 路由定义
├── cmd/                    # 独立服务启动入口
├── docs/                   # 文档与 Swagger 生成文件
├── config.yaml             # 配置文件
└── main.go                 # 主入口（fx.New）
```

### 17.3 更新技术栈表格

```markdown
## 技术栈

| 组件 | 技术 | 组件 | 技术 |
|------|------|------|------|
| Web 框架 | Gin | ORM | GORM |
| 缓存 | Redis | 日志 | Zap |
| 配置 | Viper | 认证 | golang-jwt |
| 消息队列 | RabbitMQ | 定时任务 | robfig/cron |
| WebSocket | Melody | **依赖注入** | **Uber fx** |  ← 更新
| 参数验证 | validator | API 文档 | Swaggo |
```

### 17.4 更新快速开始

```markdown
### 安装运行

# 克隆项目
git clone <repository-url>
cd gin-web

# 安装依赖
go mod tidy

# 配置环境
cp example-config.yaml config.yaml

# 启动服务（fx 自动管理生命周期）
go run main.go

# 启动时会看到 fx 的依赖注入日志
# [Fx] PROVIDE    *config.Configuration
# [Fx] PROVIDE    *zap.Logger
# [Fx] PROVIDE    *gorm.DB
# ...
```

---

## Phase 18: 更新开发文档

### 18.1 更新 docs/API_DEVELOPMENT.md

**主要改动**:

1. **更新概述说明**（移除 Wire 相关）

```markdown
## 概述

本项目采用 **Controller → Service → Repository → Model** 四层架构模式，
并通过 **Uber fx 依赖注入容器** 管理所有依赖关系：

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

2. **更新 Step 6 - 注册到 DI 容器**

```markdown
### Step 6: 注册到 fx 容器

项目使用 Uber fx 进行依赖注入，只需在对应模块文件中添加 Provider：

#### 6.1 添加 Repository Provider

在 `internal/fx/repository.go` 中添加：

func ProvideArticleRepository(db *gorm.DB) repository.ArticleRepository {
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

#### 6.2 添加 Service Provider

在 `internal/fx/service.go` 中添加：

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

#### 6.3 添加 Controller Provider（使用分组注入）

在 `internal/fx/controller.go` 中添加：

func ProvideArticleController(
    articleSvc *services.ArticleService,
    jwtMw *middleware.JwtMiddleware,
) controllers.Controller {
    return controllers.NewArticleController(articleSvc, jwtMw)
}

// 更新 ControllerModule（使用 group 自动注册）
var ControllerModule = fx.Module("controller",
    fx.Provide(
        fx.Annotate(ProvideAuthController, fx.ResultTags(`group:"controllers"`)),
        fx.Annotate(ProvideModController, fx.ResultTags(`group:"controllers"`)),
        fx.Annotate(ProvideArticleController, fx.ResultTags(`group:"controllers"`)), // 新增
    ),
)

#### 6.4 无需额外步骤

- **无需运行 wire 命令** - fx 运行时自动解析
- **无需修改 main.go** - 控制器通过 group 自动注册
- **无需手动注册路由** - RegisterRoutes 自动收集所有控制器
```

3. **更新注意事项**

```markdown
## 注意事项

1. **请求结构体**必须实现 `GetMessages()` 方法以支持自定义错误信息
2. **Repository** 接口定义与实现分离，便于单元测试 mock
3. **Service** 层不应直接依赖 `*gin.Context`，保持业务逻辑纯净
4. **Controller** 只负责请求处理和响应，不包含业务逻辑
5. **必须使用 fx 依赖注入模式**开发新功能
6. 所有对外 API 响应使用 `response` 包统一格式
7. **中间件使用**: 通过注入的 `jwtMiddleware.JWTAuth()` 方法
8. **新增控制器**使用 `fx.Annotate` + `group:"controllers"` 自动注册
```

---

### 18.2 更新 docs/CRON_GUIDE.md

**移除 global.App 引用**:

```markdown
### 完整示例

// app/cron/cleanup_job.go
package cron

import (
    "context"
    "time"

    "go.uber.org/zap"
    "gorm.io/gorm"
)

// CleanupJob 清理过期数据任务
type CleanupJob struct {
    db  *gorm.DB      // 通过构造函数注入
    log *zap.Logger   // 通过构造函数注入
}

// NewCleanupJob 创建清理任务（fx 会自动注入依赖）
func NewCleanupJob(db *gorm.DB, log *zap.Logger) *CleanupJob {
    return &CleanupJob{db: db, log: log}
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

    // 使用注入的 db 而非 global.App.DB
    result := j.db.Exec("DELETE FROM jwt_blacklist WHERE expired_at < ?", time.Now())

    j.log.Info("cleanup job completed",
        zap.Int64("deleted_rows", result.RowsAffected),
        zap.Duration("duration", time.Since(startTime)))
}


### 注册任务（fx 方式）

在 `internal/fx/cron.go` 中：

func ProvideCronManager(
    lc fx.Lifecycle,
    log *zap.Logger,
    db *gorm.DB,
    redis *redis.Client,
) *cron.Manager {
    manager := cron.NewManager(log)

    // 注册任务（依赖自动注入）
    manager.Register(appCron.NewCleanupJob(db, log))
    manager.Register(appCron.NewHealthCheckJob(db, redis, log))

    // ... lifecycle hooks
    return manager
}
```

---

### 18.3 更新 docs/RABBITMQ_GUIDE.md

**移除 global.App 引用**:

```markdown
### 消费者示例（fx 依赖注入版本）

// app/amqp/consumer/order_consumer.go
package consumer

import (
    "encoding/json"

    amqp "github.com/rabbitmq/amqp091-go"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// OrderConsumer 订单消费者
type OrderConsumer struct {
    db  *gorm.DB
    log *zap.Logger
}

// NewOrderConsumer 创建订单消费者（依赖通过 fx 注入）
func NewOrderConsumer(db *gorm.DB, log *zap.Logger) *OrderConsumer {
    return &OrderConsumer{db: db, log: log}
}

func (c *OrderConsumer) HandleMessage(msg amqp.Delivery) error {
    defer func() {
        if r := recover(); r != nil {
            c.log.Error("OrderConsumer panic recovered",
                zap.Any("panic", r),
                zap.ByteString("body", msg.Body),
            )
        }
    }()

    var orderMsg OrderMessage
    if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
        c.log.Error("解析订单消息失败",
            zap.Error(err),
            zap.ByteString("body", msg.Body),
        )
        return nil
    }

    c.log.Info("处理订单消息",
        zap.Uint("order_id", orderMsg.OrderID),
        zap.String("action", orderMsg.Action),
    )

    // 使用注入的 db
    // c.db.Create(&order)

    return nil
}


### 注册消费者（fx 方式）

在 `internal/fx/rabbitmq.go` 中：

func ProvideConsumerHandlers(
    db *gorm.DB,
    log *zap.Logger,
) map[string]consumer.ConsumerHandler {
    return map[string]consumer.ConsumerHandler{
        "log_consumer":   consumer.NewLogConsumer(log),
        "order_consumer": consumer.NewOrderConsumer(db, log),
        "email_consumer": consumer.NewEmailConsumer(log),
    }
}

func ProvideRabbitMQManager(
    lc fx.Lifecycle,
    cfg *config.Configuration,
    log *zap.Logger,
    handlers map[string]consumer.ConsumerHandler,  // 自动注入
) (*rabbitmq.Manager, error) {
    // ...
}
```

---

### 18.4 更新 docs/MIDDLEWARE_GUIDE.md

**更新中间件使用方式**:

```markdown
## 使用中间件

### 在 Controller 中使用（推荐）

中间件通过构造函数注入，在 Routes() 方法中使用：

type ArticleController struct {
    articleService *services.ArticleService
    jwtMiddleware  *middleware.JwtMiddleware  // 注入的中间件
}

func NewArticleController(
    articleService *services.ArticleService,
    jwtMiddleware *middleware.JwtMiddleware,
) *ArticleController {
    return &ArticleController{
        articleService: articleService,
        jwtMiddleware:  jwtMiddleware,
    }
}

func (ctrl *ArticleController) Routes() []Route {
    return []Route{
        // 需要认证的路由
        {
            Method:      "POST",
            Path:        "/create",
            Handler:     ctrl.Create,
            Middlewares: []gin.HandlerFunc{ctrl.jwtMiddleware.JWTAuth("app")},
        },
        // 不需要认证的路由
        {
            Method:  "GET",
            Path:    "/:id",
            Handler: ctrl.Detail,
        },
    }
}
```

---

### 18.5 更新 docs/WEBSOCKET_GUIDE.md

**更新 WebSocket 使用方式**:

```markdown
## WebSocket 模块启用

### fx 模块化配置

WebSocket 模块通过 fx 条件加载：

// main.go
func main() {
    cfg, _ := fxmodule.ProvideConfig()

    app := fx.New(
        fxmodule.InfrastructureModule,
        // ...

        // 根据配置条件加载 WebSocket 模块
        fxmodule.WebSocketModule(cfg.WebSocket.Enable),

        fxmodule.RouterModule,
    )

    app.Run()
}


### 独立 WebSocket 服务

// cmd/websocket/main.go
func main() {
    app := fx.New(
        fxmodule.InfrastructureModule,
        fxmodule.RepositoryModule,
        fxmodule.ServiceModule,
        fxmodule.MiddlewareModule,

        // 强制启用 WebSocket
        fxmodule.WebSocketModule(true),

        fxmodule.RouterModule,
    )

    app.Run()
}
```

---

## Phase 19: 更新 CHANGELOG.md

在 CHANGELOG.md 顶部添加新版本记录：

```markdown
## [2.0.0] - 2024-XX-XX

### Changed - 重大更新

- **依赖注入框架迁移**: 从 Google Wire 迁移到 Uber fx
  - 运行时依赖解析，无需 `wire` 编译步骤
  - 类似 Spring Boot 的开发体验
  - 模块化组织 (`fx.Module`)
  - 自动生命周期管理 (`fx.Lifecycle`)

- **消除全局变量**: 移除 `global.App` 全局状态
  - 所有依赖通过构造函数注入
  - 更好的可测试性
  - 更清晰的依赖关系

### Added

- 新增 `internal/fx/` 目录，包含所有 fx 模块定义
- 支持控制器分组注入 (`group:"controllers"`)
- 条件模块加载 (`RabbitMQModule(enabled bool)`)

### Removed

- 移除 `internal/container/` 目录（Wire 相关）
- 移除 `global/` 目录（全局变量）
- 移除部分 `bootstrap/` 初始化文件（已迁移到 fx 模块）

### Migration Guide

从 v1.x 升级到 v2.0.0 需要：

1. 更新所有 `global.App.XXX` 引用为依赖注入
2. 将自定义 Provider 从 Wire 格式迁移到 fx 格式
3. 更新 Cron Job 和 Consumer 使用构造函数注入

详见 [FX_MIGRATION_PLAN.md](docs/FX_MIGRATION_PLAN.md)
```

---

## 注意事项

1. **循环依赖**: fx 会在启动时检测并报错，需要通过接口解耦
2. **懒加载**: fx 默认不是懒加载，所有服务启动时实例化
3. **测试**: 可以使用 `fxtest.New()` 进行测试
4. **调试**: 使用 `fx.WithLogger()` 查看依赖解析过程
5. **热重载**: 配置变更后需要重启应用（与 Wire 相同）

---

## 文档更新清单

| 文档 | 主要改动 |
|------|----------|
| `README.md` | 更新技术栈、项目结构、快速开始 |
| `docs/API_DEVELOPMENT.md` | 更新 DI 注册方式（Wire → fx） |
| `docs/CRON_GUIDE.md` | 移除 `global.App`，改用构造函数注入 |
| `docs/RABBITMQ_GUIDE.md` | 移除 `global.App`，改用构造函数注入 |
| `docs/MIDDLEWARE_GUIDE.md` | 更新中间件使用方式 |
| `docs/WEBSOCKET_GUIDE.md` | 更新模块化启动方式 |
| `CHANGELOG.md` | 添加 v2.0.0 版本记录 |
