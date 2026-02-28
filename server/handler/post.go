package handler

import (
	"fmt"
	"net/http"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PostHandler /v1/posts 帖子管理 + /v1/watermark-free 无水印下载
type PostHandler struct {
	scheduler *service.Scheduler
	taskStore *service.TaskStore
	db        *gorm.DB
}

// NewPostHandler 创建 PostHandler
func NewPostHandler(scheduler *service.Scheduler, taskStore *service.TaskStore, db *gorm.DB) *PostHandler {
	return &PostHandler{scheduler: scheduler, taskStore: taskStore, db: db}
}

// PublishPost POST /v1/posts — 发布视频帖子
func (h *PostHandler) PublishPost(c *gin.Context) {
	var req model.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	// 查找任务
	task, err := h.taskStore.Get(req.TaskID)
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

	if task.Type != "video" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: "仅支持发布视频任务"},
		})
		return
	}

	// 获取关联账号
	var account model.SoraAccount
	if err := h.db.Where("id = ?", task.AccountID).First(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: "找不到关联账号"},
		})
		return
	}

	client, err := sora.New(h.scheduler.GetProxyURL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("创建 Sora 客户端失败: %v", err)},
		})
		return
	}

	ctx := c.Request.Context()

	// 获取 generationID
	generationID, err := client.GetGenerationID(ctx, account.AccessToken, task.SoraTaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("获取 generation ID 失败: %v", err)},
		})
		return
	}

	// 生成 Sentinel Token
	sentinel, err := client.GenerateSentinelToken(ctx, account.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("生成 Sentinel Token 失败: %v", err)},
		})
		return
	}

	// 发布
	postID, err := client.PublishVideo(ctx, account.AccessToken, sentinel, generationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("发布视频失败: %v", err)},
		})
		return
	}

	c.JSON(http.StatusOK, model.PostResponse{PostID: postID})
}

// DeletePost DELETE /v1/posts/:id — 删除帖子
func (h *PostHandler) DeletePost(c *gin.Context) {
	postID := c.Param("id")

	// 从上下文获取分组 ID
	var groupID *int64
	if gid, exists := c.Get("api_key_group_id"); exists {
		id := gid.(int64)
		groupID = &id
	}

	account, err := h.scheduler.PickAccount(groupID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("无可用账号: %v", err)},
		})
		return
	}

	client, err := sora.New(h.scheduler.GetProxyURL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("创建 Sora 客户端失败: %v", err)},
		})
		return
	}

	if err := client.DeletePost(c.Request.Context(), account.AccessToken, postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("删除帖子失败: %v", err)},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetWatermarkFreeURL POST /v1/watermark-free — 获取无水印下载链接
func (h *PostHandler) GetWatermarkFreeURL(c *gin.Context) {
	var req model.WatermarkFreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	// 从上下文获取分组 ID
	var groupID *int64
	if gid, exists := c.Get("api_key_group_id"); exists {
		id := gid.(int64)
		groupID = &id
	}

	account, err := h.scheduler.PickAccount(groupID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("无可用账号: %v", err)},
		})
		return
	}

	client, err := sora.New(h.scheduler.GetProxyURL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("创建 Sora 客户端失败: %v", err)},
		})
		return
	}

	url, err := client.GetWatermarkFreeURL(c.Request.Context(), account.AccessToken, req.VideoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("获取无水印链接失败: %v", err)},
		})
		return
	}

	c.JSON(http.StatusOK, model.WatermarkFreeResponse{URL: url})
}
