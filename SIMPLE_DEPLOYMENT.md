# Spot VM 自动管理 - 简化部署指南

## 概述

这是一个精简的Spot VM自动管理程序，专门设计用于在Spot实例上运行，当检测到当前实例即将被回收时，自动创建新的最便宜Spot实例。

## 核心功能

- 🔍 **自动检测**：每10秒检查一次实例回收状态
- 🚀 **自动替换**：检测到回收时自动创建新实例
- 💰 **成本优化**：始终选择最便宜的实例类型
- 🧪 **测试支持**：提供模拟终止API进行测试

## 文件结构

```
cmd/spot-manager/main.go          # 主程序入口
internal/tcc/spot_vm/simple_auto_manager.go  # 简化的自动管理器
internal/handlers/simple_spot_vm.go          # 简化的API处理器
internal/routes/simple_routes.go             # 简化的路由配置
```

## 快速部署

### 1. 编译程序

```bash
go build -o spot-auto-manager cmd/spot-manager/main.go
```

### 2. 在Spot实例上运行

```bash
./spot-auto-manager
```

### 3. 程序会自动：

- 启动HTTP API服务器（端口8080）
- 启动自动管理器
- 开始监控实例回收状态

## API 端点

### 基础端点

- `GET /api/v1/health` - 健康检查
- `POST /api/v1/auth/validate` - API key验证

### 自动管理端点

- `GET /api/v1/spot-vm/auto-manager/status` - 获取状态
- `POST /api/v1/spot-vm/auto-manager/start` - 启动管理器
- `POST /api/v1/spot-vm/auto-manager/stop` - 停止管理器
- `POST /api/v1/spot-vm/auto-manager/simulate-termination` - 模拟终止

## 使用示例

### 检查状态

```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/status
```

### 模拟测试

```bash
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination
```

## 配置说明

### 环境变量

在 `.env` 文件中设置：

```bash
# 腾讯云API密钥
TENCENTCLOUD_SECRET_ID=your-secret-id
TENCENTCLOUD_SECRET_KEY=your-secret-key

# API服务器配置
API_KEY=your-super-secret-api-key-change-in-production
PORT=8080
```

### Region配置

在 `cmd/spot-manager/main.go` 中修改：

```go
region := "sa-saopaulo" // 修改为你的目标Region
```

### 实例配置

在 `internal/tcc/spot_vm/manager.go` 中配置：

```go
"VirtualPrivateCloud": map[string]any{
    "VpcId":    "your-vpc-id",
    "SubnetId": "your-subnet-id",
},
"ImageId": "your-image-id",
```

## 工作流程

```
1. 启动程序
   ↓
2. 自动管理器开始监控
   ↓
3. 每10秒检查回收状态
   ↓
4. 检测到回收信号
   ↓
5. 获取最便宜实例信息
   ↓
6. 创建新的Spot实例
   ↓
7. 记录创建结果
   ↓
8. 继续监控
```

## 测试

### 运行测试脚本

```bash
./test_simple_auto_manager.sh
```

### 手动测试

```bash
# 1. 启动程序
go run cmd/spot-manager/main.go

# 2. 检查状态
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/status

# 3. 模拟终止
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination
```

## 日志输出

程序会输出详细的日志信息：

```
🚀 启动Spot VM自动管理服务...
📍 监听端口: 8080
🔑 API Key: your-super-secret-api-key-change-in-production
🌐 访问地址: http://localhost:8080
🤖 自动管理器: 已启动
🚀 启动Spot VM自动管理器...
✅ Spot VM自动管理器启动成功
⚠️  检测到Spot实例即将被回收！
🔄 开始创建替换实例...
💰 选择最便宜实例: S5.MEDIUM2 (¥0.0057/小时)
✅ 替换实例创建成功: ins-8hl0lim3
🎉 新实例已启动，当前实例即将被回收
```

## 生产环境部署

### 1. 系统服务

创建systemd服务文件 `/etc/systemd/system/spot-auto-manager.service`：

```ini
[Unit]
Description=Spot VM Auto Manager
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/spot-vm
ExecStart=/path/to/spot-vm/spot-auto-manager
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 2. 启动服务

```bash
sudo systemctl daemon-reload
sudo systemctl enable spot-auto-manager
sudo systemctl start spot-auto-manager
```

### 3. 查看状态

```bash
sudo systemctl status spot-auto-manager
sudo journalctl -u spot-auto-manager -f
```

## 监控和告警

### 1. 健康检查

```bash
# 检查程序是否运行
curl http://localhost:8080/api/v1/health

# 检查自动管理器状态
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/status
```

### 2. 日志监控

```bash
# 实时查看日志
tail -f /var/log/spot-auto-manager.log

# 查看系统服务日志
sudo journalctl -u spot-auto-manager -f
```

## 故障排除

### 常见问题

1. **程序无法启动**
   - 检查端口是否被占用
   - 验证配置文件
   - 查看错误日志

2. **自动管理器不工作**
   - 检查API密钥权限
   - 验证Region配置
   - 确认VPC和子网配置

3. **实例创建失败**
   - 检查腾讯云配额
   - 验证镜像ID
   - 查看腾讯云控制台日志

### 调试命令

```bash
# 检查端口占用
netstat -tlnp | grep 8080

# 检查进程
ps aux | grep spot-auto-manager

# 查看日志
tail -f server.log
```

## 安全建议

1. **API密钥安全**
   - 使用强密钥
   - 定期轮换
   - 限制访问权限

2. **网络安全**
   - 配置防火墙
   - 使用HTTPS（生产环境）
   - 限制API访问IP

3. **实例安全**
   - 使用强密码
   - 配置安全组
   - 定期更新镜像

## 总结

这个简化版本专注于核心功能：

- ✅ **精简代码**：移除了不必要的功能
- ✅ **专注核心**：只保留自动检测和替换功能
- ✅ **易于部署**：简单的单文件部署
- ✅ **成本优化**：自动选择最便宜实例
- ✅ **高可用性**：确保服务连续性

适合在Spot实例上运行，实现自动化的实例替换管理。
