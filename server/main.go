package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/config"
	"github.com/DouDOU-start/go-sora2api/server/handler"
	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 配置文件路径（支持环境变量覆盖）
	configPath := "server/config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		configPath = p
	}

	cfg := config.Load(configPath)

	// 环境变量覆盖数据库连接
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.Database.URL = dbURL
	}

	// 初始化数据库
	db, err := initDB(cfg.Database)
	if err != nil {
		log.Fatalf("[main] 数据库初始化失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.SoraAccountGroup{}, &model.SoraAccount{}, &model.SoraTask{}, &model.SoraSetting{}, &model.SoraAPIKey{}); err != nil {
		log.Fatalf("[main] 数据库迁移失败: %v", err)
	}

	// 初始化设置存储（从数据库加载，首次启动写入默认值）
	settings := service.NewSettingsStore(db)
	defaults := map[string]string{
		model.SettingProxyURL:                 "",
		model.SettingTokenRefreshInterval:     "30m",
		model.SettingCreditSyncInterval:       "10m",
		model.SettingSubscriptionSyncInterval: "6h",
	}
	settings.InitDefaults(defaults)

	// 兼容迁移：将环境变量 API_KEYS 或旧 sora_settings 中的 api_keys 迁移到 sora_api_keys 表
	migrateAPIKeys(db)

	// 创建组件
	scheduler := service.NewScheduler(db, settings)
	manager := service.NewAccountManager(db, settings)
	taskStore := service.NewTaskStore(db, scheduler)

	// 启动后台同步
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.Start(ctx)

	// 恢复进行中的任务
	taskStore.RecoverInProgressTasks()

	// 设置路由
	r := handler.SetupRouter(&handler.RouterConfig{
		DB:        db,
		Scheduler: scheduler,
		TaskStore: taskStore,
		Manager:   manager,
		Settings:  settings,
		JWTSecret: cfg.Server.JWTSecret,
		AdminUser: cfg.Server.AdminUser,
		AdminPass: cfg.Server.AdminPassword,
	})

	// 前端静态文件（SPA）
	ServeWebUI(r)

	// 启动 HTTP 服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Printf("[main] 服务启动: http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] 服务启动失败: %v", err)
		}
	}()

	// 等待退出信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[main] 收到退出信号，正在关闭...")

	cancel() // 停止后台 goroutine

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] HTTP 服务关闭异常: %v", err)
	}
	log.Println("[main] 已退出")
}

// migrateAPIKeys 兼容迁移：将旧 sora_settings 中的 api_keys 和环境变量迁移到 sora_api_keys 表
func migrateAPIKeys(db *gorm.DB) {
	// 检查是否已有 API Key 记录（已迁移过则跳过）
	var count int64
	db.Model(&model.SoraAPIKey{}).Count(&count)
	if count > 0 {
		return
	}

	var keys []string

	// 优先从环境变量读取
	if envKeys := os.Getenv("API_KEYS"); envKeys != "" {
		keys = strings.Split(envKeys, ",")
	} else {
		// 尝试从旧的 sora_settings 表读取
		var setting model.SoraSetting
		if err := db.Where("key = ?", "api_keys").First(&setting).Error; err == nil {
			json.Unmarshal([]byte(setting.Value), &keys)
		}
	}

	for i, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		apiKey := model.SoraAPIKey{
			Name:    fmt.Sprintf("Key-%d", i+1),
			Key:     k,
			Enabled: true,
		}
		if err := db.Create(&apiKey).Error; err != nil {
			log.Printf("[main] 迁移 API Key 失败: %v", err)
		}
	}

	if len(keys) > 0 {
		// 清理旧设置
		db.Where("key = ?", "api_keys").Delete(&model.SoraSetting{})
		log.Printf("[main] 已将 %d 个 API Key 迁移到 sora_api_keys 表", len(keys))
	}
}

// initDB 初始化 PostgreSQL 连接（数据库不存在时自动创建）
func initDB(dbCfg config.DatabaseConfig) (*gorm.DB, error) {
	// 先用 database/sql 探测连接，不经过 GORM（避免打印预期的错误日志）
	probe, err := sql.Open("pgx", dbCfg.URL)
	if err == nil {
		err = probe.Ping()
		probe.Close()
	}

	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			if createErr := createDatabase(dbCfg.URL); createErr != nil {
				return nil, fmt.Errorf("自动创建数据库失败: %w", createErr)
			}
		} else {
			return nil, fmt.Errorf("连接数据库失败: %w", err)
		}
	}

	// 正式连接
	logLevel := parseDBLogLevel(dbCfg.LogLevel)
	db, err := gorm.Open(postgres.Open(dbCfg.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	log.Printf("[db] 数据库连接成功")
	return db, nil
}

// createDatabase 连接 postgres 默认库，创建目标数据库
func createDatabase(dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("解析数据库连接串失败: %w", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return fmt.Errorf("连接串中未指定数据库名")
	}

	// 连接 postgres 默认库
	u.Path = "/postgres"
	conn, err := sql.Open("pgx", u.String())
	if err != nil {
		return fmt.Errorf("连接 postgres 库失败: %w", err)
	}
	defer conn.Close()

	// 创建目标数据库（库名不支持参数化，此处值来自配置文件而非用户输入）
	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %q", dbName))
	if err != nil {
		return fmt.Errorf("执行 CREATE DATABASE 失败: %w", err)
	}

	log.Printf("[db] 数据库 %s 已自动创建", dbName)
	return nil
}

// parseDBLogLevel 解析日志级别
func parseDBLogLevel(level string) logger.LogLevel {
	switch strings.ToLower(level) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
