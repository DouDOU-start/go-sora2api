package handler

import (
	"net/http"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
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
		model.SettingAPIKeys:                  all[model.SettingAPIKeys],
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
		model.SettingAPIKeys:                  true,
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
