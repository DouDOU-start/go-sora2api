package model

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// ---- API 请求/响应 ----

// VideoSubmitRequest 创建视频任务请求
type VideoSubmitRequest struct {
	Model          string `json:"model" binding:"required"`
	Prompt         string `json:"prompt" binding:"required"`
	Duration       int    `json:"duration"`
	InputReference string `json:"input_reference,omitempty"` // 图生视频参考图（URL 或 base64 data URI）
	Style          string `json:"style,omitempty"`           // 视频风格（如 anime, retro 等）
}

// RemixSubmitRequest Remix 视频请求
type RemixSubmitRequest struct {
	Model       string `json:"model" binding:"required"`
	Prompt      string `json:"prompt" binding:"required"`
	RemixTarget string `json:"remix_target" binding:"required"` // Sora 分享链接或 s_xxx 格式 ID
	Style       string `json:"style,omitempty"`
}

// StoryboardSubmitRequest 分镜视频请求
type StoryboardSubmitRequest struct {
	Model          string `json:"model" binding:"required"`
	Prompt         string `json:"prompt" binding:"required"`         // 分镜格式: [5.0s]场景1 [5.0s]场景2
	InputReference string `json:"input_reference,omitempty"`         // 参考图（URL 或 base64 data URI）
	Style          string `json:"style,omitempty"`
}

// VideoTaskResponse 任务响应（兼容 K8Ray Creator 的 SoraTaskResponse）
type VideoTaskResponse struct {
	ID        string         `json:"id"`
	Object    string         `json:"object"`
	Model     string         `json:"model"`
	Status    string         `json:"status"`
	Progress  int            `json:"progress"`
	CreatedAt int64          `json:"created_at"`
	Size      string         `json:"size,omitempty"`
	Error     *TaskErrorInfo `json:"error,omitempty"`
}

// TaskErrorInfo 任务错误信息
type TaskErrorInfo struct {
	Message string `json:"message"`
}

// ---- 图片任务 ----

// ImageSubmitRequest 创建图片任务请求
type ImageSubmitRequest struct {
	Prompt         string `json:"prompt" binding:"required"`
	Width          int    `json:"width"`                       // 默认 1792
	Height         int    `json:"height"`                      // 默认 1024
	InputReference string `json:"input_reference,omitempty"`   // 图生图参考图（URL 或 base64 data URI）
}

// ImageTaskResponse 图片任务响应
type ImageTaskResponse struct {
	ID        string         `json:"id"`
	Object    string         `json:"object"`             // "image"
	Status    string         `json:"status"`
	Progress  int            `json:"progress"`
	CreatedAt int64          `json:"created_at"`
	Width     int            `json:"width,omitempty"`
	Height    int            `json:"height,omitempty"`
	ImageURL  string         `json:"image_url,omitempty"`
	Error     *TaskErrorInfo `json:"error,omitempty"`
}

// ---- 角色管理 ----

// CharacterCreateRequest 创建角色请求
type CharacterCreateRequest struct {
	VideoURL    string `json:"video_url" binding:"required"` // 角色视频（URL 或 base64 data URI）
	Username    string `json:"username,omitempty"`            // 可选，不传则使用推荐值
	DisplayName string `json:"display_name,omitempty"`       // 可选，不传则使用推荐值
}

// CharacterResponse 角色响应
type CharacterResponse struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	DisplayName string         `json:"display_name,omitempty"`
	Username    string         `json:"username,omitempty"`
	ProfileURL  string         `json:"profile_url,omitempty"`
	CharacterID string         `json:"character_id,omitempty"` // 定稿后可用于视频生成
	CreatedAt   int64          `json:"created_at"`
	Error       *TaskErrorInfo `json:"error,omitempty"`
}

// ---- 提示词增强 ----

// EnhancePromptRequest 提示词增强请求
type EnhancePromptRequest struct {
	Prompt         string `json:"prompt" binding:"required"`
	ExpansionLevel string `json:"expansion_level,omitempty"` // "medium" 或 "long"，默认 "medium"
	Duration       int    `json:"duration,omitempty"`        // 5/10/15/25 秒，默认 10
}

// EnhancePromptResponse 提示词增强响应
type EnhancePromptResponse struct {
	OriginalPrompt string `json:"original_prompt"`
	EnhancedPrompt string `json:"enhanced_prompt"`
}

// ---- 帖子管理 ----

// PostCreateRequest 发布帖子请求
type PostCreateRequest struct {
	TaskID string `json:"task_id" binding:"required"` // 内部视频任务 ID
}

// PostResponse 帖子响应
type PostResponse struct {
	PostID string `json:"post_id"`
}

// ---- 无水印下载 ----

