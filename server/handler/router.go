package handler

import (
	"net/http"

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
	adminHandler := NewAdminHandler(cfg.DB, cfg.Manager, cfg.TaskStore, cfg.Settings)
	admin := r.Group("/admin", AdminAuthMiddleware(cfg.JWTSecret))
	{
		admin.GET("/dashboard", adminHandler.GetDashboard)

		// 系统设置
		admin.GET("/settings", adminHandler.GetSettings)
		admin.PUT("/settings", adminHandler.UpdateSettings)
		admin.POST("/proxy-test", adminHandler.TestProxy)

		// API Key 管理
		admin.GET("/api-keys", adminHandler.ListAPIKeys)
		admin.POST("/api-keys", adminHandler.CreateAPIKey)
		admin.PUT("/api-keys/:id", adminHandler.UpdateAPIKey)
		admin.DELETE("/api-keys/:id", adminHandler.DeleteAPIKey)
		admin.GET("/api-keys/:id/reveal", adminHandler.RevealAPIKey)

		// 账号组管理
		admin.GET("/groups", adminHandler.ListGroups)
		admin.POST("/groups", adminHandler.CreateGroup)
		admin.PUT("/groups/:id", adminHandler.UpdateGroup)
		admin.DELETE("/groups/:id", adminHandler.DeleteGroup)

		// 账号管理
		admin.GET("/accounts", adminHandler.ListAllAccounts)
		admin.POST("/accounts", adminHandler.CreateAccountDirect)
		admin.PUT("/accounts/:id", adminHandler.UpdateAccountDirect)
		admin.DELETE("/accounts/:id", adminHandler.DeleteAccountDirect)
		admin.POST("/accounts/:id/refresh", adminHandler.RefreshAccountTokenDirect)
		admin.GET("/accounts/:id/status", adminHandler.GetAccountStatusDirect)
		admin.GET("/accounts/:id/tokens", adminHandler.RevealAccountTokens)

		// 任务管理
		admin.GET("/tasks", adminHandler.ListTasks)
		admin.GET("/tasks/:id", adminHandler.GetTask)
		admin.GET("/tasks/:id/content", adminHandler.DownloadTaskContent)
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

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
