# Feature Specification: Cloud Spot VM — 腾讯云竞价实例自动管理平台

**Feature Branch**: `001-codebase-logic-spec`  
**Created**: 2026-03-31  
**Status**: Draft  
**Input**: User description: "检查已有代码，重新生成工程逻辑文档，搞清楚这个工程的逻辑"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 自动管理 Spot 实例生命周期 (Priority: P1)

作为一名运维人员，我希望系统能够自动监控当前运行的腾讯云 Spot（竞价）实例，当检测到实例即将被回收时，自动在目标 Region 中找到最便宜的可用 Spot 实例并创建替代实例，从而保证服务的连续性和成本最优。

**Why this priority**: 这是整个系统的核心价值——自动化 Spot 实例的生命周期管理。没有这个功能，系统就没有存在的意义。Spot 实例随时可能被云平台回收，自动替换是保障服务可用性的关键。

**Independent Test**: 可以通过模拟终止接口 (`/auto-manager/simulate-termination`) 触发回收信号，验证系统是否自动查询最便宜实例并创建替代实例。

**Acceptance Scenarios**:

1. **Given** 自动管理器已启动且当前有一个 Spot 实例正在运行, **When** 腾讯云 metadata 接口返回实例即将被回收的信号（`/spot/termination-time` 返回非 404）, **Then** 系统自动在目标 Region 的所有可用区中查找最便宜的 Spot 实例类型并创建新实例
2. **Given** 自动管理器已启动, **When** 用户通过 API 手动触发回收 (`POST /trigger-termination`), **Then** 系统执行与自动检测相同的替换流程
3. **Given** 自动管理器已启动, **When** 目标 Region 中没有可用的 Spot 实例, **Then** 系统记录错误日志，不会崩溃，继续监控

---

### User Story 2 - 通过 REST API 查询和管理 Spot 实例 (Priority: P2)

作为一名运维人员，我希望通过 REST API 查看当前 Spot 实例的状态、查询最便宜的实例列表、手动创建或删除实例，以便在需要时进行手动干预。

**Why this priority**: API 管理接口是运维人员与系统交互的主要方式，提供了对自动化流程的可见性和控制能力。

**Independent Test**: 可以通过 curl 或 HTTP 客户端调用各 API 端点，验证返回数据的正确性和完整性。

**Acceptance Scenarios**:

1. **Given** 系统已启动且配置了有效的腾讯云凭证, **When** 用户携带有效 API Key 请求 `GET /spot-vm/cheapest`, **Then** 系统返回当前 Region 所有可用区中按价格排序的前 10 个最便宜 Spot 实例列表
2. **Given** 系统已启动, **When** 用户携带有效 API Key 请求 `POST /spot-vm/create-cheapest` 且 `dry_run=false`, **Then** 系统在最便宜的可用区创建一个 Spot 实例并返回实例 ID
3. **Given** 系统已启动且存在一个实例, **When** 用户携带有效 API Key 请求 `DELETE /spot-vm/instances/:id`, **Then** 系统终止并释放该实例
4. **Given** 用户未提供 API Key 或提供了无效的 API Key, **When** 请求任何受保护的端点, **Then** 系统返回 401 未授权错误

---

### User Story 3 - 跨 Region 切换与管理 (Priority: P3)

作为一名运维人员，我希望能够动态修改目标 Region，使得当实例被回收后，新实例可以在不同的 Region 中创建，以便利用不同 Region 的价格差异或应对某个 Region 容量不足的情况。

**Why this priority**: 跨 Region 能力扩展了成本优化的范围，但不是系统运行的基本前提。

**Independent Test**: 可以通过 `PUT /spot-vm/target-region` 修改目标 Region，然后触发模拟终止，验证新实例是否在新 Region 中创建。

**Acceptance Scenarios**:

1. **Given** 自动管理器正在运行且当前 Region 为 `ap-hongkong`, **When** 用户通过 API 将目标 Region 修改为 `ap-singapore`, **Then** 系统确认修改成功，后续替换实例将在 `ap-singapore` 中创建
2. **Given** 目标 Region 已修改, **When** 触发实例替换, **Then** 系统使用新 Region 的 SpotVMManager 查询可用区和实例类型，并在新 Region 中创建实例

---

### User Story 4 - 网络基础设施自动配置 (Priority: P4)

作为一名运维人员，我希望系统在创建新 Spot 实例后能够自动配置 Nginx 反向代理（含 SSL 证书）和 DNS 记录，使得服务能够通过域名以 HTTPS 方式访问。

**Why this priority**: 网络配置是服务可用性的最后一环，但属于增值功能，核心的实例管理可以独立运行。

**Independent Test**: 可以在新实例创建后，检查 Nginx 容器是否启动、SSL 证书是否部署、DNS 记录是否更新。

**Acceptance Scenarios**:

1. **Given** 新 Spot 实例已创建且获得了公网 IP, **When** 系统执行网络配置流程, **Then** Nginx 容器以 Docker 方式部署，80 端口重定向到 443，443 端口配置 SSL 证书
2. **Given** 新实例已获得公网 IP 且配置了域名, **When** 系统执行 DNS 配置, **Then** 域名的 A 记录更新为新实例的公网 IP

