# Feature Specification: Spot CVM 自动迁移与自愈

**Feature Branch**: `002-spot-vm-auto-migration`  
**Created**: 2026-04-01  
**Status**: Draft  
**Input**: User description: "程序跑在腾讯云 Spot CVM 中，定时检测状态，发现将被回收时自动购买新 Spot CVM 并迁移自身；通过 HTTP API 控制下一个 CVM 的区域（Region），Zone 由程序自动选择；新 CVM 自动修改 DNS 解析并启动 Docker Nginx；在现有框架基础上修改"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 自动检测回收并创建替代实例 (Priority: P1)

作为一名运维人员，我希望程序能够在腾讯云 Spot CVM 上持续运行，每隔固定间隔检测当前实例是否即将被回收。一旦检测到回收信号，系统立即在目标 Region 中自动选择有可用 Spot 容量的 Zone，并购买该 Zone 中价格最低的新 Spot CVM 实例，确保服务不中断。用户只需指定 Region，无需关心具体的 Zone 选择。

**Why this priority**: 这是整个系统存在的核心价值。Spot 实例随时可能被回收，自动检测和替换是保障服务连续性的基础。没有这个能力，后续的迁移、DNS 配置等功能都无从谈起。

**Independent Test**: 可以通过模拟终止接口触发回收信号，验证系统是否在目标 Region 中查询最便宜的 Spot 实例并成功创建替代实例。

**Acceptance Scenarios**:

1. **Given** 程序正在 Spot CVM 上运行且自动管理器已启动, **When** 腾讯云 metadata 接口返回实例即将被回收的信号, **Then** 系统在 30 秒内自动遍历目标 Region 下所有可用 Zone，选择有可用 Spot 容量且价格最低的实例类型并发起创建请求
2. **Given** 自动管理器已启动, **When** 用户通过 API 手动触发回收, **Then** 系统执行与自动检测完全相同的替换流程
3. **Given** 创建替代实例失败（如目标 Region 无可用 Spot 容量）, **When** 系统检测到创建失败, **Then** 系统使用指数退避策略重试（最多 3 次），并记录详细错误日志
4. **Given** 同时收到多个终止信号, **When** 系统处理终止事件, **Then** 系统通过去重机制确保只创建一个替代实例

---

### User Story 2 - 程序自动迁移到新实例 (Priority: P2)

作为一名运维人员，我希望当新的 Spot CVM 创建完成后，当前实例上运行的程序能够自动迁移到新实例上并启动运行，实现完全无人值守的自愈。

**Why this priority**: 仅创建新实例不够，程序必须能在新实例上运行起来才能真正实现自愈。这是从"创建空实例"到"服务恢复"的关键一步。

**Independent Test**: 可以在新实例创建后，通过 SSH 连接新实例验证程序是否已部署并正在运行，且 API 端点可正常响应。

**Acceptance Scenarios**:

1. **Given** 新 Spot CVM 实例已创建并获得公网 IP, **When** 系统执行迁移流程, **Then** 系统通过 SCP 或等效方式将程序二进制文件和配置文件传输到新实例
2. **Given** 程序文件已传输到新实例, **When** 系统执行远程启动, **Then** 新实例上的程序成功启动并开始监听 HTTP 端口
3. **Given** 新实例上的程序已启动, **When** 旧实例被回收, **Then** 新实例上的程序继续独立运行，自动管理器继续监控自身的回收状态
4. **Given** 迁移过程中网络中断或 SSH 连接失败, **When** 系统检测到迁移失败, **Then** 系统记录错误日志并重试迁移（最多 3 次）

---

### User Story 3 - 通过 HTTP API 控制目标区域 (Priority: P3)

作为一名运维人员，我希望能够通过 HTTP API 动态指定下一个 Spot CVM 的创建区域，以便利用不同 Region 的价格差异，或在某个 Region 容量不足时切换到其他 Region。

**Why this priority**: 跨 Region 控制扩展了成本优化的范围，但系统可以在单一 Region 下正常运行。这是一个增强功能。

**Independent Test**: 可以通过 API 修改目标 Region，然后触发模拟终止，验证新实例是否在指定的新 Region 中创建。

**Acceptance Scenarios**:

