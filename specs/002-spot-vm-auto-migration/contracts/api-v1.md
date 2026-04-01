# API Contract v1: Spot CVM 自动迁移与自愈

**Feature**: 002-spot-vm-auto-migration
**Date**: 2026-04-01
**Base URL**: `/api/v1`

## Authentication

所有受保护端点需要以下任一方式提供 API Key：
- Header: `X-API-Key: <key>`
- Header: `Authorization: Bearer <key>`
- Query: `?api_key=<key>`

未认证请求返回 `401 Unauthorized`。

---

## Endpoints

### 1. Health Check（无需认证）

```
GET /api/v1/health
```

**Response** `200 OK`:
```json
{
  "status": "ok",
  "timestamp": "2026-04-01T12:00:00Z"
}
```

---

### 2. 获取当前实例状态

```
GET /api/v1/spot-vm/current/status
```

**Response** `200 OK`:
```json
{
  "instance_id": "ins-xxxxxxxx",
  "instance_type": "S5.SMALL1",
  "zone": "ap-hongkong-2",
  "region": "ap-hongkong",
  "status": "RUNNING",
  "is_terminated": false,
  "private_ip": "10.0.0.1",
  "public_ip": "1.2.3.4",
  "created_time": "2026-04-01T10:00:00Z",
  "expired_time": "N/A"
}
```

---

### 3. 获取当前 Region

```
GET /api/v1/spot-vm/current/region
```

**Response** `200 OK`:
```json
{
  "current_region": "ap-hongkong",
  "target_region": "ap-singapore"
}
```

---

### 4. 修改目标 Region

```
PUT /api/v1/spot-vm/target-region
```

**Request Body**:
```json
{
  "region": "ap-singapore"
}
```

**Response** `200 OK`:
```json
{
  "message": "目标Region已更新",
  "old_region": "ap-hongkong",
  "new_region": "ap-singapore"
}
```

**Response** `400 Bad Request`（无效 Region）:
```json
{
  "error": "无效的Region: ap-invalid",
  "available_regions": ["ap-hongkong", "ap-singapore", "ap-seoul", ...]
}
```

> **注意**: 此 API 仅接受 Region 级别的参数，Zone 由系统自动选择。

---

### 5. 获取可用 Region 列表

```
GET /api/v1/spot-vm/regions
```

**Response** `200 OK`:
```json
{
  "regions": [
    {"region": "ap-hongkong", "region_name": "中国香港", "region_state": "AVAILABLE"},
    {"region": "ap-singapore", "region_name": "新加坡", "region_state": "AVAILABLE"}
  ]
}
```

---

### 6. 获取最便宜实例列表

```
GET /api/v1/spot-vm/cheapest
```

**Response** `200 OK`:
```json
{
  "instances": [
    {
      "instance_type": "S5.SMALL1",
      "cpu": 1,
      "memory": 1024,
      "zone": "ap-hongkong-2",
      "price": 0.0120,
      "status": "SELL"
    }
  ],
  "region": "ap-hongkong",
  "total": 10
}
```

---

### 7. 创建最便宜的 Spot 实例

```
POST /api/v1/spot-vm/create-cheapest
```

**Request Body**:
```json
{
  "dry_run": false
}
```

**Response** `200 OK`:
```json
{
  "instance_ids": ["ins-xxxxxxxx"],
  "zone": "ap-hongkong-2",
  "instance_type": "S5.SMALL1",
  "message": "Spot实例创建成功（Zone 由系统自动选择）"
}
```

---

### 8. 删除实例

```
DELETE /api/v1/spot-vm/instances/:id
```

**Response** `200 OK`:
```json
{
  "message": "实例已删除",
  "instance_id": "ins-xxxxxxxx"
}
```

---

### 9. 手动触发回收

```
POST /api/v1/spot-vm/trigger-termination
```

**Response** `200 OK`:
```json
{
  "message": "已触发实例回收流程（将自动创建替代实例、迁移程序、更新 DNS）"
}
```

---

### 10. 获取自动管理器状态

```
GET /api/v1/spot-vm/auto-manager/status
```

**Response** `200 OK`:
```json
{
  "is_running": true,
  "region": "ap-hongkong",
  "target_region": "ap-singapore",
  "last_check": "2026-04-01T12:00:00Z",
  "migration_status": null
}
```

当有活跃迁移任务时：
```json
{
  "is_running": true,
  "region": "ap-hongkong",
  "target_region": "ap-singapore",
  "last_check": "2026-04-01T12:00:00Z",
  "migration_status": {
    "target_instance_id": "ins-yyyyyyyy",
    "target_ip": "5.6.7.8",
    "target_region": "ap-singapore",
    "target_zone": "ap-singapore-3",
    "status": "TRANSFERRING",
    "retry_count": 0,
    "start_time": "2026-04-01T12:00:30Z"
  }
}
```

---

### 11. 启动自动管理器

```
POST /api/v1/spot-vm/auto-manager/start
```

**Response** `200 OK`:
```json
{
  "message": "自动管理器已启动"
}
```

---

### 12. 停止自动管理器

```
POST /api/v1/spot-vm/auto-manager/stop
```

**Response** `200 OK`:
```json
{
  "message": "自动管理器已停止"
}
```

---

### 13. 模拟实例终止

```
POST /api/v1/spot-vm/auto-manager/simulate-termination
```

**Response** `200 OK`:
```json
{
  "message": "已模拟实例终止，将触发完整的替换 + 迁移流程"
}
```

---

### 14. 获取可用区列表

```
GET /api/v1/spot-vm/zones
```

**Response** `200 OK`:
```json
{
  "zones": [
    {"zone": "ap-hongkong-1", "zone_name": "香港一区", "zone_state": "AVAILABLE"},
    {"zone": "ap-hongkong-2", "zone_name": "香港二区", "zone_state": "AVAILABLE"}
  ]
}
```

---

### 15. 获取指定可用区的实例类型

```
GET /api/v1/spot-vm/instance-types?zone=ap-hongkong-2
```

**Response** `200 OK`:
```json
{
  "instance_types": [
    {
      "instance_type": "S5.SMALL1",
      "cpu": 1,
      "memory": 1024,
      "price": 0.0120,
      "status": "SELL"
    }
  ],
  "zone": "ap-hongkong-2"
}
```

---

## Error Responses

所有端点的通用错误格式：

**401 Unauthorized**:
```json
{
  "error": "未授权：缺少或无效的 API Key"
}
```

**400 Bad Request**:
```json
{
  "error": "请求参数错误：<具体描述>"
}
```

**500 Internal Server Error**:
```json
{
  "error": "内部服务器错误：<具体描述>"
}
```

---

## Migration Flow（内部流程，非 API）

当检测到回收信号或手动触发时，系统自动执行以下流程：

```
1. 检测回收信号
2. 遍历目标 Region 所有 Zone，选择最便宜的实例类型
3. 创建替代实例（指数退避重试）
4. 轮询 DescribeInstances 获取公网 IP
5. 等待 SSH 端口就绪
6. SCP 传输文件（二进制 + .env + SSL 证书）
7. SSH 远程启动程序（nohup）
8. 健康检查验证新实例程序运行正常
9. 更新 DNS A 记录
10. 部署 Nginx（Docker 容器）
```
