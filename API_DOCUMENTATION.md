# Spot VM 管理 API 文档

## 概述

这是一套完整的Spot VM管理API，提供当前实例状态查询、Region管理、自动替换等功能。

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **认证方式**: API Key
- **认证头**: `X-API-Key: your-api-key`
- **内容类型**: `application/json`

## API 端点

### 1. 健康检查

**GET** `/health`

无需认证的健康检查端点。

**响应示例:**
```json
{
  "status": "healthy",
  "message": "API server is running"
}
```

### 2. API Key 验证

**POST** `/auth/validate`

验证API Key是否有效。

**请求体:**
```json
{
  "api_key": "your-super-secret-api-key-change-in-production"
}
```

**响应示例:**
```json
{
  "status": "success",
  "message": "API key is valid"
}
```

### 3. 获取当前Spot机器状态

**GET** `/spot-vm/current/status`

获取当前Spot实例的详细状态信息。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Retrieved current instance status successfully",
  "data": {
    "instance_id": "ins-12345678",
    "instance_type": "S5.MEDIUM2",
    "zone": "sa-saopaulo-1",
    "region": "sa-saopaulo",
    "status": "RUNNING",
    "is_terminated": false,
    "private_ip": "10.0.0.100",
    "public_ip": "1.2.3.4",
    "created_time": "2025-09-05T17:21:43+08:00",
    "expired_time": "N/A"
  }
}
```

### 4. 查看当前Region

**GET** `/spot-vm/current/region`

获取当前实例所在的Region。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Retrieved current region successfully",
  "data": {
    "current_region": "sa-saopaulo"
  }
}
```

### 5. 修改目标Region

**PUT** `/spot-vm/target-region`

修改目标Region，下次创建新实例时会在指定的Region中创建。

**请求头:**
```
X-API-Key: your-api-key
Content-Type: application/json
```

**请求体:**
```json
{
  "region": "ap-beijing"
}
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Target region updated successfully",
  "data": {
    "current_region": "sa-saopaulo",
    "target_region": "ap-beijing"
  }
}
```

### 6. 手动触发回收

**POST** `/spot-vm/trigger-termination`

手动触发当前实例回收，用于测试自动替换功能。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Manual termination triggered successfully"
}
```

### 7. 获取自动管理器状态

**GET** `/spot-vm/auto-manager/status`

获取自动管理器的运行状态。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Retrieved auto manager status successfully",
  "data": {
    "is_running": true,
    "region": "sa-saopaulo",
    "target_region": "ap-beijing",
    "last_check": "2025-09-05T17:21:43+08:00"
  }
}
```

### 8. 启动自动管理器

**POST** `/spot-vm/auto-manager/start`

启动自动管理器。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Auto manager started successfully"
}
```

### 9. 停止自动管理器

**POST** `/spot-vm/auto-manager/stop`

停止自动管理器。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Auto manager stopped successfully"
}
```

### 10. 模拟实例终止

**POST** `/spot-vm/auto-manager/simulate-termination`

模拟实例终止，用于测试自动替换功能。

**请求头:**
```
X-API-Key: your-api-key
```

**响应示例:**
```json
{
  "status": "success",
  "message": "Termination simulation triggered successfully"
}
```

## 使用示例

### 使用 curl 测试

```bash
# 设置变量
API_KEY="your-super-secret-api-key-change-in-production"
BASE_URL="http://localhost:8080/api/v1"

# 1. 健康检查
curl "$BASE_URL/health"

# 2. 获取当前实例状态
curl -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/current/status"

# 3. 查看当前Region
curl -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/current/region"

# 4. 修改目标Region
curl -X PUT -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"region":"ap-beijing"}' \
  "$BASE_URL/spot-vm/target-region"

# 5. 手动触发回收
curl -X POST -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/trigger-termination"

# 6. 获取自动管理器状态
curl -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/auto-manager/status"
```

### 使用 JavaScript 测试

