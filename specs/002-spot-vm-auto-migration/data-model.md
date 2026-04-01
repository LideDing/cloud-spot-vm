# Data Model: Spot CVM 自动迁移与自愈

**Feature**: 002-spot-vm-auto-migration
**Date**: 2026-04-01

## Entities

### SpotVM（现有，无变更）

代表当前运行的腾讯云 Spot 实例，通过 metadata API 采集自身状态。

| Field | Type | Description |
|-------|------|-------------|
| InstanceId | *string | 实例 ID |
| InstanceType | *string | 实例类型（如 S5.SMALL1） |
| Zone | *string | 所在可用区（如 ap-hongkong-2） |
| PublicIp | *string | 公网 IP |
| PrivateIp | *string | 内网 IP |

**State Transitions**: N/A（只读，从 metadata 采集）

---

### InstanceType（现有，无变更）

代表一种可售卖的 Spot 实例规格。

| Field | Type | Description |
|-------|------|-------------|
| InstanceType | string | 实例类型名称 |
| InstanceFamily | string | 实例族 |
| TypeName | string | 类型名称 |
| Cpu | int | CPU 核数 |
| Memory | int | 内存大小（MB） |
| Status | string | 售卖状态（SELL/SOLD_OUT） |
| Price | Price | 价格信息 |
| Zone | string | 所属可用区 |

---

### Price（现有，无变更）

实例价格信息。

| Field | Type | Description |
|-------|------|-------------|
| ChargeUnit | string | 计费单位 |
| UnitPrice | float64 | 原价 |
| UnitPriceDiscount | float64 | 折扣价（排序依据） |
| Discount | float64 | 折扣率 |

---

### Zone（现有，无变更）

腾讯云可用区。

| Field | Type | Description |
|-------|------|-------------|
| Zone | string | 可用区 ID（如 ap-hongkong-2） |
| ZoneId | string | 可用区数字 ID |
| ZoneName | string | 可用区名称 |
| ZoneState | string | 可用区状态（AVAILABLE/UNAVAILABLE） |

---

### MigrationTask（新增）

代表一次程序迁移操作。

| Field | Type | Description |
|-------|------|-------------|
| TargetInstanceId | string | 目标实例 ID |
| TargetIP | string | 目标实例公网 IP |
| TargetRegion | string | 目标 Region |
| TargetZone | string | 目标 Zone（系统自动选择） |
| Status | MigrationStatus | 迁移状态 |
| RetryCount | int | 已重试次数 |
| MaxRetries | int | 最大重试次数（默认 3） |
| StartTime | time.Time | 迁移开始时间 |
| Error | string | 最后一次错误信息 |

**State Transitions**:

```
PENDING → WAITING_SSH → TRANSFERRING → STARTING → VERIFYING → COMPLETED
    ↓         ↓              ↓            ↓           ↓
  FAILED    FAILED        FAILED       FAILED      FAILED
    ↓         ↓              ↓            ↓
  (retry)  (retry)       (retry)      (retry)
```

- **PENDING**: 迁移任务已创建，等待新实例创建完成
- **WAITING_SSH**: 等待新实例 SSH 端口就绪
- **TRANSFERRING**: 正在通过 SCP 传输文件
- **STARTING**: 正在通过 SSH 远程启动程序
- **VERIFYING**: 正在验证新实例上的程序是否正常运行（健康检查）
- **COMPLETED**: 迁移完成
- **FAILED**: 迁移失败（可重试）

---

### VMConfig（现有，无变更）

实例创建配置。

| Field | Type | Description |
|-------|------|-------------|
| ImageId | string | 镜像 ID |
| InstancePassword | string | 实例登录密码 |
| DiskType | string | 系统盘类型（如 CLOUD_BSSD） |
| DiskSize | int | 系统盘大小（GB） |
| Bandwidth | int | 公网带宽（Mbps） |

---

### Config（现有，需扩展）

应用配置结构，新增迁移相关配置项。

| Field | Type | Description | New? |
|-------|------|-------------|------|
| Environment | string | 运行环境 | |
| Port | string | 监听端口 | |
| JWTSecret | string | JWT 密钥 | |
| APIKey | string | API Key | |
| Region | string | 当前 Region | |
| Domain | string | 域名 | |
| CertificateId | string | SSL 证书 ID | |
| ImageId | string | 镜像 ID | |
| InstancePassword | string | 实例登录密码 | |
| DiskType | string | 系统盘类型 | |
| DiskSize | int | 系统盘大小 | |
| Bandwidth | int | 公网带宽 | |
| SSHPort | int | SSH 端口（默认 22） | ✅ |
| SSHTimeout | int | SSH 连接超时（秒，默认 10） | ✅ |
| SSHWaitTimeout | int | 等待 SSH 就绪超时（秒，默认 180） | ✅ |
| MigrationMaxRetries | int | 迁移最大重试次数（默认 3） | ✅ |
| RemoteBinaryPath | string | 远程二进制文件路径（默认 /opt/spot-manager/spot-manager） | ✅ |
| RemoteEnvPath | string | 远程 .env 文件路径（默认 /opt/spot-manager/.env） | ✅ |

## Relationships

```
TCC (1) ──── (1) SpotVMManager
TCC (1) ──── (1) SimpleAutoManager
TCC (1) ──── (1) SSLManager
TCC (1) ──── (1) DNSManager
TCC (1) ──── (1) Migrator [新增]

SimpleAutoManager (1) ──── (1) SpotVMManager
SimpleAutoManager (1) ──── (0..1) MigrationTask [新增：当前活跃的迁移任务]

SpotVMManager (1) ──── (1) SpotVM
SpotVMManager (1) ──── (*) Zone [查询]
SpotVMManager (1) ──── (*) InstanceType [查询]

Migrator (1) ──── (*) MigrationTask [执行]
```

## Validation Rules

- **InstancePassword**: 不能为空，长度 8-30 字符
- **Region**: 必须是腾讯云有效的 Region ID（通过 DescribeRegions API 验证）
- **Zone**: 由系统自动选择，不接受用户输入
- **SSHPort**: 范围 1-65535，默认 22
- **SSHTimeout**: 范围 1-60 秒
- **SSHWaitTimeout**: 范围 30-600 秒
- **MigrationMaxRetries**: 范围 1-10
- **CPU 过滤**: 实例 CPU >= 1
- **Memory 过滤**: 实例 Memory >= 1024 MB
