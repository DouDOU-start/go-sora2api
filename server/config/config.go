package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 服务配置（仅保留启动必需项，其余动态配置存储在数据库 sora_settings 表）
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig 服务端配置
type ServerConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	AdminUser     string `yaml:"admin_user"`     // 管理员用户名
	AdminPassword string `yaml:"admin_password"` // 管理员密码
	JWTSecret     string `yaml:"jwt_secret"`     // JWT 签名密钥（可选，默认自动生成）
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	URL      string `yaml:"url"`       // PostgreSQL 连接串
	LogLevel string `yaml:"log_level"` // silent/error/warn/info
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:          "0.0.0.0",
			Port:          8686,
			AdminUser:     "admin",
			AdminPassword: "admin123",
			JWTSecret:     "sora2api-default-jwt-secret-key",
		},
		Database: DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/sora2api?sslmode=disable",
			LogLevel: "warn",
		},
	}
}

// Load 从 YAML 文件加载配置
func Load(path string) *Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[config] 配置文件 %s 不存在，使用默认配置", path)
			return cfg
		}
		log.Fatalf("[config] 读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("[config] 解析配置文件失败: %v", err)
	}

	// 补齐默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8686
	}
	if cfg.Server.AdminUser == "" {
		cfg.Server.AdminUser = "admin"
	}
	if cfg.Server.AdminPassword == "" {
		cfg.Server.AdminPassword = "admin123"
	}
	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://postgres:postgres@localhost:5432/sora2api?sslmode=disable"
	}

	log.Printf("[config] 已加载配置：%s", path)
	return cfg
}
