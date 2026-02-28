package handler

import (
	"net/http"
	"strconv"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
)

// ListCharacters GET /admin/characters — 角色列表（分页 + 状态筛选）
func (h *AdminHandler) ListCharacters(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&model.SoraCharacter{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var characters []model.SoraCharacter
	query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&characters)

	// 批量获取关联账号邮箱
	accountIDs := make([]int64, 0, len(characters))
	for _, ch := range characters {
		accountIDs = append(accountIDs, ch.AccountID)
	}
	emailMap := make(map[int64]string)
	if len(accountIDs) > 0 {
		var accounts []model.SoraAccount
		h.db.Select("id, email").Where("id IN ?", accountIDs).Find(&accounts)
		for _, a := range accounts {
			emailMap[a.ID] = a.Email
		}
	}

	// 构建响应
	list := make([]model.AdminCharacterResponse, 0, len(characters))
	for _, ch := range characters {
		list = append(list, model.AdminCharacterResponse{
			SoraCharacter: ch,
			AccountEmail:  emailMap[ch.AccountID],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCharacterAdmin GET /admin/characters/:id — 角色详情
func (h *AdminHandler) GetCharacterAdmin(c *gin.Context) {
	charID := c.Param("id")

	var ch model.SoraCharacter
	if err := h.db.Where("id = ?", charID).First(&ch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	var accountEmail string
	var account model.SoraAccount
	if err := h.db.Select("id, email").Where("id = ?", ch.AccountID).First(&account).Error; err == nil {
		accountEmail = account.Email
	}

	c.JSON(http.StatusOK, model.AdminCharacterResponse{
		SoraCharacter: ch,
		AccountEmail:  accountEmail,
	})
}

// DeleteCharacterAdmin DELETE /admin/characters/:id — 删除角色
func (h *AdminHandler) DeleteCharacterAdmin(c *gin.Context) {
	charID := c.Param("id")

	var ch model.SoraCharacter
	if err := h.db.Where("id = ?", charID).First(&ch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	// 如果已定稿，调用 Sora 删除
	if ch.CharacterID != "" {
		var account model.SoraAccount
		if err := h.db.Where("id = ?", ch.AccountID).First(&account).Error; err == nil {
			proxyURL := h.settings.GetProxyURL()
			client, err := sora.New(proxyURL)
			if err == nil {
				_ = client.DeleteCharacter(c.Request.Context(), account.AccessToken, ch.CharacterID)
			}
		}
	}

	h.db.Delete(&ch)
	c.Status(http.StatusNoContent)
}
