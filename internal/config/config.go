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

	// 实例创建相关配置
	ImageId          string // 镜像ID
	InstancePassword string // 实例登录密码
	DiskType         string // 系统盘类型
	DiskSize         int    // 系统盘大小（GB）
	Bandwidth        int    // 公网带宽（Mbps）

	// SSH 和迁移相关配置
	SSHPort             int    // SSH 端口（默认 22）
	SSHTimeout          int    // SSH 连接超时（秒，默认 10）
	SSHWaitTimeout      int    // 等待 SSH 就绪超时（秒，默认 180）
	MigrationMaxRetries int    // 迁移最大重试次数（默认 3）
	RemoteBinaryPath    string // 远程二进制文件路径
	RemoteEnvPath       string // 远程 .env 文件路径
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

		// 实例创建相关配置
		ImageId:          getEnv("IMAGE_ID", "img-hdt9xxkt"),
		InstancePassword: getEnv("INSTANCE_PASSWORD", ""),
		DiskType:         getEnv("DISK_TYPE", "CLOUD_BSSD"),
		DiskSize:         getEnvAsInt("DISK_SIZE", 20),
		Bandwidth:        getEnvAsInt("BANDWIDTH", 10),

		// SSH 和迁移相关配置
		SSHPort:             getEnvAsInt("SSH_PORT", 22),
		SSHTimeout:          getEnvAsInt("SSH_TIMEOUT", 10),
		SSHWaitTimeout:      getEnvAsInt("SSH_WAIT_TIMEOUT", 180),
		MigrationMaxRetries: getEnvAsInt("MIGRATION_MAX_RETRIES", 3),
		RemoteBinaryPath:    getEnv("REMOTE_BINARY_PATH", "/opt/spot-manager/spot-manager"),
		RemoteEnvPath:       getEnv("REMOTE_ENV_PATH", "/opt/spot-manager/.env"),
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