1. **Given** 当前 Region 为 `ap-hongkong`, **When** 用户通过 `PUT /spot-vm/target-region` 将目标 Region 修改为 `ap-singapore`, **Then** 系统确认修改成功，后续替换实例将在 `ap-singapore` 中创建（Zone 由系统自动选择）
2. **Given** 用户指定了一个无效的 Region ID, **When** 系统验证 Region, **Then** 系统返回明确的错误信息，列出所有可用的 Region
3. **Given** 用户请求查看所有可用 Region, **When** 调用 `GET /spot-vm/regions`, **Then** 系统返回腾讯云所有可用 Region 的列表
4. **Given** 用户仅指定了目标 Region 而未指定 Zone, **When** 系统创建替代实例, **Then** 系统自动遍历该 Region 下所有 Zone，选择有可用 Spot 容量且价格最低的 Zone 进行创建

---

### User Story 4 - 新实例自动配置 DNS 和 Nginx (Priority: P4)

作为一名运维人员，我希望新 Spot CVM 创建并完成程序迁移后，能够自动修改域名的 DNS 解析指向新实例的公网 IP，并启动一个 Docker Nginx 服务器，使得域名访问能够无缝切换到新实例。

**Why this priority**: DNS 和 Nginx 配置是服务对外可达的最后一环。核心的实例管理和程序迁移可以独立运行，网络配置是锦上添花。

**Independent Test**: 可以在新实例部署完成后，检查 DNS A 记录是否已更新为新 IP，Nginx 容器是否正常运行并监听 80/443 端口。

**Acceptance Scenarios**:

1. **Given** 新实例已创建并获得公网 IP, **When** 系统执行 DNS 配置, **Then** 域名的 A 记录更新为新实例的公网 IP，并通过查询验证记录已生效
2. **Given** DNS 记录已更新, **When** 系统执行 Nginx 部署, **Then** Docker 安装完成（如未安装），Nginx 容器启动并监听 80 和 443 端口
3. **Given** 配置了 SSL 证书 ID, **When** 系统部署 Nginx, **Then** 从腾讯云获取 SSL 证书并配置 HTTPS，80 端口自动重定向到 443
4. **Given** SSL 证书即将过期（少于 30 天）, **When** 系统获取证书信息, **Then** 系统输出警告日志提醒运维人员续期

---

### Edge Cases

- 当腾讯云 metadata 服务不可达时（如程序不在腾讯云实例上运行），终止检测返回 false，不触发替换流程
- 当目标 Region 所有可用区（Zone）都没有可售卖的 Spot 实例时，系统记录错误日志并通过重试机制等待容量恢复
- 当某个 Zone 的 Spot 容量不足但同 Region 下其他 Zone 有可用容量时，系统自动跳过该 Zone 并选择其他可用 Zone
- 当新实例创建成功但 SSH 连接失败（如安全组未开放 22 端口）时，系统记录迁移失败日志并重试
- 当 DNS 更新成功但传播延迟导致域名暂时不可达时，系统不视为错误，仅记录信息日志
- 当旧实例在迁移完成前就被回收时，新实例上的程序可能未完全就绪，需要通过健康检查确认
- 当腾讯云 API 返回速率限制错误时，系统使用指数退避重试

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 每 10 秒通过腾讯云 metadata 接口 (`/latest/meta-data/spot/termination-time`) 检查当前 Spot 实例是否即将被回收
- **FR-002**: 系统 MUST 在检测到回收信号后，自动遍历目标 Region 下所有可用区（Zone），查询每个 Zone 的 Spot 实例价格和可用性，选择有可用容量且价格最低的 Zone 和实例类型创建替代实例。用户只需指定 Region，Zone 选择完全由系统自动完成
- **FR-003**: 系统 MUST 在新实例创建成功后，自动将自身程序（二进制文件 + 配置文件）迁移到新实例并启动运行
- **FR-004**: 系统 MUST 提供 REST API 供运维人员查询实例状态、管理实例生命周期、控制自动管理器和修改目标 Region
- **FR-005**: 系统 MUST 通过 API Key 认证保护所有管理端点
- **FR-006**: 系统 MUST 支持动态修改目标 Region，使替换实例可以在不同 Region 中创建。API 仅暴露 Region 级别的控制，不要求用户指定 Zone
- **FR-016**: 系统 MUST 在选择 Zone 时自动跳过无可用 Spot 容量的 Zone，优先选择价格最低的 Zone
- **FR-007**: 系统 MUST 在新实例上自动更新域名 DNS A 记录指向新实例的公网 IP
- **FR-008**: 系统 MUST 在新实例上自动安装 Docker 并启动 Nginx 容器
- **FR-009**: 系统 MUST 支持 SSL 证书配置，Nginx 提供 HTTPS 服务并将 HTTP 重定向到 HTTPS
- **FR-010**: 系统 MUST 在创建替代实例失败时使用指数退避策略重试（最多 3 次）
- **FR-011**: 系统 MUST 通过去重机制防止同时创建多个替代实例
- **FR-012**: 系统 MUST 过滤掉不可用的极低配置实例（CPU < 1 或内存 < 1024MB）
- **FR-013**: 系统 MUST 提供健康检查端点（无需认证）
- **FR-014**: 系统 MUST 支持优雅关闭，收到终止信号时先停止自动管理器再关闭 HTTP 服务器
- **FR-015**: 系统 MUST 记录所有关键操作的结构化日志

