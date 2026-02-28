package model

import "time"

// ---- 数据库模型 ----

// SoraAccountGroup 账号组
type SoraAccountGroup struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:128;not null;uniqueIndex"`
	Description string    `json:"description" gorm:"size:512"`
	Enabled     bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SoraAccountGroup) TableName() string { return "sora_account_groups" }

// SoraAccount Sora 账号
type SoraAccount struct {
	ID                int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupID           *int64     `json:"group_id" gorm:"index"`
	Name              string     `json:"name" gorm:"size:128"`
	Email             string     `json:"email" gorm:"size:256"`            // 从 JWT 自动提取
	AccessToken       string     `json:"-" gorm:"type:text;not null"`      // 不对外暴露
	RefreshToken      string     `json:"-" gorm:"type:text"`               // 不对外暴露
	TokenExpiresAt    *time.Time `json:"token_expires_at"`
	PlanTitle         string     `json:"plan_title" gorm:"size:64"`
	PlanExpiresAt     *time.Time `json:"plan_expires_at"`
	RemainingCount    int        `json:"remaining_count" gorm:"default:-1"` // -1=未知
	RateLimitReached  bool       `json:"rate_limit_reached" gorm:"default:false"`
	RateLimitResetsAt *time.Time `json:"rate_limit_resets_at"`
	Enabled           bool       `json:"enabled" gorm:"not null;default:true"`
	Status            string     `json:"status" gorm:"size:32;default:active"` // active/token_expired/quota_exhausted
	LastUsedAt        *time.Time `json:"last_used_at"`
	LastError         string     `json:"last_error" gorm:"type:text"`
	LastSyncAt        *time.Time `json:"last_sync_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SoraAccount) TableName() string { return "sora_accounts" }

// SoraTask 内部任务记录
type SoraTask struct {
	ID           string     `json:"id" gorm:"primaryKey;size:64"`
	SoraTaskID   string     `json:"sora_task_id" gorm:"size:128;not null;index"`
	AccountID    int64      `json:"account_id" gorm:"not null;index"`
	APIKeyID     int64      `json:"api_key_id" gorm:"index;default:0"` // 创建该任务的 API Key ID（0 表示未知）
	Type         string     `json:"type" gorm:"size:32;not null;default:video"` // video/image
	Model        string     `json:"model" gorm:"size:128"`
	Prompt       string     `json:"prompt" gorm:"type:text"`
	Status       string     `json:"status" gorm:"size:32;not null;index;default:queued"` // queued/in_progress/completed/failed
	Progress     int        `json:"progress" gorm:"default:0"`
	ErrorMessage string     `json:"error_message,omitempty" gorm:"type:text"`
	DownloadURL  string     `json:"-" gorm:"size:1024"`                  // 完成后的下载链接（内部使用）
	ImageURL     string     `json:"image_url,omitempty" gorm:"size:1024"` // 图片任务结果
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

func (SoraTask) TableName() string { return "sora_tasks" }

// ---- 状态常量 ----

// 账号状态
const (
	AccountStatusActive         = "active"
	AccountStatusTokenExpired   = "token_expired"
	AccountStatusQuotaExhausted = "quota_exhausted"
)

// 任务状态
const (
	TaskStatusQueued     = "queued"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// SoraCharacter 角色记录
type SoraCharacter struct {
	ID           string     `json:"id" gorm:"primaryKey;size:64"`                              // 内部 ID: char_xxxxxxxx
	AccountID    int64      `json:"account_id" gorm:"not null;index"`
	CameoID      string     `json:"cameo_id" gorm:"size:128;index"`                            // Sora cameo ID
	CharacterID  string     `json:"character_id" gorm:"size:128;index"`                        // 定稿后的 character ID
	Status       string     `json:"status" gorm:"size:32;not null;default:processing"`         // processing/ready/failed
	DisplayName  string     `json:"display_name" gorm:"size:128"`
	Username     string     `json:"username" gorm:"size:128"`
	ProfileURL   string     `json:"profile_url" gorm:"size:1024"`
	ProfileImage []byte     `json:"-" gorm:"type:bytea"`                        // 头像图片二进制数据（不对外暴露）
	IsPublic     bool       `json:"is_public" gorm:"default:false"`             // 是否公开
	ErrorMessage string     `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

func (SoraCharacter) TableName() string { return "sora_characters" }

// 角色状态
const (
	CharacterStatusProcessing = "processing"
	CharacterStatusReady      = "ready"
	CharacterStatusFailed     = "failed"
)

// SoraAPIKey API 密钥（独立管理，可绑定分组）
type SoraAPIKey struct {
	ID         int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string     `json:"name" gorm:"size:128;not null"`
	Key        string     `json:"key" gorm:"size:256;not null;uniqueIndex"`
	GroupID    *int64     `json:"group_id" gorm:"index"`
	Enabled    bool       `json:"enabled" gorm:"not null;default:true"`
	UsageCount int64      `json:"usage_count" gorm:"default:0"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SoraAPIKey) TableName() string { return "sora_api_keys" }

// SoraSetting KV 配置项（存储动态配置）
type SoraSetting struct {
	Key       string    `json:"key" gorm:"primaryKey;size:64"`
	Value     string    `json:"value" gorm:"type:text;not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SoraSetting) TableName() string { return "sora_settings" }

// 配置项 Key 常量
const (
	SettingProxyURL                 = "proxy_url"                  // 字符串
	SettingTokenRefreshInterval     = "token_refresh_interval"     // Duration 字符串
	SettingCreditSyncInterval       = "credit_sync_interval"       // Duration 字符串
	SettingSubscriptionSyncInterval = "subscription_sync_interval" // Duration 字符串
)
