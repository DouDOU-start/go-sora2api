package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VideoHandler /v1/videos 核心端点
type VideoHandler struct {
	scheduler *service.Scheduler
	taskStore *service.TaskStore
}

// NewVideoHandler 创建 VideoHandler
func NewVideoHandler(scheduler *service.Scheduler, taskStore *service.TaskStore) *VideoHandler {
	return &VideoHandler{scheduler: scheduler, taskStore: taskStore}
}

// CreateTask POST /v1/videos — 创建任务
func (h *VideoHandler) CreateTask(c *gin.Context) {
	var req model.VideoSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	// 解析模型名称
	params, err := model.ParseModelName(req.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: err.Error()},
		})
		return
	}

	// 从上下文获取 API Key 绑定的分组 ID（由中间件设置）
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

	// 生成 Sentinel Token（PoW）
	sentinel, err := client.GenerateSentinelToken(ctx, account.AccessToken)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	// 处理图片引用（图生视频）
	var mediaID string
	if req.InputReference != "" {
		// 下载图片
		imgData, dlErr := client.DownloadFile(ctx, req.InputReference)
		if dlErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": &model.TaskErrorInfo{Message: fmt.Sprintf("下载参考图片失败: %v", dlErr)},
			})
			return
		}
		ext := sora.ExtFromURL(req.InputReference, ".png")
		mediaID, err = client.UploadImage(ctx, account.AccessToken, imgData, "reference"+ext)
		if err != nil {
			h.handleSubmitError(c, account, err)
			return
		}
	}

	// 提交任务到 Sora
	soraTaskID, err := client.CreateVideoTaskWithOptions(
		ctx, account.AccessToken, sentinel,
		req.Prompt, params.Orientation, params.NFrames,
		params.Model, params.Size, mediaID, "",
	)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	// 创建内部任务记录
	taskID := "task_" + uuid.New().String()[:8]
	task := &model.SoraTask{
		ID:         taskID,
		SoraTaskID: soraTaskID,
		AccountID:  account.ID,
		Type:       "video",
		Model:      req.Model,
		Prompt:     req.Prompt,
		Status:     model.TaskStatusQueued,
	}

	if err := h.taskStore.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("保存任务记录失败: %v", err)},
		})
		return
	}

	// 启动后台轮询
	h.taskStore.StartPolling(task, account)

	log.Printf("[handler] 任务已创建: %s → Sora: %s（账号: %d, 模型: %s）",
		taskID, soraTaskID, account.ID, req.Model)

	// 返回响应
	c.JSON(http.StatusOK, model.VideoTaskResponse{
		ID:        taskID,
		Object:    "video",
		Model:     req.Model,
		Status:    model.TaskStatusQueued,
		Progress:  0,
		CreatedAt: time.Now().Unix(),
		Size:      model.SizeToResolution(params.Size, params.Orientation),
	})
}

// GetTaskStatus GET /v1/videos/:id — 查询任务状态
func (h *VideoHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskStore.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": &model.TaskErrorInfo{Message: "任务不存在"},
		})
		return
	}

	resp := model.VideoTaskResponse{
		ID:        task.ID,
		Object:    task.Type,
		Model:     task.Model,
		Status:    task.Status,
		Progress:  task.Progress,
		CreatedAt: task.CreatedAt.Unix(),
	}

	// 解析模型获取分辨率
	if params, err := model.ParseModelName(task.Model); err == nil {
		resp.Size = model.SizeToResolution(params.Size, params.Orientation)
	}

	if task.Status == model.TaskStatusFailed && task.ErrorMessage != "" {
		resp.Error = &model.TaskErrorInfo{Message: task.ErrorMessage}
	}

	c.JSON(http.StatusOK, resp)
}

// DownloadVideo GET /v1/videos/:id/content — 下载视频
func (h *VideoHandler) DownloadVideo(c *gin.Context) {
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

	body, contentLength, contentType, err := h.taskStore.DownloadVideo(c.Request.Context(), task)
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

	// 流式转发
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 32*1024)
		_, err := io.CopyBuffer(w, body, buf)
		return err != nil
	})
}

// handleSubmitError 处理提交任务时的错误
func (h *VideoHandler) handleSubmitError(c *gin.Context, account *model.SoraAccount, err error) {
	errMsg := err.Error()

	// 根据错误类型更新账号状态
	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized") {
		h.scheduler.MarkAccountError(account.ID, model.AccountStatusTokenExpired, errMsg)
	} else if strings.Contains(errMsg, "429") || strings.Contains(errMsg, "rate limit") {
		h.scheduler.MarkRateLimited(account.ID, 300) // 默认 5 分钟
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": &model.TaskErrorInfo{Message: fmt.Sprintf("提交 Sora 任务失败: %v", err)},
	})
}
