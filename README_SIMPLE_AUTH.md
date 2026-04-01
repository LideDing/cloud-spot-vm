# 简化 API 鉴权系统

这是一个基于 Gin 框架的简化 API 鉴权系统，**不依赖数据库**，只通过 JWT token 进行鉴权。

## 特性

- ✅ 无数据库依赖
- ✅ 基于 JWT token 鉴权
- ✅ 简单的用户名/密码验证
- ✅ 角色权限控制
- ✅ 完整的 API 示例

## 快速开始

### 1. 启动服务

```bash
# 启动 Web API 服务
go run cmd/web.go
```

### 2. 默认登录凭据

- **用户名**: `admin`
- **密码**: `admin123`

### 3. API 使用示例

#### 登录获取 Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

响应示例：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "admin",
  "expires_at": "2024-01-01T12:00:00Z"
}
```

#### 使用 Token 访问受保护的 API

```bash
# 获取用户资料
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 更新用户资料
curl -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "管理员"
  }'

# 获取用户统计
curl -X GET http://localhost:8080/api/v1/users/stats \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 管理员专用 API
curl -X GET http://localhost:8080/api/v1/admin/stats \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

curl -X GET http://localhost:8080/api/v1/admin/system-info \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 登出

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## API 端点

### 无需认证的端点

- `GET /api/v1/health` - 健康检查
- `POST /api/v1/auth/login` - 用户登录

### 需要认证的端点

- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/users/profile` - 获取用户资料
- `PUT /api/v1/users/profile` - 更新用户资料
- `GET /api/v1/users/stats` - 获取用户统计

### 管理员专用端点

- `GET /api/v1/admin/stats` - 管理员统计
- `GET /api/v1/admin/system-info` - 系统信息

## 配置

通过环境变量配置：

```bash
# 服务端口
export PORT=8080

# JWT 密钥（生产环境请使用强密钥）
export JWT_SECRET=your-super-secret-jwt-key-change-in-production

# 环境
export ENVIRONMENT=development
```

## 文件结构

```
internal/
├── handlers/
│   ├── simple_auth.go    # 简化认证处理器
│   └── simple_user.go    # 简化用户处理器
├── middleware/
│   └── simple_auth.go    # 简化认证中间件
├── routes/
│   └── simple_routes.go  # 简化路由配置
└── config/
    └── config.go         # 配置管理

cmd/
└── web.go               # Web 服务入口点
```

## 自定义用户凭据

要修改默认的用户名和密码，编辑 `internal/handlers/simple_auth.go` 文件中的 `SimpleLogin` 函数：

```go
// 硬编码的用户名和密码验证
if req.Username != "your_username" || req.Password != "your_password" {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
    return
}
```

## 安全注意事项

1. **生产环境**：请修改默认的用户名和密码
2. **JWT 密钥**：使用强密钥，不要使用默认值
3. **HTTPS**：生产环境请使用 HTTPS
4. **Token 过期**：当前设置为 24 小时，可根据需要调整

## 扩展功能

如需添加更多功能，可以：

1. 在 `SimpleAuthHandler` 中添加新的认证方法
2. 在 `SimpleUserHandler` 中添加新的用户相关功能
3. 在路由中添加新的端点
4. 扩展角色权限系统

## 故障排除

### 常见错误

1. **401 Unauthorized**: Token 无效或已过期
2. **403 Forbidden**: 权限不足
3. **400 Bad Request**: 请求格式错误

### 调试

启用调试模式：

```bash
export ENVIRONMENT=development
```

查看详细日志输出。
