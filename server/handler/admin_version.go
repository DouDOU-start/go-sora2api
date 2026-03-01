package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

// GetVersion GET /admin/version — 获取版本信息及最新版本
func (h *AdminHandler) GetVersion(c *gin.Context) {
	type githubRelease struct {
		TagName string `json:"tag_name"`
	}

	current := h.version
	latest := ""
	hasUpdate := false

	// 从 GitHub API 获取最新版本（10 秒超时）
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get("https://api.github.com/repos/DouDOU-start/go-sora2api/releases/latest")
	if err == nil {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("[admin_version] close response body failed: %v", err)
			}
		}()
		var rel githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil {
			latest = rel.TagName
			if latest != "" && latest != current && current != "dev" {
				hasUpdate = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current":    current,
		"latest":     latest,
		"has_update": hasUpdate,
	})
}

// TriggerUpgrade POST /admin/upgrade — 触发升级（仅一键安装方式支持）
func (h *AdminHandler) TriggerUpgrade(c *gin.Context) {
	if _, err := exec.LookPath("sora2api"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sora2api 命令不可用，仅一键安装方式支持此功能"})
		return
	}

	// 升级脚本需要 root 权限（systemctl stop/start），通过 sudoers 规则免密执行
	cmd := exec.Command("sudo", "sora2api", "upgrade")
	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("启动升级失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "升级命令已启动，服务将在几分钟内重启"})
}
