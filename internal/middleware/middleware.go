package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitee.com/dinglide/spot-vm/internal/config" // "gitee.com/dinglide/spot-vm/internal/database"

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

// RateLimitMiddleware 基于令牌桶算法的速率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
	// 简化版令牌桶：每个IP每秒最多10个请求
	type visitor struct {
		tokens    int
		lastReset time.Time
	}
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	const (
		maxTokens  = 10
		refillRate = 10 // 每秒补充的令牌数
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			v = &visitor{tokens: maxTokens, lastReset: time.Now()}
			visitors[ip] = v
		}

		// 补充令牌
		now := time.Now()
		elapsed := now.Sub(v.lastReset).Seconds()
		v.tokens += int(elapsed * float64(refillRate))
		if v.tokens > maxTokens {
			v.tokens = maxTokens
		}
		v.lastReset = now

		if v.tokens <= 0 {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded. Please try again later.",
				"error_code": "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		v.tokens--
		mu.Unlock()

		c.Next()
	}
}

// SplitToken 分割 Authorization header
func SplitToken(authHeader string) []string {
	return strings.Split(authHeader, " ")
}

// AuditMiddleware 审计日志中间件，记录所有写操作
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 仅记录写操作（POST/PUT/DELETE）的审计日志
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "DELETE" {
			latency := time.Since(startTime)
			statusCode := c.Writer.Status()
			clientIP := c.ClientIP()
			path := c.Request.URL.Path
			userAgent := c.Request.UserAgent()

			log.Printf("📝 [AUDIT] %s %s | status=%d | ip=%s | latency=%v | ua=%s",
				method, path, statusCode, clientIP, latency, userAgent)
		}
	}
}
