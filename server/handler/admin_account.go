package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/gin-gonic/gin"
)

// buildAccountResponse 构建账号响应（填充分组名称、Token 掩码）
func (h *AdminHandler) buildAccountResponse(acc model.SoraAccount) model.AdminAccountResponse {
	r := model.AdminAccountResponse{
		SoraAccount: acc,
		ATHint:      model.MaskToken(acc.AccessToken),
		RTHint:      model.MaskToken(acc.RefreshToken),
	}
	if acc.GroupID != nil {
		var group model.SoraAccountGroup
		if err := h.db.First(&group, *acc.GroupID).Error; err == nil {
			r.GroupName = group.Name
		}
	}
	return r
}

// ListAllAccounts GET /admin/accounts
func (h *AdminHandler) ListAllAccounts(c *gin.Context) {
	var accounts []model.SoraAccount
	h.db.Order("id ASC").Find(&accounts)

	var resp []model.AdminAccountResponse
	for _, acc := range accounts {
		resp = append(resp, h.buildAccountResponse(acc))
	}

	c.JSON(http.StatusOK, resp)
}

// CreateAccountDirect POST /admin/accounts
func (h *AdminHandler) CreateAccountDirect(c *gin.Context) {
	var req model.AdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AccessToken == "" && req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要提供 Access Token 或 Refresh Token"})
		return
	}

	// 验证分组存在（如果指定了分组）
	if req.GroupID != nil {
		var group model.SoraAccountGroup
		if err := h.db.First(&group, *req.GroupID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "指定的账号组不存在"})
			return
		}
	}

	account := model.SoraAccount{
		GroupID:      req.GroupID,
		Name:         req.Name,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		Enabled:      true,
		Status:       model.AccountStatusActive,
	}
	if req.Enabled != nil {
		account.Enabled = *req.Enabled
	}

	// 如果只提供了 RT，先刷新获取 AT
	if account.AccessToken == "" && account.RefreshToken != "" {
		if err := h.manager.RefreshSingleToken(c.Request.Context(), &account); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("通过 Refresh Token 获取 Access Token 失败: %v", err)})
			return
		}
	}

	// 从 AT 的 JWT payload 中提取邮箱
	if account.AccessToken != "" {
		if email := model.ExtractEmailFromJWT(account.AccessToken); email != "" {
			account.Email = email
		}
	}

	if err := h.db.Create(&account).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("创建账号失败: %v", err)})
		return
	}

	// 同步配额和订阅信息，确保返回给前端的数据完整
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = h.manager.SyncSingleAccountCredit(ctx, &account)
	_ = h.manager.SyncSingleAccountSubscription(ctx, &account)

	// 重新从数据库读取最新状态
	h.db.First(&account, account.ID)
	c.JSON(http.StatusCreated, h.buildAccountResponse(account))
}

// UpdateAccountDirect PUT /admin/accounts/:id
func (h *AdminHandler) UpdateAccountDirect(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var account model.SoraAccount
	if err := h.db.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}

	var req model.AdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证分组存在（如果指定了分组）
	if req.GroupID != nil {
		var group model.SoraAccountGroup
		if err := h.db.First(&group, *req.GroupID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "指定的账号组不存在"})
			return
		}
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	account.GroupID = req.GroupID
	if req.AccessToken != "" {
		account.AccessToken = req.AccessToken
		// 更新 AT 时重新提取邮箱
		if email := model.ExtractEmailFromJWT(req.AccessToken); email != "" {
			account.Email = email
		}
	}
	if req.RefreshToken != "" {
		account.RefreshToken = req.RefreshToken
	}
	if req.Enabled != nil {
		account.Enabled = *req.Enabled
	}

	if err := h.db.Save(&account).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("更新账号失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, h.buildAccountResponse(account))
}

// DeleteAccountDirect DELETE /admin/accounts/:id
func (h *AdminHandler) DeleteAccountDirect(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.SoraAccount{}, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RefreshAccountTokenDirect POST /admin/accounts/:id/refresh
func (h *AdminHandler) RefreshAccountTokenDirect(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var account model.SoraAccount
	if err := h.db.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}

	if account.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该账号没有 Refresh Token"})
		return
	}

	if err := h.manager.RefreshSingleToken(c.Request.Context(), &account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("刷新 Token 失败: %v", err)})
		return
	}

	h.db.First(&account, accountID)
	c.JSON(http.StatusOK, h.buildAccountResponse(account))
}

