# Quickstart: Cloud Spot VM

**Feature**: 001-codebase-logic-spec  
**Date**: 2026-03-31

## Prerequisites

- Go 1.24.5+
- 腾讯云账号，已创建 API 密钥（SecretId / SecretKey）
- 腾讯云 SSL 证书 ID（如需 Nginx SSL 功能）
- 域名（如需 DNS 管理功能）

## 1. 克隆项目

```bash
git clone https://gitee.com/dinglide/spot-vm.git
cd spot-vm
```

## 2. 配置环境变量

创建 `.env` 文件：

```bash
cat > .env << 'EOF'
# 运行环境
ENVIRONMENT=development
PORT=8080

# 认证
API_KEY=your-api-key-here
JWT_SECRET=your-jwt-secret-here

# 腾讯云凭证
TENCENTCLOUD_SECRET_ID=your-secret-id
TENCENTCLOUD_SECRET_KEY=your-secret-key

# 可选配置
REGION=ap-hongkong
DOMAIN=example.com
CERTIFICATE_ID=your-cert-id
EOF
```

## 3. 安装依赖

```bash
go mod download
```

## 4. 启动服务

```bash
go run cmd/spot-manager/main.go
```

启动后会看到：
```
🚀 启动Spot VM自动管理服务...
📍 监听端口: 8080
🔑 API Key: your-api-key-here
🌐 访问地址: http://localhost:8080
🤖 自动管理器: 已启动
```

## 5. 验证服务

```bash
# 健康检查（无需认证）
curl http://localhost:8080/api/v1/health

# 查看当前实例状态
curl -H 'X-API-Key: your-api-key-here' \
  http://localhost:8080/api/v1/spot-vm/current/status

# 查询最便宜的 Spot 实例
curl -H 'X-API-Key: your-api-key-here' \
  http://localhost:8080/api/v1/spot-vm/cheapest

# 查看自动管理器状态
curl -H 'X-API-Key: your-api-key-here' \
  http://localhost:8080/api/v1/spot-vm/auto-manager/status
```

## 6. 核心操作

### 修改目标 Region

```bash
curl -X PUT \
  -H 'X-API-Key: your-api-key-here' \
  -H 'Content-Type: application/json' \
  -d '{"region":"ap-singapore"}' \
  http://localhost:8080/api/v1/spot-vm/target-region
```

### 手动触发回收替换

```bash
curl -X POST \
  -H 'X-API-Key: your-api-key-here' \
  http://localhost:8080/api/v1/spot-vm/trigger-termination
```

### 模拟终止（测试用）

```bash
curl -X POST \
  -H 'X-API-Key: your-api-key-here' \
  http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination
```

### 创建最便宜实例（DryRun 预检）

```bash
curl -X POST \
  -H 'X-API-Key: your-api-key-here' \
  -H 'Content-Type: application/json' \
  -d '{"dry_run": true}' \
  http://localhost:8080/api/v1/spot-vm/create-cheapest
```

## 7. 架构概览

```
用户 → REST API (Gin) → Handler → TCC Client → 腾讯云 API
                                      ↓
                              AutoManager (后台)
                                      ↓
                          每10秒检查 metadata API
                                      ↓
                          检测到回收 → 创建替换实例
```

## 注意事项

- 系统需要运行在腾讯云 CVM 实例上才能访问 metadata 服务
- 在非腾讯云环境运行时，metadata 相关功能（实例状态、终止检测）将不可用
- 当前 Region、Domain、CertificateId 在 `cmd/spot-manager/main.go` 中硬编码，需要手动修改
- 实例创建密码硬编码为 `1qazZSE$`，生产环境请修改
