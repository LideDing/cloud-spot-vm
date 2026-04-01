# Tasks: Spot CVM 自动迁移与自愈

**Input**: Design documents from `/specs/002-spot-vm-auto-migration/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api-v1.md ✅

**Tests**: 未在 spec 中明确要求 TDD，不生成测试任务。

**Organization**: 任务按用户故事分组，每个故事可独立实现和测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行执行（不同文件，无依赖）
- **[Story]**: 任务所属用户故事（US1, US2, US3, US4）
- 所有路径均为相对于项目根目录的路径

---

## Phase 1: Setup（项目初始化）

**Purpose**: 新增依赖、扩展配置结构、新增数据模型

- [x] T001 添加 `golang.org/x/crypto` 依赖到 `go.mod` 并运行 `go mod tidy`
- [x] T002 [P] 扩展 Config 结构体，新增 SSH 和迁移相关配置项（SSHPort, SSHTimeout, SSHWaitTimeout, MigrationMaxRetries, RemoteBinaryPath, RemoteEnvPath）并从 .env 加载，在 `internal/config/config.go` 中实现
- [x] T003 [P] 新增 MigrationTask 模型和 MigrationStatus 枚举（PENDING, WAITING_SSH, TRANSFERRING, STARTING, VERIFYING, COMPLETED, FAILED），在 `internal/models/models.go` 中实现
- [x] T004 [P] 将 `cmd/spot-manager/main.go` 和 `internal/tcc/tcc.go` 中硬编码的 Region/Domain/CertificateId 改为从 Config 读取（修复 spec-001 发现的问题）

---

## Phase 2: Foundational（阻塞性前置任务）

**Purpose**: 核心基础设施，所有用户故事都依赖这些任务

**⚠️ CRITICAL**: 必须完成此阶段后才能开始用户故事

- [x] T005 在 `internal/tcc/spot_vm/manager.go` 中新增 `GetInstanceDetails` 方法，封装腾讯云 `DescribeInstances` API，支持通过 InstanceId 查询实例详情（公网 IP、状态等）
- [x] T006 在 `internal/tcc/spot_vm/manager.go` 中新增 `WaitForInstanceRunning` 方法，轮询 `GetInstanceDetails` 等待实例状态变为 RUNNING 并返回公网 IP（超时可配置）
- [x] T007 [P] 创建 `internal/migration/migrator.go`，实现 Migrator 结构体骨架：SSH 客户端配置、密码认证、连接管理、接口定义（Connect, TransferFile, ExecuteCommand, Close）
- [x] T008 在 `internal/tcc/tcc.go` 中注册 Migrator 实例，将其注入到 TCC 结构体中，使 SimpleAutoManager 可以访问

**Checkpoint**: 基础设施就绪 — 用户故事实现可以开始

---

## Phase 3: User Story 1 — 自动检测回收并创建替代实例 (Priority: P1) 🎯 MVP

**Goal**: 程序在 Spot CVM 上持续运行，检测到回收信号后自动在目标 Region 中选择最便宜的 Zone 创建替代实例

**Independent Test**: 通过 `POST /api/v1/spot-vm/auto-manager/simulate-termination` 触发模拟终止，验证系统在目标 Region 中查询最便宜的 Spot 实例并成功创建替代实例

### Implementation for User Story 1

- [x] T009 [US1] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中重构 `handleTermination` 方法，集成完整的替换流程：遍历目标 Region 所有 Zone → 选择最便宜实例 → 创建替代实例 → 调用 `WaitForInstanceRunning` 获取公网 IP
- [x] T010 [US1] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中确保指数退避重试逻辑（最多 3 次）和去重机制在新流程中正确工作
- [x] T011 [US1] 在 `internal/handlers/spot_vm.go` 中更新 `SimulateTermination` 和 `TriggerTermination` 处理器，确保触发完整的替换流程（含获取新实例 IP）
- [x] T012 [US1] 在 `internal/handlers/spot_vm.go` 中更新 `GetAutoManagerStatus` 处理器，返回当前迁移状态信息（migration_status 字段）

**Checkpoint**: US1 完成 — 系统可自动检测回收并创建替代实例，可通过 API 验证

---

## Phase 4: User Story 2 — 程序自动迁移到新实例 (Priority: P2)

**Goal**: 新 Spot CVM 创建完成后，自动通过 SSH/SCP 将程序迁移到新实例并启动运行

**Independent Test**: 创建新实例后，通过 SSH 连接新实例验证程序已部署并正在运行，API 端点可正常响应

### Implementation for User Story 2

- [x] T013 [US2] 在 `internal/migration/migrator.go` 中实现 `WaitForSSH` 方法：轮询 TCP 连接目标 IP 的 SSH 端口，每 5 秒重试，超时时间从配置读取（默认 180 秒）
- [x] T014 [US2] 在 `internal/migration/migrator.go` 中实现 `Connect` 方法：使用 `golang.org/x/crypto/ssh` 建立 SSH 连接，密码认证方式，密码从 Config.InstancePassword 读取
- [x] T015 [US2] 在 `internal/migration/migrator.go` 中实现 `TransferFile` 方法：通过 SCP 协议传输文件到远程实例（支持设置远程路径和文件权限）
- [x] T016 [US2] 在 `internal/migration/migrator.go` 中实现 `ExecuteCommand` 方法：通过 SSH Session 执行远程命令并返回输出
- [x] T017 [US2] 在 `internal/migration/migrator.go` 中实现 `Migrate` 方法：编排完整迁移流程 — 等待 SSH → 创建远程目录 → SCP 传输二进制文件（通过 `os.Executable()` 获取路径）→ SCP 传输 .env → SCP 传输 SSL 证书（如存在）→ `nohup` 远程启动程序 → 健康检查验证
- [x] T018 [US2] 在 `internal/migration/migrator.go` 中实现 `HealthCheck` 方法：通过 HTTP GET 请求新实例的 `/api/v1/health` 端点验证程序运行正常，最多重试 10 次，每次间隔 5 秒
- [x] T019 [US2] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中集成迁移流程：创建替代实例成功后，创建 MigrationTask，调用 Migrator.Migrate 执行迁移，更新迁移状态

**Checkpoint**: US2 完成 — 系统可自动将程序迁移到新实例并启动运行

---

## Phase 5: User Story 3 — 通过 HTTP API 控制目标区域 (Priority: P3)

**Goal**: 运维人员可通过 HTTP API 动态指定下一个 Spot CVM 的创建 Region

**Independent Test**: 通过 API 修改目标 Region，然后触发模拟终止，验证新实例在指定的新 Region 中创建

### Implementation for User Story 3

- [x] T020 [P] [US3] 在 `internal/tcc/spot_vm/manager.go` 中新增 `GetRegions` 方法，封装腾讯云 `DescribeRegions` API，返回所有可用 Region 列表
- [x] T021 [P] [US3] 在 `internal/handlers/spot_vm.go` 中实现 `GetRegions` 处理器，调用 `GetRegions` 返回可用 Region 列表（对应 `GET /api/v1/spot-vm/regions`）
- [x] T022 [US3] 在 `internal/handlers/spot_vm.go` 中更新 `UpdateTargetRegion` 处理器，增加 Region 有效性验证（调用 `GetRegions` 校验），无效 Region 返回 400 错误并列出可用 Region
- [x] T023 [US3] 在 `internal/routes/routes.go` 中注册新增的 `GET /api/v1/spot-vm/regions` 路由

**Checkpoint**: US3 完成 — 运维人员可通过 API 查看和切换目标 Region

---

## Phase 6: User Story 4 — 新实例自动配置 DNS 和 Nginx (Priority: P4)

**Goal**: 新实例完成程序迁移后，自动更新 DNS A 记录并部署 Nginx

**Independent Test**: 新实例部署完成后，检查 DNS A 记录是否已更新为新 IP，Nginx 容器是否正常运行并监听 80/443 端口

### Implementation for User Story 4

- [x] T024 [US4] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中集成 DNS 更新流程：迁移完成后，调用 DNSManager 更新域名 A 记录指向新实例公网 IP
- [x] T025 [US4] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中集成 Nginx 部署流程：DNS 更新后，通过 SSH 在新实例上执行 Docker 安装和 Nginx 容器部署（复用现有 `docker.go` 和 `nginx.go` 的逻辑，但通过 SSH 远程执行）
- [x] T026 [US4] 在 `internal/migration/migrator.go` 中新增 `DeployNginxRemotely` 方法：通过 SSH 在远程实例上执行 Docker 安装命令、拉取 Nginx 镜像、启动容器（含 SSL 配置）
- [x] T027 [US4] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中集成 SSL 证书获取：部署 Nginx 前，通过 SSLManager 获取证书内容，SCP 传输到新实例的 `/etc/nginx/ssl/` 目录

**Checkpoint**: US4 完成 — 新实例自动完成 DNS 更新和 Nginx 部署，域名访问无缝切换

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 跨故事的改进和优化

- [x] T028 [P] 在 `cmd/spot-manager/main.go` 中完善启动日志，输出所有新增配置项和迁移相关信息
- [x] T029 [P] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中添加完整的结构化日志：迁移各阶段的开始/完成/失败日志
- [x] T030 [P] 更新 `.env.example` 文件（如存在）或在项目根目录创建，包含所有新增配置项及注释说明
- [x] T031 在 `cmd/spot-manager/main.go` 中实现优雅关闭：监听 SIGINT/SIGTERM 信号，先停止 AutoManager 再关闭 HTTP 服务器
- [x] T032 端到端验证：按照 `specs/002-spot-vm-auto-migration/quickstart.md` 执行完整流程验证，确保所有功能正常工作

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 无依赖 — 可立即开始
- **Foundational (Phase 2)**: 依赖 Setup 完成 — **阻塞所有用户故事**
- **User Story 1 (Phase 3)**: 依赖 Foundational 完成
- **User Story 2 (Phase 4)**: 依赖 Foundational 完成（可与 US1 并行，但建议在 US1 之后，因为迁移依赖实例创建）
- **User Story 3 (Phase 5)**: 依赖 Foundational 完成（可与 US1/US2 并行）
- **User Story 4 (Phase 6)**: 依赖 US2 完成（DNS/Nginx 部署在迁移之后执行）
- **Polish (Phase 7)**: 依赖所有用户故事完成

### User Story Dependencies

- **US1 (P1)**: Foundational 完成后即可开始 — 无其他故事依赖
- **US2 (P2)**: Foundational 完成后即可开始 — 与 US1 可并行，但集成时需要 US1 的实例创建能力
- **US3 (P3)**: Foundational 完成后即可开始 — 完全独立，可与 US1/US2 并行
- **US4 (P4)**: 依赖 US2 的迁移能力 — 必须在 US2 之后

### Within Each User Story

- 模型/方法定义 → 服务层实现 → 处理器/路由集成
- 核心实现 → 集成到 AutoManager

### Parallel Opportunities

- **Phase 1**: T002, T003, T004 可并行（不同文件）
- **Phase 2**: T005+T006（同文件，顺序执行）与 T007（不同文件）可并行
- **Phase 5**: T020, T021 可并行（不同文件）
- **Phase 7**: T028, T029, T030 可并行（不同文件）

---

## Parallel Example: Phase 1 Setup

```bash
# 先完成 T001（添加依赖），然后并行执行：
Task T002: "扩展 Config 结构体 in internal/config/config.go"
Task T003: "新增 MigrationTask 模型 in internal/models/models.go"
Task T004: "修复硬编码问题 in cmd/spot-manager/main.go + internal/tcc/tcc.go"
```

## Parallel Example: Phase 2 Foundational

```bash
# 并行执行（不同文件）：
Task T005+T006: "DescribeInstances + WaitForRunning in internal/tcc/spot_vm/manager.go"
Task T007: "Migrator 骨架 in internal/migration/migrator.go"
# T008 依赖 T007 完成
```

## Parallel Example: Phase 5 User Story 3

```bash
# 并行执行（不同文件）：
Task T020: "GetRegions 方法 in internal/tcc/spot_vm/manager.go"
Task T021: "GetRegions 处理器 in internal/handlers/spot_vm.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup（T001-T004）
2. 完成 Phase 2: Foundational（T005-T008）— **CRITICAL: 阻塞所有故事**
3. 完成 Phase 3: User Story 1（T009-T012）
4. **STOP and VALIDATE**: 通过模拟终止验证自动创建替代实例
5. 可部署/演示 MVP

### Incremental Delivery

1. Setup + Foundational → 基础设施就绪
2. + User Story 1 → 自动检测回收并创建替代实例 → **MVP!**
3. + User Story 2 → 程序自动迁移到新实例 → 完全自愈闭环
4. + User Story 3 → HTTP API 控制目标区域 → 跨 Region 灵活切换
5. + User Story 4 → DNS + Nginx 自动配置 → 域名访问无缝切换
6. + Polish → 日志完善、优雅关闭、端到端验证

### Single Developer Strategy

按优先级顺序执行：Phase 1 → Phase 2 → Phase 3 (US1) → Phase 4 (US2) → Phase 5 (US3) → Phase 6 (US4) → Phase 7

---

## Notes

- [P] 任务 = 不同文件，无依赖，可并行
- [Story] 标签将任务映射到具体用户故事，便于追踪
- 每个用户故事应可独立完成和测试
- 每个任务或逻辑组完成后提交代码
- 在任何 Checkpoint 处可停下来独立验证故事
- 避免：模糊任务、同文件冲突、破坏独立性的跨故事依赖