// WatermarkFreeRequest 无水印下载请求
type WatermarkFreeRequest struct {
	VideoID string `json:"video_id" binding:"required"` // Sora 分享链接或视频 ID
}

// WatermarkFreeResponse 无水印下载响应
type WatermarkFreeResponse struct {
	URL string `json:"url"`
}

// ---- 管理端点请求/响应 ----

// AdminGroupRequest 账号组创建/编辑请求
type AdminGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

// AdminAccountRequest 账号创建/编辑请求
type AdminAccountRequest struct {
	Name         string `json:"name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	GroupID      *int64 `json:"group_id"`
	Enabled      *bool  `json:"enabled"`
}

// AdminAccountResponse 账号响应（含 Token 掩码）
type AdminAccountResponse struct {
	SoraAccount
	ATHint    string `json:"at_hint"`              // AT 掩码
	RTHint    string `json:"rt_hint"`              // RT 掩码
	GroupName string `json:"group_name,omitempty"` // 所属分组名称
}

// DashboardStats 概览统计
type DashboardStats struct {
	TotalAccounts     int64 `json:"total_accounts"`
	ActiveAccounts    int64 `json:"active_accounts"`
	ExpiredAccounts   int64 `json:"expired_accounts"`
	ExhaustedAccounts int64 `json:"exhausted_accounts"`
	TotalTasks        int64 `json:"total_tasks"`
	PendingTasks      int64 `json:"pending_tasks"`
	CompletedTasks    int64 `json:"completed_tasks"`
	FailedTasks       int64 `json:"failed_tasks"`
	TotalCharacters      int64 `json:"total_characters"`
	ReadyCharacters      int64 `json:"ready_characters"`
	ProcessingCharacters int64 `json:"processing_characters"`
	FailedCharacters     int64 `json:"failed_characters"`
}

// AdminAPIKeyRequest API Key 创建/编辑请求
type AdminAPIKeyRequest struct {
	Name    string `json:"name" binding:"required"`
	Key     string `json:"key"`     // 创建时必填，编辑时可选（不传则不修改）
	GroupID *int64 `json:"group_id"`
	Enabled *bool  `json:"enabled"`
}

// AdminAPIKeyResponse API Key 响应（含分组名和 Key 掩码）
type AdminAPIKeyResponse struct {
	SoraAPIKey
	KeyHint   string `json:"key_hint"`             // Key 掩码
	GroupName string `json:"group_name,omitempty"` // 所属分组名称
}

// AdminBatchImportRequest 批量导入账号请求
type AdminBatchImportRequest struct {
	Tokens  []string `json:"tokens" binding:"required"`
	GroupID *int64   `json:"group_id"`
}

// AdminBatchImportItemResult 单个 Token 导入结果
type AdminBatchImportItemResult struct {
	Token  string `json:"token"`           // Token 掩码
	Action string `json:"action"`          // "created" / "updated" / "failed"
	Email  string `json:"email,omitempty"` // 识别出的邮箱
	Error  string `json:"error,omitempty"` // 错误信息
}

// AdminBatchImportResult 批量导入汇总结果
type AdminBatchImportResult struct {
	Total   int                          `json:"total"`
	Created int                          `json:"created"`
	Updated int                          `json:"updated"`
	Failed  int                          `json:"failed"`
	Details []AdminBatchImportItemResult `json:"details"`
}

// AdminCharacterResponse 角色管理响应（含关联账号邮箱）
type AdminCharacterResponse struct {
	SoraCharacter
	AccountEmail string `json:"account_email,omitempty"` // 关联账号邮箱
}

// ---- 工具函数 ----

// ExtractEmailFromJWT 从 JWT Access Token 的 payload 中提取邮箱
// JWT 格式为 header.payload.signature，payload 是 base64url 编码的 JSON
func ExtractEmailFromJWT(token string) string {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	// base64url 解码 payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	// 尝试常见的邮箱字段名
	for _, key := range []string{"email", "https://api.openai.com/profile", "https://api.openai.com/auth", "preferred_username", "sub"} {
		if v, ok := claims[key]; ok {
			if s, ok := v.(string); ok && strings.Contains(s, "@") {
				return s
			}
			// OpenAI JWT 的 profile/auth 字段可能是嵌套对象
			if m, ok := v.(map[string]interface{}); ok {
				if email, ok := m["email"].(string); ok && email != "" {
					return email
				}
			}
		}
	}
	return ""
}

// MaskToken 生成 Token 掩码
func MaskToken(token string) string {
	if len(token) <= 8 {
		if len(token) > 0 {
			return "****"
		}
		return ""
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// MaskURL 生成 URL 掩码
func MaskURL(rawURL string) string {
	if len(rawURL) <= 10 {
		return "****"
	}
	return rawURL[:8] + "****" + rawURL[len(rawURL)-4:]
}
