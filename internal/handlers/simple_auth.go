package handlers

import (
	"net/http"

	"gitee.com/dinglide/spot-vm/internal/config"
	"github.com/gin-gonic/gin"
)

// SimpleAuthHandler 简化认证处理器（使用API key）
type SimpleAuthHandler struct {
	cfg *config.Config
}

// NewSimpleAuthHandler 创建简化认证处理器
func NewSimpleAuthHandler(cfg *config.Config) *SimpleAuthHandler {
	return &SimpleAuthHandler{cfg: cfg}
}

// APIKeyRequest API key验证请求结构
type APIKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// APIKeyResponse API key验证响应结构
type APIKeyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ValidateAPIKey 验证API key
func (h *SimpleAuthHandler) ValidateAPIKey(c *gin.Context) {
	var req APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证API key
	if req.APIKey != h.cfg.APIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	c.JSON(http.StatusOK, APIKeyResponse{
		Status:  "success",
		Message: "API key is valid",
	})
}

// Health 健康检查端点
func (h *SimpleAuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "API server is running",
	})
}
