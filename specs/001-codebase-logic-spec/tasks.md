# Tasks: Cloud Spot VM — 腾讯云竞价实例自动管理平台

**Input**: Design documents from `/specs/001-codebase-logic-spec/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: 未在功能规格中明确要求自动化测试，因此本任务列表不包含测试任务。

**Organization**: 任务按用户故事分组，每个故事可独立实现和测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行执行（不同文件，无依赖）
- **[Story]**: 所属用户故事（US1, US2, US3, US4）
- 所有路径均为绝对路径，基于仓库根目录 `/Users/dld/work/github.com/lideding/cloud-spot-vm/`

---

## Phase 1: Setup（项目初始化）

**Purpose**: 配置化改造基础设施，将硬编码参数迁移到配置系统

- [x] T001 扩展 Config 结构体，添加 ImageId、InstancePassword、DiskType、DiskSize、Bandwidth 配置字段，在 `internal/config/config.go` 中
- [x] T002 [P] 更新 `.env.example` 文件，添加所有新增配置项的示例值和注释说明，在仓库根目录
- [x] T003 [P] 清理旧版入口 `cmd/main.go`，添加弃用注释或删除，避免与 `cmd/spot-manager/main.go` 混淆

---

## Phase 2: Foundational（阻塞性前置任务）

**Purpose**: 核心基础设施改造，所有用户故事依赖这些任务完成

**⚠️ CRITICAL**: 用户故事的实现必须等待本阶段完成

- [x] T004 修改 `cmd/spot-manager/main.go`，将硬编码的 Region (`"sa-saopaulo"`)、Domain (`"oitcep.com"`)、CertificateId (`"Qi1S1ItN"`) 替换为从 `config.Load()` 加载的配置值
- [x] T005 修改 `internal/tcc/spot_vm/manager.go` 中的 `NewSpotVMManager` 函数签名，接受配置参数（ImageId、InstancePassword、DiskType、DiskSize、Bandwidth），替换硬编码值 `"img-hdt9xxkt"`、`"1qazZSE$"`、`"CLOUD_BSSD"`、`20`、`10`
- [x] T006 修改 `internal/tcc/tcc.go` 中的 `NewTCC` 函数，将新增配置参数传递给 `NewSpotVMManager`
- [x] T007 [P] 增强错误处理：在 `internal/tcc/spot_vm/manager.go` 的 `CreateCheapestInstance` 方法中添加创建失败的错误日志和上下文信息
- [x] T008 [P] 增强日志输出：在 `internal/tcc/spot_vm/simple_auto_manager.go` 中为关键操作（启动、停止、检测回收、创建替换实例）添加结构化日志

**Checkpoint**: 基础设施就绪 — 所有硬编码已配置化，用户故事实现可以开始

---

## Phase 3: User Story 1 — 自动管理 Spot 实例生命周期 (Priority: P1) 🎯 MVP

**Goal**: 系统自动监控当前 Spot 实例，检测到回收信号时自动在目标 Region 创建最便宜的替代实例

**Independent Test**: 通过 `/auto-manager/simulate-termination` 触发模拟回收，验证系统自动查询最便宜实例并创建替代实例

### Implementation for User Story 1

- [x] T009 [US1] 优化 `internal/tcc/spot_vm/vm.go` 中的 `CheckTermination` 方法，增加 metadata 服务不可达时的超时处理（当前 HTTP 请求无超时设置），添加 `http.Client{Timeout: 5 * time.Second}`
- [x] T010 [US1] 优化 `internal/tcc/spot_vm/simple_auto_manager.go` 中的 `monitorTermination` 方法，将 `terminationCh` 的发送改为非阻塞 select 模式，防止通道满时阻塞 goroutine
- [x] T011 [US1] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 的 `handleTermination` 方法中添加重试机制：当 `CreateCheapestInstance` 失败时，使用指数退避重试（最多 3 次，间隔 5s/15s/45s）
- [x] T012 [US1] 在 `internal/tcc/spot_vm/manager.go` 的 `GetCheapestInstances` 方法中添加最低配置过滤：跳过 CPU < 1 或 Memory < 1024MB 的实例类型，避免选到不可用的极低配置
- [x] T013 [US1] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 中添加去重机制：使用 `sync.Once` 或原子标志位防止同时收到多个终止信号时创建多个替换实例

**Checkpoint**: User Story 1 完成 — 自动管理器能可靠地检测回收并创建替换实例

---

## Phase 4: User Story 2 — 通过 REST API 查询和管理 Spot 实例 (Priority: P2)

**Goal**: 运维人员通过 REST API 查看实例状态、查询价格、手动创建/删除实例

**Independent Test**: 通过 curl 调用各 API 端点，验证返回数据的正确性

### Implementation for User Story 2

- [x] T014 [P] [US2] 优化 `internal/handlers/spot_vm.go` 中的 `GetCheapestInstances` 处理器，添加 `limit` 查询参数支持（当前硬编码返回前 10 个），允许用户自定义返回数量
- [x] T015 [P] [US2] 优化 `internal/handlers/spot_vm.go` 中的 `CreateCheapestSpotVM` 处理器，在响应中添加实例创建的详细信息（实例类型、可用区、价格、DryRun 状态）
- [x] T016 [P] [US2] 优化 `internal/handlers/spot_vm.go` 中的 `GetCurrentInstanceStatus` 处理器，添加 metadata 获取失败时的友好错误提示（当前在非腾讯云环境会返回空数据）
- [x] T017 [US2] 在 `internal/handlers/spot_vm.go` 中为所有错误响应添加统一的错误码字段 `error_code`，便于客户端程序化处理错误
- [x] T018 [US2] 优化 `internal/middleware/simple_auth.go` 中的 API Key 认证中间件，添加认证失败时的详细日志（记录请求来源 IP 和使用的认证方式）

**Checkpoint**: User Story 2 完成 — 所有 REST API 端点返回完整、准确的数据

---

## Phase 5: User Story 3 — 跨 Region 切换与管理 (Priority: P3)

**Goal**: 运维人员可以动态修改目标 Region，新实例在指定 Region 中创建

**Independent Test**: 通过 `PUT /spot-vm/target-region` 修改目标 Region，触发模拟终止，验证新实例在新 Region 中创建

### Implementation for User Story 3

- [x] T019 [US3] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 的 `SetTargetRegion` 方法中添加 Region 有效性验证：调用腾讯云 DescribeRegions API 验证 Region ID 是否存在
- [x] T020 [US3] 在 `internal/tcc/spot_vm/manager.go` 中添加 `NewSpotVMManagerForRegion` 工厂方法，支持为不同 Region 创建独立的 SpotVMManager 实例（当前 Region 切换需要重建整个 TCC 客户端）
- [x] T021 [US3] 在 `internal/handlers/spot_vm.go` 的 `SetTargetRegion` 处理器中添加 Region 切换前后的状态对比信息（旧 Region、新 Region、新 Region 可用区数量）
- [x] T022 [US3] 在 `internal/handlers/spot_vm.go` 中添加 `GET /spot-vm/regions` 端点处理器，返回所有可用的腾讯云 Region 列表，方便运维人员选择目标 Region
- [x] T023 [US3] 在 `internal/routes/routes.go` 中注册新的 `GET /spot-vm/regions` 路由

**Checkpoint**: User Story 3 完成 — 支持动态 Region 切换，替换实例在目标 Region 中创建

---

## Phase 6: User Story 4 — 网络基础设施自动配置 (Priority: P4)

**Goal**: 新实例创建后自动配置 Nginx 反向代理（含 SSL）和 DNS 记录

**Independent Test**: 创建新实例后，检查 Nginx 容器是否启动、SSL 证书是否部署、DNS 记录是否更新

### Implementation for User Story 4

- [x] T024 [P] [US4] 优化 `internal/service/docker.go` 中的 `InstallDocker` 函数，添加安装结果验证（检查 `docker --version` 是否成功）和安装失败的详细错误信息
- [x] T025 [P] [US4] 优化 `internal/service/nginx.go` 中的 `DeployNginx` 方法，添加容器启动后的健康检查（检查 80/443 端口是否监听），确保 Nginx 实际可用
- [x] T026 [US4] 优化 `internal/tcc/network/dns.go` 中的 `CreateDNSRecord` 方法，添加记录创建后的验证（查询 DNS 记录确认已生效）
- [x] T027 [US4] 优化 `internal/tcc/network/ssl.go` 中的 `GetSSLCertificate` 方法，添加证书有效期检查，当证书即将过期（< 30 天）时输出警告日志
- [x] T028 [US4] 在 `internal/tcc/spot_vm/simple_auto_manager.go` 的 `handleTermination` 方法中，在创建替换实例成功后集成网络配置流程：调用 DNS 更新 → SSL 证书获取 → Nginx 部署

**Checkpoint**: User Story 4 完成 — 新实例创建后自动完成网络配置

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 跨用户故事的改进和优化

- [x] T029 [P] 在 `internal/middleware/middleware.go` 中实现 `RateLimitMiddleware`（当前为空实现），使用令牌桶算法限制每个 IP 的请求速率
- [x] T030 [P] 在 `internal/middleware/middleware.go` 中实现 `AuditMiddleware`（当前为空实现），记录所有写操作（POST/PUT/DELETE）的审计日志
- [x] T031 [P] 在 `cmd/spot-manager/main.go` 中添加优雅关闭（graceful shutdown）：监听 SIGINT/SIGTERM 信号，先停止 AutoManager，再关闭 HTTP 服务器
- [x] T032 [P] 在 `internal/tcc/spot_vm/vm.go` 中为所有 metadata API 请求添加统一的 `http.Client`（带超时和连接池配置），替换当前每次请求创建新 Client 的方式
- [ ] T033 运行 `specs/001-codebase-logic-spec/quickstart.md` 中的验证步骤，确保所有端点正常工作（需要腾讯云环境，跳过）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 无依赖 — 可立即开始
- **Foundational (Phase 2)**: 依赖 Phase 1 中的 T001 完成（新配置字段定义）— **阻塞所有用户故事**
- **User Story 1 (Phase 3)**: 依赖 Phase 2 完成
- **User Story 2 (Phase 4)**: 依赖 Phase 2 完成 — 可与 US1 并行
- **User Story 3 (Phase 5)**: 依赖 Phase 2 完成 — 可与 US1/US2 并行
- **User Story 4 (Phase 6)**: 依赖 Phase 2 完成 — 建议在 US1 之后执行（T028 依赖 handleTermination 的改进）
- **Polish (Phase 7)**: 依赖所有用户故事完成

### User Story Dependencies

- **US1 (P1)**: Phase 2 完成后可开始 — 不依赖其他用户故事
- **US2 (P2)**: Phase 2 完成后可开始 — 不依赖其他用户故事
- **US3 (P3)**: Phase 2 完成后可开始 — 不依赖其他用户故事
- **US4 (P4)**: Phase 2 完成后可开始 — T028 建议在 US1 的 T011 之后执行（重试机制先就位）

### Within Each User Story

- 模型/工具方法 → 服务层 → 处理器层 → 路由注册
- 核心实现 → 集成优化

### Parallel Opportunities

- Phase 1: T002 和 T003 可与 T001 并行
- Phase 2: T007 和 T008 可并行（与 T004-T006 的串行链并行）
- Phase 3 (US1): T009 和 T010 可并行
- Phase 4 (US2): T014、T015、T016 可并行
- Phase 6 (US4): T024 和 T025 可并行
- Phase 7: T029、T030、T031、T032 全部可并行

---

## Parallel Example: User Story 2

```bash
# 并行启动所有独立的处理器优化任务：
Task T014: "优化 GetCheapestInstances 处理器添加 limit 参数"
Task T015: "优化 CreateCheapestSpotVM 处理器添加详细响应"
Task T016: "优化 GetCurrentInstanceStatus 处理器添加友好错误提示"

