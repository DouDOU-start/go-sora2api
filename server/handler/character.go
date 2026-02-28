package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/DouDOU-start/go-sora2api/sora"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CharacterHandler /v1/characters 角色管理端点
type CharacterHandler struct {
	scheduler *service.Scheduler
	db        *gorm.DB
}

// NewCharacterHandler 创建 CharacterHandler
func NewCharacterHandler(scheduler *service.Scheduler, db *gorm.DB) *CharacterHandler {
	return &CharacterHandler{scheduler: scheduler, db: db}
}

// CreateCharacter POST /v1/characters — 创建角色
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	var req model.CharacterCreateRequest
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

	// 选取可用账号
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

	ctx := c.Request.Context()

	// 获取视频数据，支持 URL 和 base64 data URI
	var videoData []byte
	if sora.IsDataURI(req.VideoURL) {
		var parseErr error
		videoData, _, parseErr = sora.ParseDataURI(req.VideoURL)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": &model.TaskErrorInfo{Message: fmt.Sprintf("解析视频 base64 失败: %v", parseErr)},
			})
			return
		}
	} else {
		var dlErr error
		videoData, dlErr = client.DownloadFile(ctx, req.VideoURL)
		if dlErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": &model.TaskErrorInfo{Message: fmt.Sprintf("下载角色视频失败: %v", dlErr)},
			})
			return
		}
	}

	// 上传视频获取 cameoID
	cameoID, err := client.UploadCharacterVideo(ctx, account.AccessToken, videoData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("上传角色视频失败: %v", err)},
		})
		return
	}

	// 创建内部角色记录
	charID := "char_" + uuid.New().String()[:8]
	character := &model.SoraCharacter{
		ID:        charID,
		AccountID: account.ID,
		CameoID:   cameoID,
		Status:    model.CharacterStatusProcessing,
	}

	if err := h.db.Create(character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("保存角色记录失败: %v", err)},
		})
		return
	}

	// 启动后台异步处理（轮询 → 下载头像 → 上传 → 定稿）
	go h.processCharacter(character, account, req.Username, req.DisplayName)

	log.Printf("[handler] 角色已创建: %s → Cameo: %s（账号: %s）", charID, cameoID, account.Email)

	c.JSON(http.StatusOK, model.CharacterResponse{
		ID:        charID,
		Status:    model.CharacterStatusProcessing,
		CreatedAt: time.Now().Unix(),
	})
}

// processCharacter 后台处理角色：轮询 → 下载头像 → 上传 → 定稿
func (h *CharacterHandler) processCharacter(char *model.SoraCharacter, account *model.SoraAccount, username, displayName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := sora.New(h.scheduler.GetProxyURL())
	if err != nil {
		h.failCharacter(char.ID, fmt.Sprintf("创建 Sora 客户端失败: %v", err))
		return
	}

	// 1. 轮询 cameo 状态直到完成
	cameoStatus, err := client.PollCameoStatus(ctx, account.AccessToken, char.CameoID, 5*time.Second, 8*time.Minute, nil)
	if err != nil {
		h.failCharacter(char.ID, fmt.Sprintf("角色处理失败: %v", err))
		return
	}

	// 2. 使用推荐的名称（如果请求中未指定）
	if username == "" {
		username = cameoStatus.UsernameHint
	}
	if displayName == "" {
		displayName = cameoStatus.DisplayNameHint
	}
	// 确保有默认值
	if username == "" {
		username = "character_" + char.ID[5:]
	}
	if displayName == "" {
		displayName = "Character"
	}

	// 3. 下载推荐头像并上传获取 assetPointer
	imgData, err := client.DownloadCharacterImage(ctx, cameoStatus.ProfileAssetURL)
	if err != nil {
		h.failCharacter(char.ID, fmt.Sprintf("下载角色头像失败: %v", err))
		return
	}

	assetPointer, err := client.UploadCharacterImage(ctx, account.AccessToken, imgData)
	if err != nil {
		h.failCharacter(char.ID, fmt.Sprintf("上传角色头像失败: %v", err))
		return
	}

	// 4. 定稿角色
	characterID, err := client.FinalizeCharacter(ctx, account.AccessToken, char.CameoID, username, displayName, assetPointer)
	if err != nil {
		h.failCharacter(char.ID, fmt.Sprintf("定稿角色失败: %v", err))
		return
	}

	// 5. 自动设置角色公开
	isPublic := false
	if err := client.SetCharacterPublic(ctx, account.AccessToken, char.CameoID); err != nil {
		log.Printf("[handler] 角色公开设置失败（非致命）: %s: %v", char.ID, err)
	} else {
		isPublic = true
	}

	// 6. 更新数据库（同时保存图片二进制数据，避免依赖外部临时 URL）
	now := time.Now()
	h.db.Model(&model.SoraCharacter{}).Where("id = ?", char.ID).Updates(map[string]interface{}{
		"status":        model.CharacterStatusReady,
		"character_id":  characterID,
		"display_name":  displayName,
		"username":      username,
		"profile_url":   cameoStatus.ProfileAssetURL,
		"profile_image": imgData,
		"is_public":     isPublic,
		"completed_at":  &now,
	})

	log.Printf("[handler] 角色处理完成: %s → Character: %s", char.ID, characterID)
}

