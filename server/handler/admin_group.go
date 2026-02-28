package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/gin-gonic/gin"
)

// ListGroups GET /admin/groups
func (h *AdminHandler) ListGroups(c *gin.Context) {
	var groups []model.SoraAccountGroup
	h.db.Order("id ASC").Find(&groups)

	type GroupWithCount struct {
		model.SoraAccountGroup
		AccountCount int64 `json:"account_count"`
	}

	var resp []GroupWithCount
	for _, g := range groups {
		var count int64
		h.db.Model(&model.SoraAccount{}).Where("group_id = ?", g.ID).Count(&count)
		resp = append(resp, GroupWithCount{
			SoraAccountGroup: g,
			AccountCount:     count,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// CreateGroup POST /admin/groups
func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var req model.AdminGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := model.SoraAccountGroup{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     true,
	}
	if req.Enabled != nil {
		group.Enabled = *req.Enabled
	}

	if err := h.db.Create(&group).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("创建账号组失败: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// UpdateGroup PUT /admin/groups/:id
func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var group model.SoraAccountGroup
	if err := h.db.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号组不存在"})
		return
	}

	var req model.AdminGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.Name = req.Name
	group.Description = req.Description
	if req.Enabled != nil {
		group.Enabled = *req.Enabled
	}

	if err := h.db.Save(&group).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("更新账号组失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup DELETE /admin/groups/:id
func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// 解绑该分组下的所有账号（设为未分组）
	h.db.Model(&model.SoraAccount{}).Where("group_id = ?", id).Update("group_id", nil)

	if err := h.db.Delete(&model.SoraAccountGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
