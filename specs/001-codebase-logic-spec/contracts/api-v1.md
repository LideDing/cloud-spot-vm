# API Contract: Cloud Spot VM REST API v1

**Feature**: 001-codebase-logic-spec  
**Date**: 2026-03-31  
**Base URL**: `http://{host}:{port}/api/v1`  
**Authentication**: API Key（详见下方认证章节）

## Authentication

所有 `/api/v1/spot-vm/*` 端点需要 API Key 认证。支持三种传递方式：

| 方式 | 格式 | 示例 |
|------|------|------|
| 请求头 `X-API-Key` | `X-API-Key: {api_key}` | `X-API-Key: my-secret-key` |
| 请求头 `Authorization` | `Authorization: Bearer {api_key}` | `Authorization: Bearer my-secret-key` |
| 查询参数 `api_key` | `?api_key={api_key}` | `?api_key=my-secret-key` |

**认证失败响应**:
```json
{
  "error": "Invalid or missing API key"
}
```
HTTP Status: `401 Unauthorized`

---

## Endpoints

### 1. Health Check

**无需认证**

```
GET /api/v1/health
```

**Response** `200 OK`:
```json
{
  "status": "healthy",
  "message": "Service is running",
  "timestamp": "2026-03-31T17:00:00Z"
}
```

---

### 2. Validate API Key

**无需认证**

```
POST /api/v1/auth/validate
```

**Request Headers**:
```
X-API-Key: {api_key}
```

**Response** `200 OK`:
```json
{
  "status": "valid",
  "message": "API key is valid"
}
```

**Response** `401 Unauthorized`:
```json
{
  "error": "Invalid API key"
}
```

---

### 3. Get Cheapest Instances

查询当前 Region 所有可用区中按价格排序的前 10 个最便宜 Spot 实例。

```
GET /api/v1/spot-vm/cheapest
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Retrieved cheapest instances successfully",
  "instances": [
    {
      "InstanceType": "SA5.MEDIUM2",
      "InstanceFamily": "SA5",
      "TypeName": "标准型SA5",
      "Cpu": 1,
      "Memory": 2,
      "CpuType": "AMD EPYC™ Genoa",
      "Status": "SELL",
      "Price": {
        "ChargeUnit": "HOUR",
        "UnitPrice": 0.04,
        "UnitPriceDiscount": 0.012,
        "Discount": 30.0
      },
      "Zone": "sa-saopaulo-1"
    }
  ],
  "total": 150
}
```

**Response** `404 Not Found`:
```json
{
  "error": "No available spot instances found"
}
```

---

### 4. Create Cheapest Spot VM

在最便宜的可用区创建一个 Spot 实例。

```
POST /api/v1/spot-vm/create-cheapest
```

**Request Body**:
```json
{
  "dry_run": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| dry_run | boolean | Yes | `true` 仅预检不创建，`false` 实际创建 |

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Spot VM created successfully",
  "instance_info": {
    "InstanceType": "SA5.MEDIUM2",
    "Zone": "sa-saopaulo-1",
    "Price": {
      "UnitPriceDiscount": 0.012
    }
  },
  "instance_ids": ["ins-abc12345"],
  "dry_run": false
}
```

---

### 5. Get Available Zones

获取当前 Region 的可用区列表。

```
GET /api/v1/spot-vm/zones
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Retrieved available zones successfully",
  "zones": [
    {
      "Zone": "sa-saopaulo-1",
      "ZoneId": "800001",
      "ZoneName": "圣保罗一区",
      "ZoneState": "AVAILABLE"
    }
  ],
  "total": 2
}
```

---

### 6. Get Instance Types by Zone

获取指定可用区的所有可售卖 Spot 实例类型。

