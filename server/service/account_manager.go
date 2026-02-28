package service

import (
	"context"
	"log"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/sora"
	"gorm.io/gorm"
)

// SyncConfig 同步配置
type SyncConfig struct {
	TokenRefreshInterval     time.Duration
	CreditSyncInterval       time.Duration
	SubscriptionSyncInterval time.Duration
}

// AccountManager 账号池管理（Token 刷新、配额同步、订阅同步）
type AccountManager struct {
	db       *gorm.DB
	settings *SettingsStore
}

// NewAccountManager 创建账号管理器
func NewAccountManager(db *gorm.DB, settings *SettingsStore) *AccountManager {
	return &AccountManager{db: db, settings: settings}
}

// proxyURL 获取当前代理 URL
func (am *AccountManager) proxyURL() string {
	return am.settings.GetProxyURL()
}

// Start 启动后台同步任务
func (am *AccountManager) Start(ctx context.Context) {
	cfg := am.settings.GetSyncConfig()
	go am.tokenRefreshLoop(ctx)
	go am.creditSyncLoop(ctx)
	go am.subscriptionSyncLoop(ctx)
	log.Printf("[account_manager] 后台同步已启动（Token: %v, 配额: %v, 订阅: %v）",
		cfg.TokenRefreshInterval, cfg.CreditSyncInterval, cfg.SubscriptionSyncInterval)
}

// tokenRefreshLoop Token 刷新循环
func (am *AccountManager) tokenRefreshLoop(ctx context.Context) {
	am.refreshAllTokens(ctx)

	for {
		cfg := am.settings.GetSyncConfig()
		select {
		case <-ctx.Done():
			return
		case <-time.After(cfg.TokenRefreshInterval):
			am.refreshAllTokens(ctx)
		}
	}
}

// refreshAllTokens 刷新所有有 RT 的账号
func (am *AccountManager) refreshAllTokens(ctx context.Context) {
	var accounts []model.SoraAccount
	if err := am.db.Where("enabled = ? AND refresh_token != ''", true).Find(&accounts).Error; err != nil {
		log.Printf("[token_refresh] 查询账号失败: %v", err)
		return
	}

	if len(accounts) == 0 {
		return
	}

	log.Printf("[token_refresh] 开始刷新 %d 个账号的 Token", len(accounts))
	success, fail := 0, 0

	for i := range accounts {
		acc := &accounts[i]
		if err := am.refreshAccountToken(ctx, acc); err != nil {
			fail++
			log.Printf("[token_refresh] 账号 %d(%s) 刷新失败: %v", acc.ID, acc.Name, err)
		} else {
			success++
		}
	}

	log.Printf("[token_refresh] 刷新完成：成功 %d，失败 %d", success, fail)
}

// refreshAccountToken 刷新单个账号的 Token
func (am *AccountManager) refreshAccountToken(ctx context.Context, acc *model.SoraAccount) error {
	client, err := sora.New(am.proxyURL())
	if err != nil {
		am.markError(acc.ID, model.AccountStatusTokenExpired, err.Error())
		return err
	}

	newAT, newRT, err := client.RefreshAccessToken(ctx, acc.RefreshToken, "")
	if err != nil {
		am.markError(acc.ID, model.AccountStatusTokenExpired, err.Error())
		return err
	}

	// 回写到内存对象，确保调用方能拿到最新的 Token
	acc.AccessToken = newAT
	acc.RefreshToken = newRT

	updates := map[string]interface{}{
		"access_token":  newAT,
		"refresh_token": newRT,
		"last_sync_at":  time.Now(),
	}

	// 从新 AT 提取邮箱（如果之前未获取到）
	if acc.Email == "" {
		if email := model.ExtractEmailFromJWT(newAT); email != "" {
			updates["email"] = email
			acc.Email = email
		}
	}

	if acc.Status == model.AccountStatusTokenExpired {
		updates["status"] = model.AccountStatusActive
		updates["last_error"] = ""
	}

	return am.db.Model(&model.SoraAccount{}).Where("id = ?", acc.ID).Updates(updates).Error
}

