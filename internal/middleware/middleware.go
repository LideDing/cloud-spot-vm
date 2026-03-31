package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitee.com/dinglide/spot-vm/internal/config"
	// "gitee.com/dinglide/spot-vm/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims JWT 声明结构
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 检查 Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		claims := &JWTClaims{}

		// 解析 JWT token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 检查 token 是否在黑名单中（简化版本，不依赖数据库）
		// var session models.Session
		// if err := database.GetDB().Where("token = ? AND expires_at > ?", tokenString, time.Now()).First(&session).Error; err != nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
		// 	c.Abort()
		// 	return
		// }

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RoleMiddleware 角色权限中间件
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		role := userRole.(string)
		hasPermission := false

		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 简单的速率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
	// 这里可以实现更复杂的速率限制逻辑
	// 暂时返回一个简单的中间件
	return func(c *gin.Context) {
		c.Next()
	}
}

// SplitToken 分割 Authorization header
func SplitToken(authHeader string) []string {
	return strings.Split(authHeader, " ")
}

// AuditMiddleware 审计日志中间件
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间（用于未来扩展）
		_ = time.Now()

		// 处理请求
		c.Next()

		// 记录审计日志（简化版本，不依赖数据库）
		// userID, _ := c.Get("user_id")
		// var userIDPtr *uuid.UUID
		// if userID != nil {
		// 	if id, ok := userID.(uuid.UUID); ok {
		// 		userIDPtr = &id
		// 	}
		// }

		// auditLog := models.AuditLog{
		// 	UserID:    userIDPtr,
		// 	Action:    c.Request.Method,
		// 	Resource:  c.Request.URL.Path,
		// 	IPAddress: c.ClientIP(),
		// 	UserAgent: c.Request.UserAgent(),
		// }

		// 异步记录审计日志（仅在数据库可用时）
		// go func() {
		// 	db := database.GetDB()
		// 	if db != nil {
		// 		if err := db.Create(&auditLog).Error; err != nil {
		// 			log.Printf("Failed to create audit log: %v", err)
		// 		}
		// 	}
		// }()
	}
}
