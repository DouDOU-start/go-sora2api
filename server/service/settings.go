package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SettingsStore 系统设置存储（内存缓存 + 数据库持久化）
type SettingsStore struct {
	db    *gorm.DB
	mu    sync.RWMutex
	cache map[string]string
}

// NewSettingsStore 创建设置存储
func NewSettingsStore(db *gorm.DB) *SettingsStore {
	s := &SettingsStore{db: db, cache: make(map[string]string)}
	s.loadAll()
	return s
}

// InitDefaults 初始化默认值（仅在设置不存在时写入）
func (s *SettingsStore) InitDefaults(defaults map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, val := range defaults {
		if _, exists := s.cache[key]; !exists {
			s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SoraSetting{
				Key:   key,
				Value: val,
			})
			s.cache[key] = val
		}
	}
}

// Get 获取设置值
func (s *SettingsStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key]
}

// Set 设置值
func (s *SettingsStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&model.SoraSetting{
		Key:   key,
		Value: value,
	})

	s.cache[key] = value
}

// GetAll 获取所有设置（返回副本）
func (s *SettingsStore) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string, len(s.cache))
	for k, v := range s.cache {
		result[k] = v
	}
	return result
}

// GetAPIKeys 获取 API Keys 列表
func (s *SettingsStore) GetAPIKeys() []string {
	raw := s.Get(model.SettingAPIKeys)
	if raw == "" {
		return nil
	}
	var keys []string
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return nil
	}
	return keys
}

// GetProxyURL 获取全局代理 URL
func (s *SettingsStore) GetProxyURL() string {
	return s.Get(model.SettingProxyURL)
}

// GetSyncConfig 获取同步配置
func (s *SettingsStore) GetSyncConfig() *SyncConfig {
	cfg := &SyncConfig{
		TokenRefreshInterval:     30 * time.Minute,
		CreditSyncInterval:       10 * time.Minute,
		SubscriptionSyncInterval: 6 * time.Hour,
	}

	if v := s.Get(model.SettingTokenRefreshInterval); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.TokenRefreshInterval = d
		}
	}
	if v := s.Get(model.SettingCreditSyncInterval); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CreditSyncInterval = d
		}
	}
	if v := s.Get(model.SettingSubscriptionSyncInterval); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.SubscriptionSyncInterval = d
		}
	}
	return cfg
}

// loadAll 从数据库加载所有设置到缓存
func (s *SettingsStore) loadAll() {
	var settings []model.SoraSetting
	if err := s.db.Find(&settings).Error; err != nil {
		log.Printf("[settings] 加载设置失败: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range settings {
		s.cache[item.Key] = item.Value
	}
}
