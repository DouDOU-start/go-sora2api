package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImageHandler /v1/images 图片生成端点
type ImageHandler struct {
	scheduler *service.Scheduler
	taskStore *service.TaskStore
}

// NewImageHandler 创建 ImageHandler
func NewImageHandler(scheduler *service.Scheduler, taskStore *service.TaskStore) *ImageHandler {
	return &ImageHandler{scheduler: scheduler, taskStore: taskStore}
}

// CreateImageTask POST /v1/images — 创建图片任务
func (h *ImageHandler) CreateImageTask(c *gin.Context) {
	var req model.ImageSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	// 默认宽高
	if req.Width <= 0 {
		req.Width = 1792
	}
	if req.Height <= 0 {
		req.Height = 1024
	}

	// 从上下文获取 API Key 绑定的分组 ID
	var groupID *int64
	if gid, exists := c.Get("api_key_group_id"); exists {
		id := gid.(int64)
		groupID = &id
	}

	// 选取可用账号
	account, err := h.scheduler.PickAccount(groupID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("无可用账号: %v", err)},
		})
		return
	}

	// 创建 Sora 客户端
	client, err := sora.New(h.scheduler.GetProxyURL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("创建 Sora 客户端失败: %v", err)},
		})
		return
	}

	ctx := c.Request.Context()

	// 生成 Sentinel Token
	sentinel, err := client.GenerateSentinelToken(ctx, account.AccessToken)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	// 处理图片引用（图生图），支持 URL 和 base64 data URI
	var mediaID string
	if req.InputReference != "" {
		var imgData []byte
		var ext string
		if sora.IsDataURI(req.InputReference) {
			var parseErr error
			imgData, ext, parseErr = sora.ParseDataURI(req.InputReference)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": &model.TaskErrorInfo{Message: fmt.Sprintf("解析参考图片 base64 失败: %v", parseErr)},
				})
				return
			}
		} else {
			var dlErr error
			imgData, dlErr = client.DownloadFile(ctx, req.InputReference)
			if dlErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": &model.TaskErrorInfo{Message: fmt.Sprintf("下载参考图片失败: %v", dlErr)},
				})
				return
			}
			ext = sora.ExtFromURL(req.InputReference, ".png")
		}
		mediaID, err = client.UploadImage(ctx, account.AccessToken, imgData, "reference"+ext)
		if err != nil {
			h.handleSubmitError(c, account, err)
			return
		}
	}

	// 提交图片任务
	soraTaskID, err := client.CreateImageTaskWithImage(
		ctx, account.AccessToken, sentinel,
		req.Prompt, req.Width, req.Height, mediaID,
	)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	// 创建内部任务记录
	taskID := "task_" + uuid.New().String()[:8]
	var apiKeyID int64
	if kid, exists := c.Get("api_key_id"); exists {
		apiKeyID = kid.(int64)
	}
	task := &model.SoraTask{
		ID:         taskID,
		SoraTaskID: soraTaskID,
		AccountID:  account.ID,
		APIKeyID:   apiKeyID,
		Type:       "image",
		Model:      "sora-image",
		Prompt:     req.Prompt,
		Status:     model.TaskStatusQueued,
	}

	if err := h.taskStore.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("保存任务记录失败: %v", err)},
		})
		return
	}

	// 启动后台轮询（TaskStore 根据 Type="image" 自动走 pollImageTask）
	h.taskStore.StartPolling(task, account)

	log.Printf("[handler] 图片任务已创建: %s → Sora: %s（账号: %s）",
		taskID, soraTaskID, account.Email)

	c.JSON(http.StatusOK, model.ImageTaskResponse{
		ID:        taskID,
		Object:    "image",
		Status:    model.TaskStatusQueued,
		Progress:  0,
		CreatedAt: time.Now().Unix(),
		Width:     req.Width,
		Height:    req.Height,
	})
}

// GetImageTaskStatus GET /v1/images/:id — 查询图片任务状态
func (h *ImageHandler) GetImageTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskStore.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": &model.TaskErrorInfo{Message: "任务不存在"},
		})
		return
	}

	resp := model.ImageTaskResponse{
		ID:        task.ID,
		Object:    "image",
		Status:    task.Status,
		Progress:  task.Progress,
		CreatedAt: task.CreatedAt.Unix(),
		ImageURL:  task.ImageURL,
	}

	if task.Status == model.TaskStatusFailed && task.ErrorMessage != "" {
		resp.Error = &model.TaskErrorInfo{Message: task.ErrorMessage}
	}

	c.JSON(http.StatusOK, resp)
}

// DownloadImage GET /v1/images/:id/content — 下载图片
func (h *ImageHandler) DownloadImage(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskStore.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": &model.TaskErrorInfo{Message: "任务不存在"},
		})
		return
	}

	if task.Status != model.TaskStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("任务尚未完成（当前状态: %s）", task.Status)},
		})
		return
	}

	if task.ImageURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: "图片 URL 为空"},
		})
		return
	}

	body, contentLength, contentType, err := h.taskStore.DownloadImage(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: err.Error()},
		})
		return
	}
	defer body.Close()

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

// handleSubmitError 处理提交任务时的错误
func (h *ImageHandler) handleSubmitError(c *gin.Context, account *model.SoraAccount, err error) {
	errMsg := err.Error()
	if contains401(errMsg) {
		h.scheduler.MarkAccountError(account.ID, model.AccountStatusTokenExpired, errMsg)
	} else if containsRateLimit(errMsg) {
		h.scheduler.MarkRateLimited(account.ID, 300)
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": &model.TaskErrorInfo{Message: fmt.Sprintf("提交 Sora 任务失败: %v", err)},
	})
}