### Key Entities

- **Spot 实例 (SpotVM)**: 代表一个腾讯云竞价实例，包含实例 ID、实例类型、可用区、公网 IP、运行状态、是否即将被回收等属性
- **实例类型 (InstanceType)**: 代表一种可售卖的 Spot 实例规格，包含名称、CPU/内存配置、价格、所属可用区、售卖状态
- **自动管理器 (AutoManager)**: 自动化管理的控制器，维护当前 Region、目标 Region、运行状态，协调终止检测、实例替换和程序迁移
- **迁移任务 (MigrationTask)**: 代表一次程序迁移操作，包含源实例、目标实例、迁移状态（传输中/启动中/完成/失败）、重试次数
- **网络配置 (NetworkConfig)**: 聚合 DNS 记录管理、SSL 证书获取和 Nginx 部署的配置流程

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 从检测到回收信号到新替代实例创建请求发出，整个流程在 30 秒内完成
- **SC-002**: 程序迁移到新实例并恢复服务的总时间不超过 5 分钟（含实例启动、文件传输、程序启动）
- **SC-003**: 域名 DNS 记录在新实例就绪后 1 分钟内完成更新
- **SC-004**: 系统能够在 10 个以上腾讯云 Region 中查询和创建 Spot 实例
- **SC-005**: 自动管理器能够 7×24 小时持续运行，无内存泄漏或 goroutine 泄漏
- **SC-006**: 所有受保护的 API 端点在未提供有效凭证时 100% 返回 401 错误
- **SC-007**: 系统始终自动选择目标 Region 中有可用容量且价格最低的 Zone 和实例类型，确保成本最优，用户无需手动指定 Zone
- **SC-008**: 运维人员能够在 1 分钟内通过 API 完成 Region 切换操作
- **SC-009**: Nginx 容器在部署后 80 和 443 端口正常监听，健康检查通过

## Assumptions

- 系统运行在腾讯云 Spot CVM 实例上，能够访问 metadata 服务 (`http://metadata.tencentyun.com/latest/meta-data/`)
- 用户已在腾讯云控制台创建了 API 密钥（SecretId/SecretKey），并通过环境变量提供
- 新创建的实例使用与当前实例相同的镜像，且镜像中已预装基本的 Linux 工具（scp、ssh 等）
- 实例的安全组已开放 SSH（22）、HTTP（80）、HTTPS（443）和程序 API 端口
- 迁移方式采用 SCP 传输二进制文件 + 配置文件，通过 SSH 远程执行启动命令
- 实例登录密码通过配置文件管理，不硬编码在代码中
- DNS 管理通过腾讯云 DNSPod API 实现，支持单一域名的 A 记录更新
- 当前阶段 Nginx 仅作为简单的 Web 服务器使用，不需要复杂的反向代理配置
- 在现有 Go 项目框架基础上修改实现，不需要重新创建项目
- 单实例部署模式，不考虑多实例高可用场景
- 程序为单一 Go 二进制文件，迁移时只需传输二进制文件、`.env` 配置文件和必要的证书文件
- Zone 选择完全由系统自动完成，用户只需通过 API 指定 Region 级别的偏好，无需了解或关心具体的 Zone