---

### Edge Cases

- 当腾讯云 API 返回速率限制错误时，系统如何处理？（当前：跳过失败的可用区继续查询）
- 当所有可用区都没有可售卖的 Spot 实例时，系统如何处理？（当前：记录错误日志，不创建实例）
- 当 metadata 服务不可达时（例如不在腾讯云实例上运行），终止检测如何表现？（当前：返回 false，不触发替换）
- 当自动管理器的 terminationCh 通道已满时，新的终止信号如何处理？（当前：通道缓冲区为 1，可能阻塞）
- 当创建替换实例失败时，系统是否会重试？（当前：不重试，仅记录错误日志）
- 当同时收到多个终止信号时，是否会创建多个替换实例？（当前：可能会，因为没有去重机制）

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 每 10 秒通过腾讯云 metadata 接口检查当前 Spot 实例是否即将被回收
- **FR-002**: 系统 MUST 在检测到回收信号后，自动查询目标 Region 所有可用区的 Spot 实例价格，选择最便宜的实例类型创建替代实例
- **FR-003**: 系统 MUST 提供 REST API 供运维人员查询实例状态、管理实例生命周期和控制自动管理器
- **FR-004**: 系统 MUST 通过 API Key 认证保护所有管理端点，支持通过 `X-API-Key` 请求头、`Authorization: Bearer` 请求头或 `api_key` 查询参数传递凭证
- **FR-005**: 系统 MUST 支持动态修改目标 Region，使替换实例可以在不同 Region 中创建
- **FR-006**: 系统 MUST 支持 DryRun 模式，允许用户预检创建请求而不实际创建实例
- **FR-007**: 系统 MUST 提供模拟终止接口，允许在不等待真实回收事件的情况下测试替换流程
- **FR-008**: 系统 MUST 支持通过 Docker 部署 Nginx 反向代理，配置 SSL 证书和 HTTP→HTTPS 重定向
- **FR-009**: 系统 MUST 支持通过腾讯云 DNSPod API 自动更新域名 A 记录
- **FR-010**: 系统 MUST 支持从腾讯云 SSL 证书服务获取 SSL 证书
- **FR-011**: 系统 MUST 提供健康检查端点（无需认证）
- **FR-012**: 系统 MUST 记录所有关键操作的日志（实例创建、删除、回收检测、Region 切换等）

### Key Entities

- **Spot 实例 (SpotVM)**: 代表一个腾讯云竞价实例，包含实例 ID、实例类型、可用区、公网 IP、私网 IP、运行状态、是否即将被回收等属性。通过 metadata 服务获取自身信息。
- **实例类型 (InstanceType)**: 代表一种可售卖的 Spot 实例规格，包含实例类型名称、CPU/内存配置、价格（含折扣价）、所属可用区、售卖状态等属性。
- **可用区 (Zone)**: 代表腾讯云的一个可用区，包含可用区 ID、名称、状态。一个 Region 下有多个可用区。
- **自动管理器 (SimpleAutoManager)**: 代表自动化管理的控制器，维护当前 Region、目标 Region、运行状态，负责协调终止检测和实例替换。
- **TCC 客户端**: 腾讯云服务的统一入口，聚合了 SpotVMManager（实例管理）、SSLManager（证书管理）、DNSManager（域名管理）。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 从检测到 Spot 实例回收信号到新替换实例创建请求发出，整个流程在 30 秒内完成
- **SC-002**: 系统能够在 10 个以上腾讯云 Region 中查询和创建 Spot 实例
- **SC-003**: 自动管理器能够 7×24 小时持续运行，无内存泄漏或 goroutine 泄漏
- **SC-004**: 所有受保护的 API 端点在未提供有效凭证时 100% 返回 401 错误
- **SC-005**: 系统始终选择目标 Region 中价格最低的 Spot 实例类型，确保成本最优
- **SC-006**: 运维人员能够在 1 分钟内通过 API 完成 Region 切换操作

## Assumptions

- 系统运行在腾讯云 CVM 实例上，能够访问 metadata 服务 (`http://metadata.tencentyun.com/latest/meta-data/`)
- 用户已在腾讯云控制台创建了 API 密钥（SecretId/SecretKey），并通过环境变量 `TENCENTCLOUD_SECRET_ID` 和 `TENCENTCLOUD_SECRET_KEY` 提供
- 系统使用的镜像 ID (`img-hdt9xxkt`) 在目标 Region 中可用
- 系统创建的实例使用 CLOUD_BSSD 类型的 20GB 系统盘，按流量计费的 10Mbps 公网带宽
- 实例登录密码硬编码在代码中（当前为 `1qazZSE$`），生产环境需要改为配置化
- DNS 管理仅支持单一域名和 `www` 子域名
- Docker 已安装或系统支持自动安装 Docker（Ubuntu/CentOS/Fedora）
- 当前不需要持久化存储（无数据库依赖），所有状态保存在内存中
- 单实例部署模式，不考虑多实例高可用场景
