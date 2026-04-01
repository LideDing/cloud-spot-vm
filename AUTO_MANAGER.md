# Spot VM 自动管理功能

## 概述

Spot VM自动管理功能可以自动检测Spot实例的回收信号，并在检测到回收时自动创建新的最便宜实例，确保服务的连续性和成本优化。

## 功能特性

### 🔍 自动检测
- **回收监控**：每10秒检查一次实例是否即将被回收
- **实时响应**：检测到回收信号时立即触发替换流程
- **多实例支持**：支持监控多个实例的回收状态

### 🚀 自动替换
- **智能选择**：自动选择整个Region中最便宜的Spot实例
- **无缝切换**：在检测到回收时立即创建新实例
- **成本优化**：始终选择最便宜的实例类型

### 📊 状态管理
- **实时状态**：提供自动管理器的实时运行状态
- **历史记录**：记录所有替换实例的详细信息
- **控制接口**：支持启动、停止自动管理器

### 🧪 测试支持
- **模拟终止**：提供API模拟实例终止场景
- **安全测试**：在不影响生产环境的情况下测试功能
- **完整验证**：验证整个自动替换流程

## API 端点

### 状态管理

#### 获取自动管理器状态
```bash
GET /api/v1/spot-vm/auto-manager/status
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Retrieved auto manager status successfully",
  "data": {
    "is_running": true,
    "region": "ap-singapore",
    "replacement_count": 1,
    "last_check": "2025-09-05T16:59:21+08:00"
  }
}
```

#### 启动自动管理器
```bash
POST /api/v1/spot-vm/auto-manager/start
```

#### 停止自动管理器
```bash
POST /api/v1/spot-vm/auto-manager/stop
```

### 测试功能

#### 模拟实例终止
```bash
POST /api/v1/spot-vm/auto-manager/simulate-termination
```

**用途：**
- 测试自动替换功能
- 验证替换流程
- 不依赖真实的回收信号

### 历史记录

#### 获取替换历史
```bash
GET /api/v1/spot-vm/auto-manager/replacement-history
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Retrieved replacement history successfully",
  "history": {
    "ins-luj1r4py": {
      "InstanceType": "S5.MEDIUM2",
      "InstanceFamily": "S5",
      "TypeName": "标准型S5",
      "Cpu": 2,
      "Memory": 2,
      "CpuType": "Intel Xeon Cascade Lake 8255C/Intel Xeon Cooper Lake",
      "Frequency": "2.5GHz/3.1GHz",
      "Gpu": 0,
      "GpuCount": 0,
      "InstanceBandwidth": 1.5,
      "InstancePps": 30,
      "Status": "SELL",
      "StatusCategory": "EnoughStock",
      "Price": {
        "ChargeUnit": "HOUR",
        "UnitPrice": 0.25,
        "UnitPriceDiscount": 0.01,
        "Discount": 4
      },
      "Zone": "ap-singapore-1"
    }
  },
  "total": 1
}
```

## 工作流程

### 1. 启动阶段
```
启动服务器 → 创建TCC实例 → 启动自动管理器 → 开始监控
```

### 2. 监控阶段
```
每10秒检查 → 查询回收状态 → 判断是否需要替换
```

### 3. 替换阶段
```
检测到回收 → 获取最便宜实例 → 创建新实例 → 记录历史
```

### 4. 完成阶段
```
新实例创建成功 → 更新状态 → 继续监控
```

## 使用示例

### 基本使用

```bash
# 1. 启动服务器（自动管理器会自动启动）
go run cmd/main.go

# 2. 检查自动管理器状态
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/status

# 3. 模拟实例终止（测试用）
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination

# 4. 查看替换历史
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/replacement-history
```

### 测试脚本

```bash
# 运行完整的自动管理测试
./test_auto_manager.sh
```

## 配置说明

### 检查频率
默认每10秒检查一次，可以在 `auto_manager.go` 中修改：

```go
ticker := time.NewTicker(10 * time.Second) // 修改这个值
```

### Region配置
在 `main.go` 中配置目标Region：

```go
region := "ap-singapore" // 修改为你的目标Region
```

### 实例配置
在 `manager.go` 中配置实例创建参数：

```go
"VirtualPrivateCloud": map[string]any{
    "VpcId":    "your-vpc-id",
    "SubnetId": "your-subnet-id",
},
```

## 监控和日志

### 日志输出
自动管理器会输出详细的日志信息：

```
🚀 启动Spot VM自动管理器...
✅ Spot VM自动管理器启动成功
⚠️  检测到Spot实例即将被回收！
🔄 开始创建替换实例...
💰 选择最便宜实例: S5.MEDIUM2 (¥0.0100/小时)
✅ 替换实例创建成功: ins-luj1r4py
🎉 替换实例处理完成: S5.MEDIUM2
```

### 状态监控
通过API可以实时监控自动管理器的状态：

- `is_running`: 是否正在运行
- `region`: 目标Region
- `replacement_count`: 替换次数
- `last_check`: 最后检查时间

## 最佳实践

### 1. 生产环境部署
- 确保配置正确的VPC和子网
- 设置合适的检查频率
- 监控替换历史

### 2. 成本优化
- 自动选择最便宜实例
- 定期检查替换频率
- 监控实例使用情况

### 3. 安全考虑
- 使用强API密钥
- 限制API访问权限
- 定期轮换密钥

### 4. 故障处理
- 监控自动管理器状态
- 设置告警机制
- 准备手动干预方案

## 故障排除

### 常见问题

1. **自动管理器未启动**
   - 检查服务器日志
   - 确认TCC实例创建成功
   - 验证Region配置

2. **替换实例创建失败**
   - 检查VPC和子网配置
   - 验证API密钥权限
   - 查看腾讯云控制台日志

3. **监控不工作**
   - 确认实例在正确的Region
   - 检查网络连接
   - 验证元数据服务

### 调试方法

```bash
# 查看服务器日志
tail -f server.log

# 检查自动管理器状态
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/status

# 手动触发测试
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination
```

## 总结

Spot VM自动管理功能提供了完整的实例生命周期管理，包括：

- ✅ **自动检测**：实时监控回收信号
- ✅ **智能替换**：自动选择最便宜实例
- ✅ **状态管理**：完整的控制接口
- ✅ **测试支持**：安全的测试环境
- ✅ **历史记录**：详细的替换记录

这个功能确保了Spot实例的高可用性和成本优化，是生产环境中的重要组件。