```javascript
const API_KEY = 'your-super-secret-api-key-change-in-production';
const BASE_URL = 'http://localhost:8080/api/v1';

// 获取当前实例状态
async function getCurrentStatus() {
  const response = await fetch(`${BASE_URL}/spot-vm/current/status`, {
    headers: {
      'X-API-Key': API_KEY
    }
  });
  return await response.json();
}

// 修改目标Region
async function setTargetRegion(region) {
  const response = await fetch(`${BASE_URL}/spot-vm/target-region`, {
    method: 'PUT',
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ region })
  });
  return await response.json();
}

// 手动触发回收
async function triggerTermination() {
  const response = await fetch(`${BASE_URL}/spot-vm/trigger-termination`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY
    }
  });
  return await response.json();
}
```

## 错误处理

### 常见错误码

- **400 Bad Request**: 请求格式错误
- **401 Unauthorized**: API Key无效或缺失
- **404 Not Found**: 端点不存在
- **500 Internal Server Error**: 服务器内部错误

### 错误响应格式

```json
{
  "error": "错误描述信息"
}
```

### 错误示例

```json
{
  "error": "Invalid or missing API key"
}
```

```json
{
  "error": "Invalid request format: Key: 'Region' Error:Field validation for 'Region' failed on the 'required' tag"
}
```

## 工作流程

### 自动替换流程

1. **监控阶段**: 自动管理器每10秒检查一次实例回收状态
2. **检测阶段**: 检测到回收信号时触发替换流程
3. **查询阶段**: 在目标Region中查询最便宜的实例类型
4. **创建阶段**: 创建新的Spot实例
5. **完成阶段**: 记录创建结果，继续监控

### Region切换流程

1. **设置目标**: 通过API设置目标Region
2. **验证配置**: 系统验证Region配置
3. **应用配置**: 下次创建实例时使用新Region
4. **状态更新**: 自动管理器状态中显示目标Region

## 配置说明

### 环境变量

```bash
# 腾讯云API密钥
TENCENTCLOUD_SECRET_ID=your-secret-id
TENCENTCLOUD_SECRET_KEY=your-secret-key

# API服务器配置
API_KEY=your-super-secret-api-key-change-in-production
PORT=8080
```

### 支持的Region

- `ap-beijing`: 北京
- `ap-shanghai`: 上海
- `ap-guangzhou`: 广州
- `ap-chengdu`: 成都
- `ap-chongqing`: 重庆
- `ap-singapore`: 新加坡
- `ap-mumbai`: 孟买
- `ap-seoul`: 首尔
- `ap-tokyo`: 东京
- `na-siliconvalley`: 硅谷
- `na-ashburn`: 弗吉尼亚
- `eu-frankfurt`: 法兰克福
- `sa-saopaulo`: 圣保罗

## 安全建议

1. **API Key安全**
   - 使用强密钥
   - 定期轮换
   - 不要在代码中硬编码

2. **网络安全**
   - 配置防火墙
   - 使用HTTPS（生产环境）
   - 限制访问IP

3. **实例安全**
   - 使用强密码
   - 配置安全组
   - 定期更新镜像

## 监控和告警

### 健康检查

```bash
# 检查API服务状态
curl http://localhost:8080/api/v1/health

# 检查自动管理器状态
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/spot-vm/auto-manager/status
```

### 日志监控

```bash
# 查看服务器日志
tail -f server.log

# 查看系统服务日志
sudo journalctl -u spot-auto-manager -f
```

## 总结

这套API提供了完整的Spot VM管理功能：

- ✅ **状态查询**: 实时获取实例状态
- ✅ **Region管理**: 灵活切换目标Region
- ✅ **自动替换**: 智能检测和替换实例
- ✅ **手动控制**: 支持手动触发和测试
- ✅ **监控告警**: 提供状态查询和健康检查

适合在生产环境中部署，实现Spot VM的自动化管理。
