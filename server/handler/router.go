package handler

import (
	"net/http"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/server/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouterConfig 路由配置
type RouterConfig struct {
	DB        *gorm.DB
	Scheduler *service.Scheduler
	TaskStore *service.TaskStore
	Manager   *service.AccountManager
	Settings  *service.SettingsStore
	JWTSecret string
	AdminUser string
	AdminPass string
	Version   string
}

// SetupRouter 注册所有路由
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 登录端点（无需认证）
	r.POST("/admin/login", loginHandler(cfg.JWTSecret, cfg.AdminUser, cfg.AdminPass))
	r.POST("/admin/login/apikey", apiKeyLoginHandler(cfg.JWTSecret, cfg.DB))

	// API 端点（API Key 认证，从数据库查询）
	videoHandler := NewVideoHandler(cfg.Scheduler, cfg.TaskStore)
	imageHandler := NewImageHandler(cfg.Scheduler, cfg.TaskStore)
	characterHandler := NewCharacterHandler(cfg.Scheduler, cfg.DB)
	promptHandler := NewPromptHandler(cfg.Scheduler)
	postHandler := NewPostHandler(cfg.Scheduler, cfg.TaskStore, cfg.DB)

	api := r.Group("/v1", APIKeyAuthMiddleware(cfg.DB))
	{
		// 视频任务
		api.POST("/videos", videoHandler.CreateTask)
		api.POST("/videos/remix", videoHandler.RemixTask)
		api.POST("/videos/storyboard", videoHandler.StoryboardTask)
		api.GET("/videos/:id", videoHandler.GetTaskStatus)
		api.GET("/videos/:id/content", videoHandler.DownloadVideo)

		// 图片任务
		api.POST("/images", imageHandler.CreateImageTask)
		api.GET("/images/:id", imageHandler.GetImageTaskStatus)
		api.GET("/images/:id/content", imageHandler.DownloadImage)

		// 角色管理
		api.POST("/characters", characterHandler.CreateCharacter)
		api.GET("/characters/:id", characterHandler.GetCharacter)
		api.POST("/characters/:id/public", characterHandler.SetPublic)
		api.DELETE("/characters/:id", characterHandler.DeleteCharacter)

		// 提示词增强
		api.POST("/enhance-prompt", promptHandler.EnhancePrompt)

		// 帖子管理
		api.POST("/posts", postHandler.PublishPost)
		api.DELETE("/posts/:id", postHandler.DeletePost)

		// 无水印下载
		api.POST("/watermark-free", postHandler.GetWatermarkFreeURL)
	}

	// 管理端点（JWT 认证）
	adminHandler := NewAdminHandler(cfg.DB, cfg.Manager, cfg.TaskStore, cfg.Settings, cfg.Version)
	admin := r.Group("/admin", AdminAuthMiddleware(cfg.JWTSecret))
	{
		// ── 所有已登录用户（admin + viewer）可访问 ──

		// 当前用户信息
		admin.GET("/me", meHandler())

		// 角色库只读接口
		admin.GET("/characters", adminHandler.ListCharacters)
		admin.GET("/characters/:id", adminHandler.GetCharacterAdmin)
		admin.GET("/characters/:id/image", adminHandler.GetCharacterImage)

		// 任务接口（viewer 只能看到自己 API Key 创建的任务，handler 内部过滤）
		admin.GET("/tasks", adminHandler.ListTasks)
		admin.GET("/tasks/:id", adminHandler.GetTask)
		admin.GET("/tasks/:id/content", adminHandler.DownloadTaskContent)

		// ── 仅管理员可访问 ──
		adminOnly := admin.Group("", AdminOnlyMiddleware())

		adminOnly.GET("/dashboard", adminHandler.GetDashboard)

		// 系统设置
		adminOnly.GET("/settings", adminHandler.GetSettings)
		adminOnly.PUT("/settings", adminHandler.UpdateSettings)
		adminOnly.POST("/proxy-test", adminHandler.TestProxy)

		// 版本管理
		adminOnly.GET("/version", adminHandler.GetVersion)
		adminOnly.POST("/upgrade", adminHandler.TriggerUpgrade)

		// API Key 管理
		adminOnly.GET("/api-keys", adminHandler.ListAPIKeys)
		adminOnly.POST("/api-keys", adminHandler.CreateAPIKey)
		adminOnly.PUT("/api-keys/:id", adminHandler.UpdateAPIKey)
		adminOnly.DELETE("/api-keys/:id", adminHandler.DeleteAPIKey)
		adminOnly.GET("/api-keys/:id/reveal", adminHandler.RevealAPIKey)

		// 账号组管理
		adminOnly.GET("/groups", adminHandler.ListGroups)
		adminOnly.POST("/groups", adminHandler.CreateGroup)
		adminOnly.PUT("/groups/:id", adminHandler.UpdateGroup)
		adminOnly.DELETE("/groups/:id", adminHandler.DeleteGroup)

		// 账号管理
		adminOnly.GET("/accounts", adminHandler.ListAllAccounts)
		adminOnly.POST("/accounts/batch", adminHandler.BatchImportAccounts)
		adminOnly.POST("/accounts", adminHandler.CreateAccountDirect)
		adminOnly.PUT("/accounts/:id", adminHandler.UpdateAccountDirect)
		adminOnly.DELETE("/accounts/:id", adminHandler.DeleteAccountDirect)
		adminOnly.POST("/accounts/:id/refresh", adminHandler.RefreshAccountTokenDirect)
		adminOnly.GET("/accounts/:id/status", adminHandler.GetAccountStatusDirect)
		adminOnly.GET("/accounts/:id/tokens", adminHandler.RevealAccountTokens)

		// 角色管理（写操作仅管理员）
		adminOnly.POST("/characters/:id/visibility", adminHandler.ToggleCharacterVisibility)
		adminOnly.DELETE("/characters/:id", adminHandler.DeleteCharacterAdmin)
	}

	return r
}

// loginHandler 管理员登录
func loginHandler(jwtSecret, adminUser, adminPass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		if req.Username != adminUser || req.Password != adminPass {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		token, err := GenerateJWT(jwtSecret, req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token, "role": RoleAdmin})
	}
}

// apiKeyLoginHandler API Key 登录（获得 viewer 角色的 JWT）
func apiKeyLoginHandler(jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			APIKey string `json:"api_key" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		var apiKey model.SoraAPIKey
		if err := db.Where("key = ? AND enabled = ?", req.APIKey, true).First(&apiKey).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key 无效或已禁用"})
			return
		}

		token, err := GenerateJWTWithRole(jwtSecret, apiKey.Name, RoleViewer, apiKey.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token, "role": RoleViewer})
	}
}

// meHandler 获取当前用户信息（角色）
func meHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"username": username,
			"role":     role,
		})
	}
}
