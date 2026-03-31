package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleUserHandler 简化用户处理器（不依赖数据库）
type SimpleUserHandler struct{}

// NewSimpleUserHandler 创建简化用户处理器
func NewSimpleUserHandler() *SimpleUserHandler {
	return &SimpleUserHandler{}
}

// SimpleUserProfile 简化的用户资料结构
type SimpleUserProfile struct {
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// GetSimpleProfile 获取简化用户资料
func (h *SimpleUserHandler) GetSimpleProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	usernameStr := username.(string)
	role := "user"
	if usernameStr == "admin" {
		role = "admin"
	}

	profile := SimpleUserProfile{
		Username:  usernameStr,
		Role:      role,
		CreatedAt: time.Now().Add(-24 * time.Hour), // 模拟创建时间
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateSimpleProfile 更新简化用户资料
func (h *SimpleUserHandler) UpdateSimpleProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 在简化版本中，我们只是返回成功消息
	// 实际的数据更新可以在客户端处理
	c.JSON(http.StatusOK, gin.H{
		"message":      "Profile updated successfully",
		"username":     username,
		"display_name": req.DisplayName,
	})
}

// GetSimpleStats 获取简化统计信息
func (h *SimpleUserHandler) GetSimpleStats(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 返回模拟的统计信息
	stats := gin.H{
		"username": username,
		"stats": gin.H{
			"total_requests": 100,
			"success_rate":   95.5,
			"last_login":     time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, stats)
}
