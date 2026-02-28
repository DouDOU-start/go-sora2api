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

// JWTClaims JWT 载荷
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT Token
func GenerateJWT(secret, username string) (string, error) {
	claims := JWTClaims{
		Username: username,
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

// AdminAuthMiddleware 管理端认证中间件（仅 JWT）
func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "缺少 Authorization 头",
			})
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization 格式错误，需要 Bearer Token",
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
		c.Next()
	}
}

// APIKeyAuthMiddleware /v1/ API 认证中间件（从数据库查询 API Keys）
func APIKeyAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否存在启用的 API Key，若无则跳过认证（开放模式）
		var count int64
		db.Model(&model.SoraAPIKey{}).Where("enabled = ?", true).Count(&count)
		if count == 0 {
			c.Next()
			return
		}

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

		// 更新使用统计
		now := time.Now()
		db.Model(&apiKey).Updates(map[string]interface{}{
			"usage_count": gorm.Expr("usage_count + 1"),
			"last_used_at": now,
		})

		// 将绑定的分组 ID 传入上下文，供调度器使用
		if apiKey.GroupID != nil {
			c.Set("api_key_group_id", *apiKey.GroupID)
		}

		c.Next()
	}
}
