package model

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// ---- API 请求/响应 ----

// VideoSubmitRequest 创建任务请求（兼容 K8Ray Creator 的 SoraSubmitRequest）
type VideoSubmitRequest struct {
	Model          string `json:"model" binding:"required"`
	Prompt         string `json:"prompt" binding:"required"`
	Duration       int    `json:"duration"`
	InputReference string `json:"input_reference,omitempty"` // 图生视频参考图
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
