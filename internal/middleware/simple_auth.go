package middleware

import (
	"log"
	"net/http"
	"strings"

	"gitee.com/dinglide/spot-vm/internal/config"
	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware API key认证中间件
func APIKeyAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// 尝试从Authorization头获取
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					apiKey = parts[1]
				}
			}
		}

		// 如果仍然没有API key，尝试从查询参数获取
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		// 验证API key
		if apiKey == "" || apiKey != cfg.APIKey {
			// T018: 记录认证失败的详细日志
			authMethod := "未提供"
			if c.GetHeader("X-API-Key") != "" {
				authMethod = "X-API-Key"
			} else if c.GetHeader("Authorization") != "" {
				authMethod = "Authorization Bearer"
			} else if c.Query("api_key") != "" {
				authMethod = "Query Parameter"
			}
			log.Printf("⚠️  API Key认证失败: ip=%s, method=%s, path=%s, auth_method=%s",
				c.ClientIP(), c.Request.Method, c.Request.URL.Path, authMethod)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
			c.Abort()
			return
		}

		// 将认证状态存储到上下文中
		c.Set("authenticated", true)
		c.Next()
	}
}

// OptionalAPIKeyAuthMiddleware 可选的API key认证中间件（不强制要求）
func OptionalAPIKeyAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// 尝试从Authorization头获取
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					apiKey = parts[1]
				}
			}
		}

		// 如果仍然没有API key，尝试从查询参数获取
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		// 验证API key（如果提供了的话）
		if apiKey != "" {
			if apiKey == cfg.APIKey {
				c.Set("authenticated", true)
			} else {
				c.Set("authenticated", false)
			}
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// RequireAuth 要求认证的中间件
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
