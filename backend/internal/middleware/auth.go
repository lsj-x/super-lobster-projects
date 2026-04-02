package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig 认证配置
type AuthConfig struct {
	APIKey       string
	ReadOnlyMethods []string
}

// DefaultAuthConfig 返回默认认证配置
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		APIKey:        "", // 从环境变量读取
		ReadOnlyMethods: []string{"GET", "HEAD"},
	}
}

// Auth 认证中间件 - 支持 Bearer Token 和 API Key
func Auth(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为只读方法
		if isReadOnlyMethod(c.Request.Method, config.ReadOnlyMethods) {
			c.Next()
			return
		}

		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少 Authorization header",
			})
			c.Abort()
			return
		}

		// 检查 Bearer Token
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "无效的 Bearer Token",
				})
				c.Abort()
				return
			}
			// TODO: 验证 Bearer Token (可以根据需要添加 JWT 验证等)
			c.Next()
			return
		}

		// 检查 API Key
		if config.APIKey != "" && authHeader == config.APIKey {
			c.Next()
			return
		}

		// API Key 认证失败
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "无效的 API Key",
		})
		c.Abort()
	}
}

// ReadOnly 只读操作免认证中间件
func ReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只读方法无需认证
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
			"error": "只允许只读操作",
		})
	}
}

// RequireAuth 需要认证的中间件
func RequireAuth(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少 Authorization header",
			})
			c.Abort()
			return
		}

		// 检查 Bearer Token
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "无效的 Bearer Token",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// 检查 API Key
		if config.APIKey != "" && authHeader == config.APIKey {
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "认证失败",
		})
		c.Abort()
	}
}

// isReadOnlyMethod 检查是否为只读方法
func isReadOnlyMethod(method string, readOnlyMethods []string) bool {
	for _, m := range readOnlyMethods {
		if method == m {
			return true
		}
	}
	return false
}
