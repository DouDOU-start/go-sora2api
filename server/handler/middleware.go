package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// 用户角色常量
const (
	RoleAdmin  = "admin"  // 管理员：拥有所有权限
	RoleViewer = "viewer" // 只读用户（API Key 登录）：只能查看文档和角色库
)

// JWTClaims JWT 载荷
type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`       // admin / viewer
	APIKeyID int64  `json:"api_key_id"` // viewer 角色对应的 API Key ID（admin 为 0）
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT Token（默认管理员角色）
func GenerateJWT(secret, username string) (string, error) {
	return GenerateJWTWithRole(secret, username, RoleAdmin, 0)
}

// GenerateJWTWithRole 生成指定角色的 JWT Token
func GenerateJWTWithRole(secret, username, role string, apiKeyID int64) (string, error) {
	claims := JWTClaims{
		Username: username,
		Role:     role,
		APIKeyID: apiKeyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT 验证 JWT Token
func ValidateJWT(secret, tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// AdminAuthMiddleware 管理端认证中间件（JWT，支持 Header 和 query 参数）
func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 优先从 Authorization Header 读取
		auth := c.GetHeader("Authorization")
		if auth != "" {
			token = strings.TrimPrefix(auth, "Bearer ")
			if token == auth {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization 格式错误，需要 Bearer Token",
				})
				return
			}
		}

		// 其次从 query 参数读取（用于 img 标签等无法设置 Header 的场景）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证凭据",
			})
			return
		}

		claims, err := ValidateJWT(jwtSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token 无效或已过期",
			})
			return
		}

		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("jwt_api_key_id", claims.APIKeyID)
		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员可访问的中间件（需放在 AdminAuthMiddleware 之后）
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "权限不足，此操作仅限管理员",
			})
			return
		}
		c.Next()
	}
}

// APIKeyAuthMiddleware /v1/ API 认证中间件（从数据库查询 API Keys）
func APIKeyAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": &model.TaskErrorInfo{Message: "缺少 Authorization 头"},
			})
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": &model.TaskErrorInfo{Message: "Authorization 格式错误，需要 Bearer Token"},
			})
			return
		}

		var apiKey model.SoraAPIKey
		if err := db.Where("key = ? AND enabled = ?", token, true).First(&apiKey).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": &model.TaskErrorInfo{Message: "无效的 API Key"},
			})
			return
		}

		// API Key 必须绑定分组才能调用
		if apiKey.GroupID == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": &model.TaskErrorInfo{Message: "此 API Key 未绑定分组，无法调用接口，请先在管理后台绑定分组"},
			})
			return
		}

		// 更新使用统计
		now := time.Now()
		db.Model(&apiKey).Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": now,
		})

		// 将 API Key 信息传入上下文
		c.Set("api_key_id", apiKey.ID)
		c.Set("api_key_group_id", *apiKey.GroupID)

		c.Next()
	}
}