# 上述任务完成后，串行执行依赖任务：
Task T017: "为所有错误响应添加统一错误码"
Task T018: "优化认证中间件添加详细日志"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup（配置化改造）
2. 完成 Phase 2: Foundational（硬编码迁移）— **CRITICAL**
3. 完成 Phase 3: User Story 1（自动管理器可靠性增强）
4. **STOP and VALIDATE**: 通过模拟终止验证自动替换流程
5. 如果验证通过，可部署 MVP

### Incremental Delivery

1. Setup + Foundational → 配置化基础就绪
2. + User Story 1 → 自动管理器可靠运行 → **MVP 部署**
3. + User Story 2 → REST API 完善 → 运维体验提升
4. + User Story 3 → 跨 Region 支持 → 成本优化范围扩大
5. + User Story 4 → 网络自动配置 → 全自动化运维
6. + Polish → 生产级质量

### Parallel Team Strategy

多人协作时：

1. 团队共同完成 Setup + Foundational
2. Foundational 完成后：
   - 开发者 A: User Story 1（核心自动管理）
   - 开发者 B: User Story 2（REST API 优化）
   - 开发者 C: User Story 3（跨 Region 支持）
3. US1 完成后开发者 A 接手 User Story 4（依赖 handleTermination 改进）
4. 所有故事完成后团队共同完成 Polish

---

## Notes

- [P] 任务 = 不同文件，无依赖，可并行执行
- [Story] 标签将任务映射到特定用户故事，便于追踪
- 每个用户故事可独立完成和测试
- 每个任务或逻辑组完成后提交代码
- 在任何 Checkpoint 处可停下来独立验证该故事
- 避免：模糊任务、同文件冲突、破坏独立性的跨故事依赖
