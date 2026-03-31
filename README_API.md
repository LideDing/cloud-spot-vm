# API Key 认证系统

本项目已经将JWT token认证替换为简单的API key认证系统，使用硬编码的API key进行API鉴权。

## 配置

### 环境变量

在 `.env` 文件中设置以下环境变量：

```bash
# API Key（生产环境中请使用强密钥）
API_KEY=your-super-secret-api-key-change-in-production

# 服务器端口
PORT=8080

# 其他配置...
```

### 默认配置

如果没有设置环境变量，系统将使用以下默认值：
- `API_KEY`: `your-super-secret-api-key-change-in-production`
- `PORT`: `8080`

## API 端点

### 无需认证的端点

#### 1. 健康检查
```http
GET /api/v1/health
```

**响应示例：**
```json
{
  "status": "healthy",
  "message": "API server is running"
}
```

#### 2. API Key 验证
```http
POST /api/v1/auth/validate
Content-Type: application/json

{
  "api_key": "your-api-key"
}
```

**响应示例：**
```json
{
  "status": "success",
  "message": "API key is valid"
}
```

### 需要认证的端点

以下端点需要提供有效的API key：

#### 1. 受保护的端点
```http
GET /api/v1/protected
```

#### 2. 获取实例信息
```http
GET /api/v1/instances
```

#### 3. 创建实例
```http
POST /api/v1/instances
```

#### 4. 删除实例
```http
DELETE /api/v1/instances/:id
```

#### 5. 系统信息
```http
GET /api/v1/system-info
```

#### 6. 统计信息
```http
GET /api/v1/stats
```

## 认证方式

API key可以通过以下三种方式提供：

### 1. X-API-Key 请求头（推荐）
```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/protected
```

### 2. Authorization 请求头
```bash
curl -H "Authorization: Bearer your-api-key" http://localhost:8080/api/v1/protected
```

### 3. 查询参数
```bash
curl "http://localhost:8080/api/v1/protected?api_key=your-api-key"
```

## 使用示例

### 启动服务器
```bash
go run cmd/main.go
```

### 测试API
```bash
# 使用默认API key测试
./test_api.sh

# 使用自定义API key测试
./test_api.sh "your-custom-api-key"

# 简单测试
./test_simple.sh
```

### 手动测试

#### 健康检查
```bash
curl http://localhost:8080/api/v1/health
```

#### 验证API key
```bash
curl -X POST http://localhost:8080/api/v1/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"api_key": "your-api-key"}'
```

#### 访问受保护的端点
```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/protected
```

## 错误响应

### 401 Unauthorized
当API key无效或缺失时：
```json
{
  "error": "Invalid or missing API key"
}
```

### 400 Bad Request
当请求格式错误时：
```json
{
  "error": "具体错误信息"
}
```

## 安全建议

1. **生产环境**：使用强随机生成的API key
2. **HTTPS**：在生产环境中使用HTTPS传输
3. **轮换**：定期轮换API key
4. **限制**：限制API key的使用范围和权限
5. **监控**：监控API key的使用情况

## 代码结构

```
internal/
├── config/
│   └── config.go          # 配置管理，包含API key配置
├── handlers/
│   └── simple_auth.go     # API key认证处理器
├── middleware/
│   ├── middleware.go      # 通用中间件
│   └── simple_auth.go     # API key认证中间件
└── routes/
    ├── routes.go          # 主路由配置
    └── simple_routes.go   # 简化路由配置
```

## 主要变更

1. **移除JWT依赖**：不再使用JWT token进行认证
2. **简化认证**：使用硬编码的API key进行简单认证
3. **多种认证方式**：支持请求头、Authorization头和查询参数三种方式
4. **向后兼容**：保持原有的API结构，只是认证方式改变
5. **修复路由错误**：修复了 `simple_routes.go` 中的JWT相关错误
6. **新增Spot VM功能**：添加了创建最便宜Spot VM的API功能
7. **自动管理功能**：实现了Spot实例回收检测和自动替换功能

## Spot VM API 功能

### 新增端点

- `GET /api/v1/spot-vm/zones` - 获取可用区列表
- `GET /api/v1/spot-vm/cheapest` - 获取最便宜的实例列表
- `POST /api/v1/spot-vm/create-cheapest` - 创建最便宜的Spot VM
- `GET /api/v1/spot-vm/instance-types?zone=xxx` - 获取指定可用区的实例类型
- `DELETE /api/v1/spot-vm/instances/:id` - 删除实例

### 自动管理端点

- `GET /api/v1/spot-vm/auto-manager/status` - 获取自动管理器状态
- `POST /api/v1/spot-vm/auto-manager/start` - 启动自动管理器
- `POST /api/v1/spot-vm/auto-manager/stop` - 停止自动管理器
- `POST /api/v1/spot-vm/auto-manager/simulate-termination` - 模拟实例终止（测试用）
- `GET /api/v1/spot-vm/auto-manager/replacement-history` - 获取替换历史

### 使用示例

```bash
# 获取最便宜的实例列表
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/spot-vm/cheapest

# 创建最便宜的Spot VM（预检模式）
curl -X POST -H "X-API-Key: your-api-key" -H "Content-Type: application/json" \
  -d '{"dry_run": true}' http://localhost:8080/api/v1/spot-vm/create-cheapest

# 创建最便宜的Spot VM（实际创建）
curl -X POST -H "X-API-Key: your-api-key" -H "Content-Type: application/json" \
  -d '{"dry_run": false}' http://localhost:8080/api/v1/spot-vm/create-cheapest

# 自动管理功能
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/spot-vm/auto-manager/status
curl -X POST -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/spot-vm/auto-manager/start
curl -X POST -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/spot-vm/auto-manager/simulate-termination
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/spot-vm/auto-manager/replacement-history
```

### 配置说明

⚠️ **重要**：在创建Spot VM之前，需要配置正确的腾讯云资源：

1. **VPC和子网**：在 `internal/tcc/spot_vm/manager.go` 中修改以下配置：
   ```go
   "VirtualPrivateCloud": map[string]any{
       "VpcId":    "your-vpc-id",      // 替换为你的VPC ID
       "SubnetId": "your-subnet-id",   // 替换为你的子网ID
   },
   ```

2. **镜像ID**：修改镜像ID为你的可用镜像：
   ```go
   "ImageId": "your-image-id",  // 替换为你的镜像ID
   ```

3. **其他配置**：根据需要修改密码、标签等配置。

## 测试结果

✅ **编译成功**：代码可以正常编译
✅ **健康检查**：`/api/v1/health` 端点正常工作
✅ **API key验证**：`/api/v1/auth/validate` 端点正常工作
✅ **认证保护**：受保护的端点正确要求API key
✅ **错误处理**：无认证访问正确返回401错误
✅ **Spot VM API**：所有Spot VM相关端点正常工作
✅ **可用区查询**：成功获取可用区列表
✅ **实例查询**：成功获取实例类型信息
✅ **自动管理功能**：自动管理器正常工作
✅ **回收检测**：每10秒检查一次实例回收状态
✅ **自动替换**：检测到回收时自动创建新实例
✅ **模拟测试**：支持模拟实例终止进行测试
⚠️ **实例创建**：需要配置正确的VPC和子网ID
