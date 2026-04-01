# Quickstart: Spot CVM 自动迁移与自愈

**Feature**: 002-spot-vm-auto-migration
**Date**: 2026-04-01

## 前置条件

1. Go 1.24.5+ 已安装
2. 腾讯云 API 密钥已创建（SecretId / SecretKey）
3. 腾讯云安全组已开放端口：22（SSH）、80（HTTP）、443（HTTPS）、8080（API）
4. 域名已在腾讯云 DNSPod 托管（可选，用于 DNS 自动更新）
5. SSL 证书已上传到腾讯云证书管理（可选，用于 HTTPS）

## 快速开始

### 1. 克隆项目

```bash
git clone https://gitee.com/dinglide/spot-vm.git
cd spot-vm
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```env
# 腾讯云凭证
TENCENTCLOUD_SECRET_ID=your-secret-id
TENCENTCLOUD_SECRET_KEY=your-secret-key

# 服务配置
PORT=8080
API_KEY=your-api-key
ENVIRONMENT=production

# Region 配置
REGION=ap-hongkong

# 域名和证书（可选）
DOMAIN=example.com
CERTIFICATE_ID=your-cert-id

# 实例配置
IMAGE_ID=img-hdt9xxkt
INSTANCE_PASSWORD=YourSecurePassword123!
DISK_TYPE=CLOUD_BSSD
DISK_SIZE=20
BANDWIDTH=10

# 迁移配置（新增）
SSH_PORT=22
SSH_TIMEOUT=10
SSH_WAIT_TIMEOUT=180
MIGRATION_MAX_RETRIES=3
REMOTE_BINARY_PATH=/opt/spot-manager/spot-manager
REMOTE_ENV_PATH=/opt/spot-manager/.env
```

### 3. 编译

```bash
go build -o spot-manager ./cmd/spot-manager/
```

### 4. 运行

```bash
./spot-manager
```

### 5. 验证

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 查看当前状态
curl -H 'X-API-Key: your-api-key' http://localhost:8080/api/v1/spot-vm/current/status

# 查看自动管理器状态
curl -H 'X-API-Key: your-api-key' http://localhost:8080/api/v1/spot-vm/auto-manager/status
```

## 常用操作

### 切换目标 Region

```bash
# 查看所有可用 Region
curl -H 'X-API-Key: your-api-key' http://localhost:8080/api/v1/spot-vm/regions

# 切换到新加坡（Zone 由系统自动选择）
curl -X PUT \
  -H 'X-API-Key: your-api-key' \
  -H 'Content-Type: application/json' \
  -d '{"region":"ap-singapore"}' \
  http://localhost:8080/api/v1/spot-vm/target-region
```

### 模拟终止（测试迁移流程）

```bash
curl -X POST \
  -H 'X-API-Key: your-api-key' \
  http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination
```

这将触发完整的替换 + 迁移流程：
1. 在目标 Region 中自动选择最便宜的 Zone 和实例类型
2. 创建替代实例
3. 等待 SSH 就绪
4. SCP 传输程序文件
5. 远程启动程序
6. 更新 DNS 记录
7. 部署 Nginx

### 手动触发回收

```bash
curl -X POST \
  -H 'X-API-Key: your-api-key' \
  http://localhost:8080/api/v1/spot-vm/trigger-termination
```

## 部署到腾讯云 Spot CVM

### 1. 交叉编译

```bash
GOOS=linux GOARCH=amd64 go build -o spot-manager ./cmd/spot-manager/
```

### 2. 上传到实例

```bash
scp spot-manager root@<instance-ip>:/opt/spot-manager/
scp .env root@<instance-ip>:/opt/spot-manager/
```

### 3. 在实例上启动

```bash
ssh root@<instance-ip>
cd /opt/spot-manager
nohup ./spot-manager > /var/log/spot-manager.log 2>&1 &
```

程序启动后将自动：
- 每 10 秒检测 Spot 实例回收状态
- 检测到回收时自动创建替代实例并迁移自身
- 新实例上的程序继续监控，形成自愈闭环
