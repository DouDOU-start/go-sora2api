package service

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"gorm.io/gorm"
)

var ErrNoAvailableAccount = errors.New("没有可用的 Sora 账号")

// Scheduler 账号调度器
type Scheduler struct {
	db       *gorm.DB
	mu       sync.Mutex
	settings *SettingsStore
}

// NewScheduler 创建调度器
func NewScheduler(db *gorm.DB, settings *SettingsStore) *Scheduler {
	return &Scheduler{db: db, settings: settings}
}

// PickAccount 选取一个可用账号（最久未用优先）
//
// 筛选条件：
//   - enabled=true 且 status=active
//   - remaining_count != 0（-1=未知视为可用，0=额度用完排除）
//   - rate_limit_reached=false 或 rate_limit_resets_at < now()（限流已解除）
//
// 排序：last_used_at ASC NULLS FIRST
func (s *Scheduler) PickAccount() (*model.SoraAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var account model.SoraAccount

	err := s.db.
		Where("enabled = ? AND status = ?", true, model.AccountStatusActive).
		Where("remaining_count != 0"). // -1(未知) 或 >0 均可用
		Where("rate_limit_reached = ? OR rate_limit_resets_at < ?", false, now).
		Order("last_used_at ASC NULLS FIRST").
		First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoAvailableAccount
		}
		return nil, err
	}

	// 更新最后使用时间
	s.db.Model(&account).Update("last_used_at", now)

	return &account, nil
}

// MarkAccountError 标记账号错误状态
func (s *Scheduler) MarkAccountError(accountID int64, status, lastError string) {
	if err := s.db.Model(&model.SoraAccount{}).Where("id = ?", accountID).
		Updates(map[string]interface{}{
			"status":     status,
			"last_error": lastError,
		}).Error; err != nil {
		log.Printf("[scheduler] 更新账号 %d 状态失败: %v", accountID, err)
	}
}

// MarkRateLimited 标记账号限流
func (s *Scheduler) MarkRateLimited(accountID int64, resetsInSec int) {
	resetsAt := time.Now().Add(time.Duration(resetsInSec) * time.Second)
	if err := s.db.Model(&model.SoraAccount{}).Where("id = ?", accountID).
		Updates(map[string]interface{}{
			"rate_limit_reached":  true,
			"rate_limit_resets_at": resetsAt,
		}).Error; err != nil {
		log.Printf("[scheduler] 标记账号 %d 限流失败: %v", accountID, err)
	}
}

// GetProxyURL 返回全局代理 URL（动态从设置读取）
func (s *Scheduler) GetProxyURL() string {
	return s.settings.GetProxyURL()
}
