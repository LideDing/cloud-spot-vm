# Research: Cloud Spot VM

**Feature**: 001-codebase-logic-spec  
**Date**: 2026-03-31  
**Status**: Complete

## Research Tasks

### RT-001: 腾讯云 Spot 实例回收检测机制

**Decision**: 通过轮询腾讯云 metadata API 的 `/spot/termination-time` 端点检测回收信号

**Rationale**:
- 腾讯云 Spot 实例在被回收前会通过 metadata 服务发出通知
- 当 `http://metadata.tencentyun.com/latest/meta-data/spot/termination-time` 返回 HTTP 404 时，表示实例不会被回收（正常状态）
- 当返回非 404 状态码时，表示实例即将被回收，返回值为预计终止时间
- 当前实现每 10 秒轮询一次，这是合理的检测频率

**Alternatives considered**:
- 腾讯云 EventBridge 事件通知：更实时但需要额外配置，增加复杂度
- CloudAudit 日志监控：延迟较高，不适合实时检测
- 更短的轮询间隔（如 2 秒）：增加 metadata 服务负载，收益有限

---

### RT-002: 最便宜实例选择策略

**Decision**: 遍历目标 Region 所有可用区，查询所有 SPOTPAID 实例类型，按 `UnitPriceDiscount`（折扣后单价）升序排序，选择最便宜的

**Rationale**:
- `UnitPriceDiscount` 是实际支付价格，比 `UnitPrice`（原价）更准确
- 遍历所有可用区确保不会遗漏更便宜的选项
- 仅筛选 `Status == "SELL"` 的实例类型，确保可以实际购买
- 跳过查询失败的可用区（容错），不会因为单个可用区 API 错误而中断整个流程

**Alternatives considered**:
- 仅查询特定可用区：可能错过更便宜的选项
- 使用腾讯云 SpotPriceHistory API：可以预测价格趋势，但增加复杂度
- 设置最低配置要求（CPU/内存）：当前未实现，可能选到配置极低的实例

---

### RT-003: 腾讯云 SDK 使用模式

**Decision**: 使用 `tencentcloud-sdk-go` 的 CommonClient 模式，而非特定服务的 Client

**Rationale**:
- CommonClient 模式更灵活，可以调用任意 API 而无需引入每个服务的独立 SDK 包
- 减少依赖数量（只需 `tencentcloud/common` 一个包）
- 通过 `tchttp.NewCommonRequest` 构造请求，手动序列化/反序列化 JSON
- 缺点是失去了类型安全和 IDE 自动补全

**Alternatives considered**:
- 使用 `tencentcloud-sdk-go/tencentcloud/cvm`：类型安全但增加依赖
- 直接调用 REST API（不使用 SDK）：需要自行处理签名，工作量大
- 使用 Terraform Provider：适合基础设施即代码场景，不适合动态管理

---

### RT-004: API 认证方案

**Decision**: 使用静态 API Key 认证，支持三种传递方式

**Rationale**:
- `X-API-Key` 请求头：最常见的 API Key 传递方式
- `Authorization: Bearer <key>` 请求头：兼容 OAuth2 风格的客户端
- `api_key` 查询参数：方便在浏览器中快速测试
- 单用户场景下，静态 API Key 足够安全且简单
- 代码中已引入 JWT 库但未在主流程中使用，可能是为未来多用户场景预留

**Alternatives considered**:
- JWT Token 认证：已有代码框架但未启用，适合多用户场景
- OAuth2：过于复杂，不适合单用户运维工具
- mTLS：安全性最高但配置复杂

---

### RT-005: Nginx 部署方案

**Decision**: 通过 Docker 容器部署 Nginx，使用 `nginx:alpine` 镜像

**Rationale**:
- Docker 容器化部署简单、可重复、易于清理
- `nginx:alpine` 镜像体积小（约 40MB）
- 通过 volume 挂载配置文件、HTML 文件和 SSL 证书
- 支持自动安装 Docker（Ubuntu/CentOS/Fedora）
- `--restart unless-stopped` 确保容器在异常退出后自动重启

**Alternatives considered**:
- 直接安装 Nginx 到宿主机：配置管理更复杂，清理困难
- 使用 Caddy：自动 HTTPS 但不如 Nginx 普及
- 使用腾讯云 CLB（负载均衡器）：增加成本，违背成本最小化原则

---

### RT-006: 硬编码参数配置化方案

**Decision**: 以下参数需要从硬编码迁移到配置

**当前硬编码清单**:

| 参数 | 当前位置 | 硬编码值 | 建议配置方式 |
|------|---------|---------|------------|
| Region | `cmd/spot-manager/main.go:28` | `"sa-saopaulo"` | 环境变量 `REGION`（config.go 已支持但 main.go 未使用） |
| Domain | `cmd/spot-manager/main.go:29` | `"oitcep.com"` | 环境变量 `DOMAIN`（config.go 已支持但 main.go 未使用） |
| CertificateId | `cmd/spot-manager/main.go:30` | `"Qi1S1ItN"` | 环境变量 `CERTIFICATE_ID`（config.go 已支持但 main.go 未使用） |
| ImageId | `manager.go:96` | `"img-hdt9xxkt"` | 环境变量 `IMAGE_ID` |
| Password | `manager.go:107` | `"1qazZSE$"` | 环境变量 `INSTANCE_PASSWORD` |
| DiskType | `manager.go:93` | `"CLOUD_BSSD"` | 环境变量 `DISK_TYPE` |
| DiskSize | `manager.go:94` | `20` | 环境变量 `DISK_SIZE` |
| Bandwidth | `manager.go:102` | `10` | 环境变量 `BANDWIDTH` |
| Tags | `manager.go:109-115` | 固定标签 | 环境变量或配置文件 |

**Rationale**: config.go 已经定义了 Region、Domain、CertificateId 的配置字段，但 main.go 中未使用这些配置值，而是直接硬编码。这是一个明显的配置化不完整问题。

## Resolved Clarifications

所有技术决策已基于代码分析确认，无需额外澄清。
