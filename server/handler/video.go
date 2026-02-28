package handler

import (
	"context"
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

// CreateTask POST /v1/videos — 创建视频任务（文生视频/图生视频）
func (h *VideoHandler) CreateTask(c *gin.Context) {
	var req model.VideoSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	params, err := model.ParseModelName(req.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: err.Error()},
		})
		return
	}

	account, client, sentinel, err := h.prepare(c)
	if err != nil {
		return // prepare 已写入响应
	}

	ctx := c.Request.Context()

	prompt := req.Prompt
	styleID := req.Style
	if styleID == "" {
		prompt, styleID = sora.ExtractStyle(req.Prompt)
	}

	mediaID, err := h.resolveInputReference(ctx, c, client, account, req.InputReference)
	if err != nil {
		return
	}

	log.Printf("[handler] 创建视频: model=%s, orientation=%s, nFrames=%d, size=%s, style=%s, mediaID=%s, 账号=%s",
		params.Model, params.Orientation, params.NFrames, params.Size, styleID, mediaID, account.Email)

	soraTaskID, err := client.CreateVideoTaskWithOptions(
		ctx, account.AccessToken, sentinel,
		prompt, params.Orientation, params.NFrames,
		params.Model, params.Size, mediaID, styleID,
	)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	h.finishTask(c, soraTaskID, account, req.Model, req.Prompt, params)
}

// RemixTask POST /v1/videos/remix — Remix 视频
func (h *VideoHandler) RemixTask(c *gin.Context) {
	var req model.RemixSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	params, err := model.ParseModelName(req.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: err.Error()},
		})
		return
	}

	remixTargetID := sora.ExtractRemixID(req.RemixTarget)
	if remixTargetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: "无效的 remix_target，需要 Sora 分享链接或 s_xxx 格式 ID"},
		})
		return
	}

	account, _, sentinel, err := h.prepare(c)
	if err != nil {
		return
	}

	ctx := c.Request.Context()
	prompt := req.Prompt
	styleID := req.Style
	if styleID == "" {
		prompt, styleID = sora.ExtractStyle(req.Prompt)
	}

	client, _ := sora.New(h.scheduler.GetProxyURL())
	soraTaskID, err := client.RemixVideo(
		ctx, account.AccessToken, sentinel,
		remixTargetID, prompt, params.Orientation, params.NFrames, styleID,
	)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	h.finishTask(c, soraTaskID, account, req.Model, req.Prompt, params)
}

// StoryboardTask POST /v1/videos/storyboard — 分镜视频
func (h *VideoHandler) StoryboardTask(c *gin.Context) {
	var req model.StoryboardSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	params, err := model.ParseModelName(req.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: err.Error()},
		})
		return
	}

	account, client, sentinel, err := h.prepare(c)
	if err != nil {
		return
	}

	ctx := c.Request.Context()
	prompt := req.Prompt
	styleID := req.Style
	if styleID == "" {
		prompt, styleID = sora.ExtractStyle(req.Prompt)
	}

	mediaID, err := h.resolveInputReference(ctx, c, client, account, req.InputReference)
	if err != nil {
		return
	}

	soraTaskID, err := client.CreateStoryboardTask(
		ctx, account.AccessToken, sentinel,
		prompt, params.Orientation, params.NFrames, mediaID, styleID,
	)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	h.finishTask(c, soraTaskID, account, req.Model, req.Prompt, params)
}

// prepare 公共准备逻辑：获取分组、选账号、创建客户端、生成 sentinel
func (h *VideoHandler) prepare(c *gin.Context) (account *model.SoraAccount, client *sora.Client, sentinel string, err error) {
	var groupID *int64
	if gid, exists := c.Get("api_key_group_id"); exists {
		id := gid.(int64)
		groupID = &id
	}

	account, err = h.scheduler.PickAccount(groupID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("无可用账号: %v", err)},
		})
		return
	}

	client, err = sora.New(h.scheduler.GetProxyURL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("创建 Sora 客户端失败: %v", err)},
		})
		return
	}

	sentinel, err = client.GenerateSentinelToken(c.Request.Context(), account.AccessToken)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return
	}

	return
}

// resolveInputReference 处理参考图输入（URL 或 base64 data URI），返回 mediaID
func (h *VideoHandler) resolveInputReference(ctx context.Context, c *gin.Context, client *sora.Client, account *model.SoraAccount, inputRef string) (string, error) {
	if inputRef == "" {
		return "", nil
	}

	var imgData []byte
	var ext string
	if sora.IsDataURI(inputRef) {
		var parseErr error
		imgData, ext, parseErr = sora.ParseDataURI(inputRef)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": &model.TaskErrorInfo{Message: fmt.Sprintf("解析参考图片 base64 失败: %v", parseErr)},
			})
			return "", parseErr
		}
	} else {
		var dlErr error
		imgData, dlErr = client.DownloadFile(ctx, inputRef)
		if dlErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": &model.TaskErrorInfo{Message: fmt.Sprintf("下载参考图片失败: %v", dlErr)},
			})
			return "", dlErr
		}
		ext = sora.ExtFromURL(inputRef, ".png")
	}

	mediaID, err := client.UploadImage(ctx, account.AccessToken, imgData, "reference"+ext)
	if err != nil {
		h.handleSubmitError(c, account, err)
		return "", err
	}
	return mediaID, nil
}

// finishTask 公共收尾逻辑：创建任务记录、启动轮询、返回响应
func (h *VideoHandler) finishTask(c *gin.Context, soraTaskID string, account *model.SoraAccount, modelName, prompt string, params *model.ModelParams) {
	taskID := "task_" + uuid.New().String()[:8]
	task := &model.SoraTask{
		ID:         taskID,
		SoraTaskID: soraTaskID,
		AccountID:  account.ID,
		Type:       "video",
		Model:      modelName,
		Prompt:     prompt,
		Status:     model.TaskStatusQueued,
	}

	if err := h.taskStore.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("保存任务记录失败: %v", err)},
		})
		return
	}

	h.taskStore.StartPolling(task, account)

	log.Printf("[handler] 任务已创建: %s → Sora: %s（账号: %s, 模型: %s）",
		taskID, soraTaskID, account.Email, modelName)

	c.JSON(http.StatusOK, model.VideoTaskResponse{
		ID:        taskID,
		Object:    "video",
		Model:     modelName,
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
	if contains401(errMsg) {
		h.scheduler.MarkAccountError(account.ID, model.AccountStatusTokenExpired, errMsg)
	} else if containsRateLimit(errMsg) {
		h.scheduler.MarkRateLimited(account.ID, 300)
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": &model.TaskErrorInfo{Message: fmt.Sprintf("提交 Sora 任务失败: %v", err)},
	})
}

// ---- 公共错误判断函数 ----

func contains401(msg string) bool {
	return strings.Contains(msg, "401") || strings.Contains(msg, "Unauthorized")
}

func containsRateLimit(msg string) bool {
	return strings.Contains(msg, "429") || strings.Contains(msg, "rate limit")
}
