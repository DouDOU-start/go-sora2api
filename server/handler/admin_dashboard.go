package handler

import (
	"net/http"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/gin-gonic/gin"
)

// GetDashboard GET /admin/dashboard — 概览统计
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	var stats model.DashboardStats

	h.db.Model(&model.SoraAccount{}).Count(&stats.TotalAccounts)
	h.db.Model(&model.SoraAccount{}).Where("status = ?", model.AccountStatusActive).Count(&stats.ActiveAccounts)
	h.db.Model(&model.SoraAccount{}).Where("status = ?", model.AccountStatusTokenExpired).Count(&stats.ExpiredAccounts)
	h.db.Model(&model.SoraAccount{}).Where("status = ?", model.AccountStatusQuotaExhausted).Count(&stats.ExhaustedAccounts)

	h.db.Model(&model.SoraTask{}).Count(&stats.TotalTasks)
	h.db.Model(&model.SoraTask{}).Where("status IN ?", []string{model.TaskStatusQueued, model.TaskStatusInProgress}).Count(&stats.PendingTasks)
	h.db.Model(&model.SoraTask{}).Where("status = ?", model.TaskStatusCompleted).Count(&stats.CompletedTasks)
	h.db.Model(&model.SoraTask{}).Where("status = ?", model.TaskStatusFailed).Count(&stats.FailedTasks)

	c.JSON(http.StatusOK, stats)
}