// creditSyncLoop 配额同步循环
func (am *AccountManager) creditSyncLoop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}
	am.syncAllCredits(ctx)

	for {
		cfg := am.settings.GetSyncConfig()
		select {
		case <-ctx.Done():
			return
		case <-time.After(cfg.CreditSyncInterval):
			am.syncAllCredits(ctx)
		}
	}
}

// syncAllCredits 同步所有账号的配额
func (am *AccountManager) syncAllCredits(ctx context.Context) {
	var accounts []model.SoraAccount
	if err := am.db.Where("enabled = ? AND status IN ?", true,
		[]string{model.AccountStatusActive, model.AccountStatusQuotaExhausted}).Find(&accounts).Error; err != nil {
		log.Printf("[credit_sync] 查询账号失败: %v", err)
		return
	}

	for i := range accounts {
		am.syncAccountCredit(ctx, &accounts[i])
	}
}

// syncAccountCredit 同步单个账号配额
func (am *AccountManager) syncAccountCredit(ctx context.Context, acc *model.SoraAccount) {
	client, err := sora.New(am.proxyURL())
	if err != nil {
		return
	}

	balance, err := client.GetCreditBalance(ctx, acc.AccessToken)
	if err != nil {
		log.Printf("[credit_sync] 账号 %d(%s) 配额查询失败: %v", acc.ID, acc.Name, err)
		return
	}

	updates := map[string]interface{}{
		"remaining_count":    balance.RemainingCount,
		"rate_limit_reached": balance.RateLimitReached,
		"last_sync_at":       time.Now(),
	}

	// 补全邮箱：优先从 /me API 获取，回退到 JWT 解析
	if acc.Email == "" {
		if userInfo, err := client.GetUserInfo(ctx, acc.AccessToken); err == nil && userInfo.Email != "" {
			updates["email"] = userInfo.Email
			if acc.Name == "" && userInfo.Name != "" {
				updates["name"] = userInfo.Name
			}
		} else if email := model.ExtractEmailFromJWT(acc.AccessToken); email != "" {
			updates["email"] = email
		}
	}

	if balance.RateLimitReached && balance.AccessResetsInSec > 0 {
		resetsAt := time.Now().Add(time.Duration(balance.AccessResetsInSec) * time.Second)
		updates["rate_limit_resets_at"] = resetsAt
	} else {
		updates["rate_limit_reached"] = false
	}

	if balance.RemainingCount == 0 {
		updates["status"] = model.AccountStatusQuotaExhausted
	} else if acc.Status == model.AccountStatusQuotaExhausted && balance.RemainingCount != 0 {
		updates["status"] = model.AccountStatusActive
		updates["last_error"] = ""
		log.Printf("[credit_sync] 账号 %d(%s) 额度已恢复，重新启用", acc.ID, acc.Name)
	}

	am.db.Model(&model.SoraAccount{}).Where("id = ?", acc.ID).Updates(updates)
}

// subscriptionSyncLoop 订阅同步循环
func (am *AccountManager) subscriptionSyncLoop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(1 * time.Minute):
	}
	am.syncAllSubscriptions(ctx)

	for {
		cfg := am.settings.GetSyncConfig()
		select {
		case <-ctx.Done():
			return
		case <-time.After(cfg.SubscriptionSyncInterval):
			am.syncAllSubscriptions(ctx)
		}
	}
}

// syncAllSubscriptions 同步所有账号的订阅信息
func (am *AccountManager) syncAllSubscriptions(ctx context.Context) {
	var accounts []model.SoraAccount
	if err := am.db.Where("enabled = ?", true).Find(&accounts).Error; err != nil {
		log.Printf("[sub_sync] 查询账号失败: %v", err)
		return
	}

	for i := range accounts {
		am.syncAccountSubscription(ctx, &accounts[i])
	}
}

