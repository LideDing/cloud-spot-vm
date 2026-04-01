# Implementation Plan: Spot CVM 自动迁移与自愈

**Branch**: `002-spot-vm-auto-migration` | **Date**: 2026-04-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-spot-vm-auto-migration/spec.md`

## Summary

在现有腾讯云 Spot CVM 自动管理平台基础上，新增程序自迁移能力。当检测到当前 Spot 实例即将被回收时，系统自动在目标 Region 中遍历所有 Zone，选择有可用容量且价格最低的 Zone 创建替代实例，然后通过 SSH/SCP 将自身程序迁移到新实例并启动运行，最后自动更新 DNS 记录并部署 Nginx。用户只需通过 HTTP API 指定 Region，Zone 选择完全由系统自动完成。

## Technical Context

**Language/Version**: Go 1.24.5
**Primary Dependencies**:
- `github.com/gin-gonic/gin v1.10.1` — HTTP Web 框架
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.1.6` — 腾讯云 SDK（CommonClient 模式）
- `golang.org/x/crypto/ssh` — SSH 客户端（新增，用于 SCP 文件传输和远程命令执行）
- `github.com/joho/godotenv v1.5.1` — .env 文件加载
- `github.com/golang-jwt/jwt/v5 v5.3.0` — JWT 认证（已引入）

**Storage**: N/A — 无持久化存储，所有状态保存在内存中
**Testing**: Shell 脚本手动测试 + 模拟终止接口
**Target Platform**: Linux 服务器（腾讯云 CVM 实例，Ubuntu/CentOS）
**Project Type**: Web Service（REST API + 后台自动管理器 + 自迁移引擎）
**Performance Goals**: 回收检测到替换实例创建 < 30 秒；完整迁移恢复 < 5 分钟
**Constraints**: 单实例部署，无数据库依赖，需要访问腾讯云 metadata 服务，需要 SSH 22 端口开放
**Scale/Scope**: 单用户运维工具，管理 1 个 Spot 实例的生命周期

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Cost Minimization First | ✅ PASS | 自动遍历所有 Zone 选择最便宜的实例类型；Zone 选择完全自动化，用户只需指定 Region |
| II. Resilience & Availability | ✅ PASS | 指数退避重试（最多 3 次）、去重机制防止重复创建、自迁移实现完全自愈闭环、SSH 连接失败重试 |
| III. Configuration-Driven Design | ✅ PASS | 本次将修复 spec-001 发现的硬编码问题：Region/Domain/CertificateId/InstancePassword 全部通过 .env 配置；新增 SSH 相关配置项 |
| IV. Security & Credential Safety | ✅ PASS | 实例密码通过环境变量配置，不硬编码；SSH 连接使用密码认证（从配置读取）；API Key 保护所有管理端点 |
| V. Simplicity & Maintainability | ✅ PASS | 在现有框架上增量修改，新增 `internal/migration/` 包封装迁移逻辑，保持包分离清晰 |

**Gate Result**: ✅ PASS — 所有原则均满足，spec-001 中的 PARTIAL 项在本次实现中修复。

## Project Structure

### Documentation (this feature)

```text
specs/002-spot-vm-auto-migration/
├── plan.md              # 本文件（/speckit.plan 输出）
├── research.md          # Phase 0 输出
├── data-model.md        # Phase 1 输出
├── quickstart.md        # Phase 1 输出
├── contracts/           # Phase 1 输出（REST API 契约）
│   └── api-v1.md
└── checklists/
    └── requirements.md  # 规范质量检查清单
```

### Source Code (repository root)

```text
cloud-spot-vm/
├── cmd/
│   ├── main.go                          # 早期入口（已弃用）
│   └── spot-manager/
│       └── main.go                      # 主入口：加载配置 → 创建 TCC → 启动 AutoManager → 启动 Gin 服务
├── internal/
│   ├── config/
│   │   └── config.go                    # 配置加载（.env + 环境变量）— 新增迁移相关配置项
│   ├── models/
│   │   └── models.go                    # 数据模型 — 新增 MigrationTask 模型
│   ├── handlers/
│   │   ├── spot_vm.go                   # Spot VM API 处理器
│   │   ├── simple_auth.go               # 认证处理器
│   │   └── simple_user.go               # 用户处理器（预留）
│   ├── middleware/
│   │   ├── middleware.go                # 通用中间件
│   │   └── simple_auth.go              # API Key 认证中间件
│   ├── routes/
│   │   └── routes.go                    # Gin 路由注册
│   ├── migration/                       # 【新增】程序自迁移模块
│   │   └── migrator.go                  # SSH/SCP 迁移引擎：文件传输 + 远程启动 + 重试
│   ├── service/
│   │   ├── spot.go                      # SpotService（早期版本，已弃用）
│   │   ├── nginx.go                     # Nginx 服务（Docker 容器部署 + SSL 配置）
│   │   ├── docker.go                    # Docker 安装
│   │   └── conf/
│   │       └── nginx.conf               # Nginx 配置模板
│   └── tcc/
│       ├── tcc.go                       # TCC 统一入口 — 注册迁移回调
│       ├── spot_vm/
│       │   ├── manager.go              # SpotVMManager：腾讯云 CVM API 封装
│       │   ├── simple_auto_manager.go  # SimpleAutoManager — 集成迁移流程
│       │   └── vm.go                   # SpotVM：metadata API 状态采集 + 终止检测
│       └── network/
│           ├── dns.go                   # DNSManager：DNSPod API 封装
│           └── ssl.go                   # SSLManager：SSL 证书 API 封装
├── .env                                 # 环境变量配置（不提交到 Git）
├── go.mod                               # Go 模块定义 — 新增 golang.org/x/crypto 依赖
└── go.sum                               # 依赖校验
```

**Structure Decision**: 在现有项目结构基础上增量修改。新增 `internal/migration/` 包封装 SSH/SCP 迁移逻辑，保持与现有包的分离。核心变更集中在 `simple_auto_manager.go`（集成迁移流程）和 `tcc.go`（注册迁移回调）。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 新增 `golang.org/x/crypto/ssh` 依赖 | SSH/SCP 是实现程序自迁移的唯一可靠方式 | 标准库不提供 SSH 客户端功能 |
| 新增 `internal/migration/` 包 | 迁移逻辑复杂度较高（SSH 连接、文件传输、远程执行、重试），需要独立封装 | 放在 `simple_auto_manager.go` 中会导致文件过大、职责不清 |
