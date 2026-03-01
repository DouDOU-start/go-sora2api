package handler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListTasks GET /admin/tasks（viewer 角色只能看到自己 API Key 创建的任务）
func (h *AdminHandler) ListTasks(c *gin.Context) {
	status := c.Query("status")
	taskType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// viewer 角色按 API Key ID 过滤
	var apiKeyID int64
	if role, _ := c.Get("role"); role == RoleViewer {
		if kid, exists := c.Get("jwt_api_key_id"); exists {
			apiKeyID = kid.(int64)
		}
	}

	tasks, total, err := h.taskStore.ListTasks(status, taskType, page, pageSize, apiKeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTask GET /admin/tasks/:id（viewer 只能查看自己 API Key 创建的任务）
func (h *AdminHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskStore.Get(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// viewer 只能查看自己的任务
	if role, _ := c.Get("role"); role == RoleViewer {
		if kid, exists := c.Get("jwt_api_key_id"); exists {
			if task.APIKeyID != kid.(int64) {
				c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, task)
}

// DownloadTaskContent GET /admin/tasks/:id/content — 下载任务产物（视频或图片）
func (h *AdminHandler) DownloadTaskContent(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskStore.Get(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// viewer 只能下载自己的任务
	if role, _ := c.Get("role"); role == RoleViewer {
		if kid, exists := c.Get("jwt_api_key_id"); exists {
			if task.APIKeyID != kid.(int64) {
				c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
				return
			}
		}
	}

	if task.Status != model.TaskStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("任务尚未完成（当前状态: %s）", task.Status)})
		return
	}

	var body io.ReadCloser
	var contentLength int64
	var contentType string

	switch task.Type {
	case "image":
		body, contentLength, contentType, err = h.taskStore.DownloadImage(c.Request.Context(), task)
	default:
		body, contentLength, contentType, err = h.taskStore.DownloadVideo(c.Request.Context(), task)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Printf("[admin_task] close body failed: %v", err)
		}
	}()

	c.Header("Content-Type", contentType)
	if contentLength > 0 {
		c.Header("Content-Length", fmt.Sprintf("%d", contentLength))
	}
	c.Status(http.StatusOK)

	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 32*1024)
		_, err := io.CopyBuffer(w, body, buf)
		return err != nil
	})
}
