package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	Environment             string
	Port                    string
	JWTSecret               string
	APIKey                  string
	Region                  string
	Domain                  string
	CertificateId           string
	TENCENTCLOUD_SECRET_ID  string
	TENCENTCLOUD_SECRET_KEY string
}

// Load 加载配置
func Load() *Config {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		// 忽略 .env 文件不存在的错误，这是正常的
		_ = err
	}

	config := &Config{
		Environment:             getEnv("ENVIRONMENT", "development"),
		Port:                    getEnv("PORT", "8080"),
		JWTSecret:               getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		APIKey:                  getEnv("API_KEY", "your-super-secret-api-key-change-in-production"),
		Region:                  getEnv("REGION", ""),
		Domain:                  getEnv("DOMAIN", ""),
		CertificateId:           getEnv("CERTIFICATE_ID", ""),
		TENCENTCLOUD_SECRET_ID:  getEnv("TENCENTCLOUD_SECRET_ID", ""),
		TENCENTCLOUD_SECRET_KEY: getEnv("TENCENTCLOUD_SECRET_KEY", ""),
	}

	return config
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
