package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/gin-gonic/gin"
)

// ListAPIKeys GET /admin/api-keys
func (h *AdminHandler) ListAPIKeys(c *gin.Context) {
	var keys []model.SoraAPIKey
	h.db.Order("id ASC").Find(&keys)

	// 构建分组 ID → 名称映射
	groupIDs := make([]int64, 0)
	for _, k := range keys {
		if k.GroupID != nil {
			groupIDs = append(groupIDs, *k.GroupID)
		}
	}
	groupMap := make(map[int64]string)
	if len(groupIDs) > 0 {
		var groups []model.SoraAccountGroup
		h.db.Where("id IN ?", groupIDs).Find(&groups)
		for _, g := range groups {
			groupMap[g.ID] = g.Name
		}
	}

	resp := make([]model.AdminAPIKeyResponse, 0, len(keys))
	for _, k := range keys {
		item := model.AdminAPIKeyResponse{
			SoraAPIKey: k,
			KeyHint:    model.MaskToken(k.Key),
		}
		if k.GroupID != nil {
			item.GroupName = groupMap[*k.GroupID]
		}
		// 不暴露完整 Key
		item.SoraAPIKey.Key = ""
		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, resp)
}

// CreateAPIKey POST /admin/api-keys
func (h *AdminHandler) CreateAPIKey(c *gin.Context) {
	var req model.AdminAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := req.Key
	if key == "" {
		// 自动生成 sk- 前缀的随机 Key
		key = generateAPIKey()
	}

	apiKey := model.SoraAPIKey{
		Name:    req.Name,
		Key:     key,
		GroupID:  req.GroupID,
		Enabled: true,
	}
	if req.Enabled != nil {
		apiKey.Enabled = *req.Enabled
	}

	if err := h.db.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("创建 API Key 失败: %v", err)})
		return
	}

	// 返回时显示完整 Key（仅创建时可见）
	c.JSON(http.StatusCreated, apiKey)
}

// UpdateAPIKey PUT /admin/api-keys/:id
func (h *AdminHandler) UpdateAPIKey(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var apiKey model.SoraAPIKey
	if err := h.db.First(&apiKey, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API Key 不存在"})
		return
	}

	var req model.AdminAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey.Name = req.Name
	apiKey.GroupID = req.GroupID
	if req.Key != "" {
		apiKey.Key = req.Key
	}
	if req.Enabled != nil {
		apiKey.Enabled = *req.Enabled
	}

	if err := h.db.Save(&apiKey).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("更新 API Key 失败: %v", err)})
		return
	}

	// 返回掩码
	resp := model.AdminAPIKeyResponse{
		SoraAPIKey: apiKey,
		KeyHint:    model.MaskToken(apiKey.Key),
	}
	resp.SoraAPIKey.Key = ""
	c.JSON(http.StatusOK, resp)
}

// DeleteAPIKey DELETE /admin/api-keys/:id
func (h *AdminHandler) DeleteAPIKey(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&model.SoraAPIKey{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// generateAPIKey 生成随机 API Key（sk- 前缀 + 32 字节十六进制）
func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "sk-" + hex.EncodeToString(b)
}