```
GET /api/v1/spot-vm/instance-types?zone={zone_id}
```

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| zone | string | Yes | 可用区 ID（如 `ap-hongkong-1`） |

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Retrieved instance types successfully",
  "instances": [...],
  "zone": "ap-hongkong-1",
  "total": 50
}
```

**Response** `400 Bad Request`:
```json
{
  "error": "zone parameter is required"
}
```

---

### 7. Delete Instance

终止并释放指定实例。

```
DELETE /api/v1/spot-vm/instances/{instance_id}
```

**Path Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| instance_id | string | Yes | 实例 ID（如 `ins-abc12345`） |

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Instance deleted successfully",
  "instance_id": "ins-abc12345"
}
```

---

### 8. Get Current Instance Status

获取当前运行的 Spot 实例状态（通过 metadata API 采集）。

```
GET /api/v1/spot-vm/current/status
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Retrieved current instance status successfully",
  "data": {
    "instance_id": "ins-abc12345",
    "instance_type": "SA5.MEDIUM2",
    "zone": "sa-saopaulo-1",
    "region": "sa-saopaulo",
    "status": "RUNNING",
    "is_terminated": false,
    "private_ip": "10.0.0.5",
    "public_ip": "203.0.113.10",
    "created_time": "2026-03-31T17:00:00Z",
    "expired_time": "N/A"
  }
}
```

---

### 9. Get Current Region

查看当前 Spot 实例所在的 Region。

```
GET /api/v1/spot-vm/current/region
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Retrieved current region successfully",
  "data": {
    "current_region": "sa-saopaulo"
  }
}
```

---

### 10. Set Target Region

修改目标 Region（替换实例将在此 Region 中创建）。

```
PUT /api/v1/spot-vm/target-region
```

**Request Body**:
```json
{
  "region": "ap-singapore"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| region | string | Yes | 目标 Region ID |

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Target region updated successfully",
  "data": {
    "current_region": "sa-saopaulo",
    "target_region": "ap-singapore"
  }
}
```

---

### 11. Trigger Termination

手动触发实例回收流程（与自动检测到回收信号效果相同）。

```
POST /api/v1/spot-vm/trigger-termination
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Manual termination triggered successfully"
}
```

**Response** `400 Bad Request`:
```json
{
  "error": "Auto manager is not running. Please start it first."
}
```

---

### 12. Get Auto Manager Status

获取自动管理器的运行状态。

```
GET /api/v1/spot-vm/auto-manager/status
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Retrieved auto manager status successfully",
  "data": {
    "is_running": true,
    "region": "sa-saopaulo",
    "target_region": "ap-singapore",
    "last_check": "2026-03-31T17:00:00Z"
  }
}
```

---

### 13. Start Auto Manager

启动自动管理器。

```
POST /api/v1/spot-vm/auto-manager/start
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Auto manager started successfully"
}
```

**Response** `400 Bad Request`:
```json
{
  "error": "Auto manager is already running"
}
```

---

### 14. Stop Auto Manager

停止自动管理器。

```
POST /api/v1/spot-vm/auto-manager/stop
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Auto manager stopped successfully"
}
```

**Response** `400 Bad Request`:
```json
{
  "error": "Auto manager is not running"
}
```

---

### 15. Simulate Termination

模拟实例终止（用于测试替换流程）。

```
POST /api/v1/spot-vm/auto-manager/simulate-termination
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "Termination simulation triggered successfully"
}
```

**Response** `400 Bad Request`:
```json
{
  "error": "Auto manager is not running. Please start it first."
}
```

---

## Common Error Responses

| HTTP Status | Scenario |
|-------------|----------|
| `400 Bad Request` | 请求参数缺失或格式错误 |
| `401 Unauthorized` | API Key 缺失或无效 |
| `404 Not Found` | 无可用 Spot 实例 |
| `500 Internal Server Error` | 腾讯云 API 调用失败 |

## Middleware Chain

所有请求经过以下中间件处理（按顺序）：

1. **LoggerMiddleware** — 记录请求日志
2. **CORSMiddleware** — 跨域支持（允许所有来源）
3. **RateLimitMiddleware** — 速率限制（当前为空实现）
4. **AuditMiddleware** — 审计日志（当前为空实现）
5. **APIKeyAuthMiddleware** — API Key 认证（仅受保护端点）