// failCharacter 标记角色处理失败
func (h *CharacterHandler) failCharacter(charID, errMsg string) {
	now := time.Now()
	h.db.Model(&model.SoraCharacter{}).Where("id = ?", charID).Updates(map[string]interface{}{
		"status":        model.CharacterStatusFailed,
		"error_message": errMsg,
		"completed_at":  &now,
	})
	log.Printf("[handler] 角色处理失败: %s: %s", charID, errMsg)
}

// GetCharacter GET /v1/characters/:id — 查询角色状态
func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	charID := c.Param("id")

	var char model.SoraCharacter
	if err := h.db.Where("id = ?", charID).First(&char).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": &model.TaskErrorInfo{Message: "角色不存在"},
		})
		return
	}

	resp := model.CharacterResponse{
		ID:          char.ID,
		Status:      char.Status,
		DisplayName: char.DisplayName,
		Username:    char.Username,
		ProfileURL:  char.ProfileURL,
		CharacterID: char.CharacterID,
		CreatedAt:   char.CreatedAt.Unix(),
	}

	if char.Status == model.CharacterStatusFailed && char.ErrorMessage != "" {
		resp.Error = &model.TaskErrorInfo{Message: char.ErrorMessage}
	}

	c.JSON(http.StatusOK, resp)
}

// SetPublic POST /v1/characters/:id/public — 设置角色公开
func (h *CharacterHandler) SetPublic(c *gin.Context) {
	charID := c.Param("id")

	var char model.SoraCharacter
	if err := h.db.Where("id = ?", charID).First(&char).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": &model.TaskErrorInfo{Message: "角色不存在"},
		})
		return
	}

	if char.Status != model.CharacterStatusReady {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("角色尚未就绪（当前状态: %s）", char.Status)},
		})
		return
	}

	// 获取关联账号
	var account model.SoraAccount
	if err := h.db.Where("id = ?", char.AccountID).First(&account).Error; err != nil {
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

	if err := client.SetCharacterPublic(c.Request.Context(), account.AccessToken, char.CameoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &model.TaskErrorInfo{Message: fmt.Sprintf("设置角色公开失败: %v", err)},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色已设为公开"})
}

// DeleteCharacter DELETE /v1/characters/:id — 删除角色
func (h *CharacterHandler) DeleteCharacter(c *gin.Context) {
	charID := c.Param("id")

	var char model.SoraCharacter
	if err := h.db.Where("id = ?", charID).First(&char).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": &model.TaskErrorInfo{Message: "角色不存在"},
		})
		return
	}

	// 如果已定稿，调用 Sora 删除
	if char.CharacterID != "" {
		var account model.SoraAccount
		if err := h.db.Where("id = ?", char.AccountID).First(&account).Error; err == nil {
			client, err := sora.New(h.scheduler.GetProxyURL())
			if err == nil {
				_ = client.DeleteCharacter(c.Request.Context(), account.AccessToken, char.CharacterID)
			}
		}
	}

	h.db.Delete(&char)
	c.Status(http.StatusNoContent)
}
