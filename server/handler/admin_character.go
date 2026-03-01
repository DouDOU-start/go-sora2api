package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
)

// ListCharacters GET /admin/characters — 角色列表（分页 + 状态筛选 + is_public 筛选）
func (h *AdminHandler) ListCharacters(c *gin.Context) {
	status := c.Query("status")
	isPublicStr := c.Query("is_public")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 排除 profile_image 大字段，避免列表查询传输大量数据
	query := h.db.Model(&model.SoraCharacter{}).Omit("profile_image")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	switch isPublicStr {
	case "true":
		query = query.Where("is_public = ?", true)
	case "false":
		query = query.Where("is_public = ?", false)
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

	// 构建响应（标记哪些角色有本地图片）
	list := make([]model.AdminCharacterResponse, 0, len(characters))
	for _, ch := range characters {
		resp := model.AdminCharacterResponse{
			SoraCharacter: ch,
			AccountEmail:  emailMap[ch.AccountID],
		}
		list = append(list, resp)
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
	// 排除 profile_image 大字段
	if err := h.db.Omit("profile_image").Where("id = ?", charID).First(&ch).Error; err != nil {
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

// GetCharacterImage GET /admin/characters/:id/image — 获取角色头像图片
func (h *AdminHandler) GetCharacterImage(c *gin.Context) {
	charID := c.Param("id")

	var ch model.SoraCharacter
	if err := h.db.Select("id, profile_image").Where("id = ?", charID).First(&ch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	if len(ch.ProfileImage) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色图片不存在"})
		return
	}

	// 检测图片格式并设置 Content-Type
	contentType := http.DetectContentType(ch.ProfileImage)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // 缓存 1 天
	c.Data(http.StatusOK, contentType, ch.ProfileImage)
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

// ToggleCharacterVisibility POST /admin/characters/:id/visibility — 切换角色公开/私密
func (h *AdminHandler) ToggleCharacterVisibility(c *gin.Context) {
	charID := c.Param("id")

	var ch model.SoraCharacter
	if err := h.db.Omit("profile_image").Where("id = ?", charID).First(&ch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	if ch.Status != model.CharacterStatusReady {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色尚未就绪，无法切换可见性"})
		return
	}

	// 获取关联账号
	var account model.SoraAccount
	if err := h.db.Where("id = ?", ch.AccountID).First(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "找不到关联账号"})
		return
	}

	proxyURL := h.settings.GetProxyURL()
	client, err := sora.New(proxyURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建客户端失败"})
		return
	}

	// 切换：当前公开 → 设为私密；当前私密 → 设为公开
	newPublic := !ch.IsPublic
	var visibility string
	if newPublic {
		visibility = "public"
	} else {
		visibility = "private"
	}

	if err := client.SetCharacterVisibility(c.Request.Context(), account.AccessToken, ch.CameoID, visibility); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("设置可见性失败: %v", err)})
		return
	}

	// 更新数据库
	h.db.Model(&model.SoraCharacter{}).Where("id = ?", charID).Update("is_public", newPublic)

	label := "公开"
	if !newPublic {
		label = "私密"
	}
	c.JSON(http.StatusOK, gin.H{"message": "角色已设为" + label, "is_public": newPublic})
}
