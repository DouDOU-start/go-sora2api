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
	if err := db.AutoMigrate(&model.SoraAccountGroup{}, &model.SoraAccount{}, &model.SoraTask{}, &model.SoraSetting{}); err != nil {
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
	// 如果配置文件中有 API Keys，用于初始化默认值
	if len(os.Getenv("API_KEYS")) > 0 {
		keysJSON, _ := json.Marshal(strings.Split(os.Getenv("API_KEYS"), ","))
		defaults[model.SettingAPIKeys] = string(keysJSON)
	} else {
		defaults[model.SettingAPIKeys] = "[]"
	}
	settings.InitDefaults(defaults)

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
