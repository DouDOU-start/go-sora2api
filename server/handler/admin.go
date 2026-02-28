package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminHandler 管理端点
type AdminHandler struct {
	db        *gorm.DB
	manager   *service.AccountManager
	taskStore *service.TaskStore
	settings  *service.SettingsStore
}

// NewAdminHandler 创建管理端点
func NewAdminHandler(db *gorm.DB, manager *service.AccountManager, taskStore *service.TaskStore, settings *service.SettingsStore) *AdminHandler {
	return &AdminHandler{db: db, manager: manager, taskStore: taskStore, settings: settings}
}

// GetSettings GET /admin/settings — 获取所有设置
func (h *AdminHandler) GetSettings(c *gin.Context) {
	all := h.settings.GetAll()

	// 返回结构化的设置
	c.JSON(http.StatusOK, gin.H{
		model.SettingProxyURL:                 all[model.SettingProxyURL],
		model.SettingTokenRefreshInterval:     all[model.SettingTokenRefreshInterval],
		model.SettingCreditSyncInterval:       all[model.SettingCreditSyncInterval],
		model.SettingSubscriptionSyncInterval: all[model.SettingSubscriptionSyncInterval],
	})
}

// UpdateSettings PUT /admin/settings — 更新设置
func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 只允许更新已知的配置项
	allowedKeys := map[string]bool{
		model.SettingProxyURL:                 true,
		model.SettingTokenRefreshInterval:     true,
		model.SettingCreditSyncInterval:       true,
		model.SettingSubscriptionSyncInterval: true,
	}

	for key, value := range req {
		if !allowedKeys[key] {
			continue
		}
		h.settings.Set(key, value)
	}

	// 返回更新后的设置
	h.GetSettings(c)
}

// TestProxy POST /admin/proxy-test — 测试代理连通性
func (h *AdminHandler) TestProxy(c *gin.Context) {
	var req struct {
		ProxyURL string `json:"proxy_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	proxyURL := sora.ParseProxy(req.ProxyURL)

	client, err := sora.New(proxyURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": fmt.Sprintf("创建客户端失败: %v", err)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	start := time.Now()
	statusCode, err := client.TestConnectivity(ctx, "https://sora.chatgpt.com")
	latency := time.Since(start).Milliseconds()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": fmt.Sprintf("连接失败: %v", err), "latency": latency})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "status_code": statusCode, "latency": latency})
}
