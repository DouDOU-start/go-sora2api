package handler

import (
	"fmt"
	"net/http"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
)

// PromptHandler /v1/enhance-prompt 提示词增强端点
type PromptHandler struct {
	scheduler *service.Scheduler
}

// NewPromptHandler 创建 PromptHandler
func NewPromptHandler(scheduler *service.Scheduler) *PromptHandler {
	return &PromptHandler{scheduler: scheduler}
}

// EnhancePrompt POST /v1/enhance-prompt — 提示词增强
func (h *PromptHandler) EnhancePrompt(c *gin.Context) {
	var req model.EnhancePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("请求参数错误: %v", err)},
		})
		return
	}

	// 默认值
	if req.ExpansionLevel == "" {
		req.ExpansionLevel = "medium"
	}
	if req.Duration <= 0 {
		req.Duration = 10
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

	enhanced, err := client.EnhancePrompt(c.Request.Context(), account.AccessToken, req.Prompt, req.ExpansionLevel, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("提示词增强失败: %v", err)},
		})
		return
	}

	c.JSON(http.StatusOK, model.EnhancePromptResponse{
		OriginalPrompt: req.Prompt,
		EnhancedPrompt: enhanced,
	})
}