// RevealAccountTokens GET /admin/accounts/:id/tokens — 获取完整 AT 和 RT
func (h *AdminHandler) RevealAccountTokens(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var account model.SoraAccount
	if err := h.db.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  account.AccessToken,
		"refresh_token": account.RefreshToken,
	})
}

// GetAccountStatusDirect GET /admin/accounts/:id/status
func (h *AdminHandler) GetAccountStatusDirect(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var account model.SoraAccount
	if err := h.db.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}

	_ = h.manager.SyncSingleAccountCredit(c.Request.Context(), &account)
	_ = h.manager.SyncSingleAccountSubscription(c.Request.Context(), &account)

	h.db.First(&account, accountID)
	c.JSON(http.StatusOK, h.buildAccountResponse(account))
}

// BatchImportAccounts POST /admin/accounts/batch
// 批量导入账号：自动识别 RT（rt_ 前缀）或 AT，以邮箱为唯一标识 upsert
func (h *AdminHandler) BatchImportAccounts(c *gin.Context) {
	var req model.AdminBatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Tokens) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tokens 不能为空"})
		return
	}

	// 验证分组
	if req.GroupID != nil {
		var group model.SoraAccountGroup
		if err := h.db.First(&group, *req.GroupID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "指定的账号组不存在"})
			return
		}
	}

	result := model.AdminBatchImportResult{}
	// 单个 token 刷新最多等 30s，整批最多 5 分钟
	batchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, rawToken := range req.Tokens {
		token := strings.TrimSpace(rawToken)
		if token == "" {
			continue
		}

		result.Total++
		item := model.AdminBatchImportItemResult{
			Token: model.MaskToken(token),
		}

		isRT := strings.HasPrefix(token, "rt_")

		// 构造待操作的账号
		acc := model.SoraAccount{
			GroupID: req.GroupID,
			Enabled: true,
			Status:  model.AccountStatusActive,
		}

		if isRT {
			acc.RefreshToken = token
			// 通过 RT 换取 AT
			refreshCtx, refreshCancel := context.WithTimeout(batchCtx, 30*time.Second)
			err := h.manager.RefreshSingleToken(refreshCtx, &acc)
			refreshCancel()
			if err != nil {
				item.Action = "failed"
				item.Error = fmt.Sprintf("刷新 RT 获取 AT 失败: %v", err)
				result.Failed++
				result.Details = append(result.Details, item)
				continue
			}
		} else {
			acc.AccessToken = token
		}

		// 从 AT 提取邮箱
		if acc.AccessToken != "" {
			acc.Email = model.ExtractEmailFromJWT(acc.AccessToken)
		}
		item.Email = acc.Email

		// 以邮箱为唯一标识做 upsert
		if acc.Email != "" {
			var existing model.SoraAccount
			if err := h.db.Where("email = ?", acc.Email).First(&existing).Error; err == nil {
				// 更新已有账号
				if isRT {
					existing.RefreshToken = acc.RefreshToken
					existing.AccessToken = acc.AccessToken
				} else {
					existing.AccessToken = acc.AccessToken
					// 保留原有 RT，不覆盖
				}
				if req.GroupID != nil {
					existing.GroupID = req.GroupID
				}
				if err := h.db.Save(&existing).Error; err != nil {
					item.Action = "failed"
					item.Error = fmt.Sprintf("更新账号失败: %v", err)
					result.Failed++
				} else {
					item.Action = "updated"
					result.Updated++
					go func(a model.SoraAccount) {
						ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
						defer cancel()
						_ = h.manager.SyncSingleAccountCredit(ctx, &a)
						_ = h.manager.SyncSingleAccountSubscription(ctx, &a)
					}(existing)
				}
				result.Details = append(result.Details, item)
				continue
			}
		}

		// 新建账号
		if err := h.db.Create(&acc).Error; err != nil {
			item.Action = "failed"
			item.Error = fmt.Sprintf("创建账号失败: %v", err)
			result.Failed++
		} else {
			item.Action = "created"
			result.Created++
			go func(a model.SoraAccount) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				_ = h.manager.SyncSingleAccountCredit(ctx, &a)
				_ = h.manager.SyncSingleAccountSubscription(ctx, &a)
			}(acc)
		}
		result.Details = append(result.Details, item)
	}

	c.JSON(http.StatusOK, result)
}
