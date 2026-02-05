package fx

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gin-web/app/models"
	"gin-web/config"
	"gin-web/utils"
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
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg config.Configuration
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 热重载
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config file changed:", in.Name)
		if err := v.Unmarshal(&cfg); err != nil {
			fmt.Println("reload config failed:", err)
		}
	})

	return &cfg, nil
}

// ProvideLogger 提供日志器
func ProvideLogger(cfg *config.Configuration) (*zap.Logger, error) {
	// 创建日志目录
	if ok, _ := utils.PathExists(cfg.Log.RootDir); !ok {
		if err := os.MkdirAll(cfg.Log.RootDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("create log dir failed: %w", err)
		}
	}

	// 设置日志等级
	var level zapcore.Level
	var options []zap.Option

	switch cfg.Log.Level {
	case "debug":
		level = zap.DebugLevel
		options = append(options, zap.AddStacktrace(level))
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
		options = append(options, zap.AddStacktrace(level))
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	if cfg.Log.ShowLine {
		options = append(options, zap.AddCaller())
	}

	// 编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(t.Format("[2006-01-02 15:04:05.000]"))
	}
	encoderConfig.EncodeLevel = func(l zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(cfg.App.Env + "." + l.String())
	}

	var encoder zapcore.Encoder
	if cfg.Log.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 日志写入器（文件）
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Log.RootDir + "/" + cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}

	// 同时输出到文件和控制台
	multiWriter := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(fileWriter),
		zapcore.AddSync(os.Stdout),
	)

	core := zapcore.NewCore(encoder, multiWriter, level)
	return zap.New(core, options...), nil
}

// ProvideDatabase 提供数据库连接
func ProvideDatabase(lc fx.Lifecycle, cfg *config.Configuration, log *zap.Logger) (*gorm.DB, error) {
	dbConfig := cfg.Database

	if dbConfig.Database == "" {
		log.Warn("database not configured, skipping")
		return nil, nil
	}

	// 构建 DSN
	dsn := dbConfig.UserName + ":" + dbConfig.Password + "@tcp(" + dbConfig.Host + ":" + strconv.Itoa(dbConfig.Port) + ")/" +
		dbConfig.Database + "?charset=" + dbConfig.Charset + "&parseTime=True&loc=Local"

	// GORM 日志配置
	gormLogger := newGormLogger(cfg)

	mysqlConfig := mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         191,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}

	db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: dbConfig.Prefix,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("connect database failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)

	// 自动迁移表结构
	if err := db.AutoMigrate(
		models.User{},
		models.Game{},
		models.Category{},
		models.Mod{},
	); err != nil {
		return nil, fmt.Errorf("migrate table failed: %w", err)
	}

	// 生命周期管理
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("database connection established",
				zap.String("host", dbConfig.Host),
				zap.Int("port", dbConfig.Port),
				zap.String("database", dbConfig.Database),
			)
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing database connection")
			return sqlDB.Close()
		},
	})

	return db, nil
}

// ProvideRedis 提供 Redis 连接
func ProvideRedis(lc fx.Lifecycle, cfg *config.Configuration, log *zap.Logger) (*redis.Client, error) {
	redisConfig := cfg.Redis

	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if _, err := client.Ping(ctx).Result(); err != nil {
				return fmt.Errorf("redis connect failed: %w", err)
			}
			log.Info("redis connection established",
				zap.String("host", redisConfig.Host),
				zap.Int("port", redisConfig.Port),
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing redis connection")
			return client.Close()
		},
	})

	return client, nil
}

// newGormLogger 创建 GORM 日志器
func newGormLogger(cfg *config.Configuration) logger.Interface {
	var writer io.Writer

	if cfg.Database.EnableFileLogWriter {
		writer = &lumberjack.Logger{
			Filename:   cfg.Log.RootDir + "/" + cfg.Database.LogFilename,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		}
	} else {
		writer = os.Stdout
	}

	var logMode logger.LogLevel
	switch cfg.Database.LogMode {
	case "silent":
		logMode = logger.Silent
	case "error":
		logMode = logger.Error
	case "warn":
		logMode = logger.Warn
	case "info":
		logMode = logger.Info
	default:
		logMode = logger.Info
	}

	return logger.New(log.New(writer, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logMode,
		IgnoreRecordNotFoundError: false,
		Colorful:                  !cfg.Database.EnableFileLogWriter,
	})
}