// syncAccountSubscription 同步单个账号订阅信息
func (am *AccountManager) syncAccountSubscription(ctx context.Context, acc *model.SoraAccount) {
	client, err := sora.New(am.proxyURL())
	if err != nil {
		return
	}

	info, err := client.GetSubscriptionInfo(ctx, acc.AccessToken)
	if err != nil {
		log.Printf("[sub_sync] 账号 %d(%s) 订阅查询失败: %v", acc.ID, acc.Name, err)
		return
	}

	updates := map[string]interface{}{
		"plan_title": info.PlanTitle,
	}
	if info.EndTs > 0 {
		expiresAt := time.Unix(info.EndTs, 0)
		updates["plan_expires_at"] = expiresAt
		if expiresAt.Before(time.Now()) {
			log.Printf("[sub_sync] ⚠ 账号 %d(%s) 订阅已过期: %s（%s）", acc.ID, acc.Name, info.PlanTitle, expiresAt.Format("2006-01-02"))
		}
	}

	am.db.Model(&model.SoraAccount{}).Where("id = ?", acc.ID).Updates(updates)
}

// SyncSingleAccountCredit 手动同步单个账号配额（管理端点使用）
func (am *AccountManager) SyncSingleAccountCredit(ctx context.Context, acc *model.SoraAccount) error {
	client, err := sora.New(am.proxyURL())
	if err != nil {
		return err
	}

	balance, err := client.GetCreditBalance(ctx, acc.AccessToken)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"remaining_count":    balance.RemainingCount,
		"rate_limit_reached": balance.RateLimitReached,
		"last_sync_at":       time.Now(),
	}

	// 补全邮箱
	if acc.Email == "" {
		if userInfo, err := client.GetUserInfo(ctx, acc.AccessToken); err == nil && userInfo.Email != "" {
			updates["email"] = userInfo.Email
			if acc.Name == "" && userInfo.Name != "" {
				updates["name"] = userInfo.Name
			}
		} else if email := model.ExtractEmailFromJWT(acc.AccessToken); email != "" {
			updates["email"] = email
		}
	}

	if balance.RateLimitReached && balance.AccessResetsInSec > 0 {
		resetsAt := time.Now().Add(time.Duration(balance.AccessResetsInSec) * time.Second)
		updates["rate_limit_resets_at"] = resetsAt
	}
	if balance.RemainingCount == 0 {
		updates["status"] = model.AccountStatusQuotaExhausted
	} else if acc.Status == model.AccountStatusQuotaExhausted {
		updates["status"] = model.AccountStatusActive
		updates["last_error"] = ""
	}

	return am.db.Model(&model.SoraAccount{}).Where("id = ?", acc.ID).Updates(updates).Error
}

// SyncSingleAccountSubscription 手动同步单个账号订阅（管理端点使用）
func (am *AccountManager) SyncSingleAccountSubscription(ctx context.Context, acc *model.SoraAccount) error {
	client, err := sora.New(am.proxyURL())
	if err != nil {
		return err
	}

	info, err := client.GetSubscriptionInfo(ctx, acc.AccessToken)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"plan_title":   info.PlanTitle,
		"last_sync_at": time.Now(),
	}
	if info.EndTs > 0 {
		expiresAt := time.Unix(info.EndTs, 0)
		updates["plan_expires_at"] = expiresAt
	}

	return am.db.Model(&model.SoraAccount{}).Where("id = ?", acc.ID).Updates(updates).Error
}

// RefreshSingleToken 手动刷新单个账号 Token（管理端点使用）
func (am *AccountManager) RefreshSingleToken(ctx context.Context, acc *model.SoraAccount) error {
	if acc.RefreshToken == "" {
		return nil
	}
	return am.refreshAccountToken(ctx, acc)
}

// markError 标记账号错误
func (am *AccountManager) markError(accountID int64, status, lastError string) {
	am.db.Model(&model.SoraAccount{}).Where("id = ?", accountID).
		Updates(map[string]interface{}{
			"status":     status,
			"last_error": lastError,
		})
}
