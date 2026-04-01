# Implementation Plan: Cloud Spot VM — 腾讯云竞价实例自动管理平台

**Branch**: `001-codebase-logic-spec` | **Date**: 2026-03-31 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-codebase-logic-spec/spec.md`

## Summary

Cloud Spot VM 是一个用 Go 编写的腾讯云竞价实例自动管理平台。核心功能是通过 metadata API 监控当前 Spot 实例的回收信号，在检测到回收时自动在目标 Region 中查找并创建最便宜的替代 Spot 实例，同时提供 REST API 供运维人员手动管理实例、查询价格和控制自动管理器。附加功能包括通过 Docker 部署 Nginx 反向代理（含 SSL）和 DNS 记录自动更新。

## Technical Context

**Language/Version**: Go 1.24.5  
**Primary Dependencies**:
- `github.com/gin-gonic/gin v1.10.1` — HTTP Web 框架
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.1.6` — 腾讯云 SDK（CommonClient 模式）
- `github.com/golang-jwt/jwt/v5 v5.3.0` — JWT 认证（已引入但未在主流程中使用）
- `github.com/joho/godotenv v1.5.1` — .env 文件加载
- `github.com/google/uuid v1.6.0` — UUID 生成（已引入但未在主流程中使用）

**Storage**: N/A — 无持久化存储，所有状态保存在内存中  
**Testing**: 无自动化测试（当前仅有 shell 脚本手动测试）  
**Target Platform**: Linux 服务器（腾讯云 CVM 实例）  
**Project Type**: Web Service（REST API + 后台自动管理器）  
**Performance Goals**: 回收检测到替换实例创建 < 30 秒  
**Constraints**: 单实例部署，无数据库依赖，需要访问腾讯云 metadata 服务  
**Scale/Scope**: 单用户运维工具，管理 1 个 Spot 实例的生命周期

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Cost Minimization First | ✅ PASS | 系统核心逻辑就是选择最便宜的 Spot 实例，按 `UnitPriceDiscount` 排序选择 |
| II. Resilience & Availability | ⚠️ PARTIAL | 回收检测和自动替换已实现；但缺少重试机制、创建失败无恢复策略、terminationCh 缓冲区仅为 1 可能丢失信号 |
| III. Configuration-Driven Design | ⚠️ PARTIAL | 大部分配置通过 .env 加载；但 Region、Domain、CertificateId 在 main.go 中硬编码，ImageId 和 Password 在 manager.go 中硬编码 |
| IV. Security & Credential Safety | ⚠️ PARTIAL | 凭证通过环境变量加载；但实例登录密码 `1qazZSE$` 硬编码在代码中，JWT Secret 有默认值 |
| V. Simplicity & Maintainability | ✅ PASS | 代码结构清晰，遵循 Go 惯例（cmd/、internal/），包分离合理 |

**Gate Result**: ⚠️ PASS WITH WARNINGS — 无阻塞性违规，但有多项需要改进的地方已记录在 Complexity Tracking 中。

## Project Structure

### Documentation (this feature)

```text
specs/001-codebase-logic-spec/
├── plan.md              # 本文件（/speckit.plan 输出）
├── research.md          # Phase 0 输出
├── data-model.md        # Phase 1 输出
├── quickstart.md        # Phase 1 输出
└── contracts/           # Phase 1 输出（REST API 契约）
    └── api-v1.md
```

### Source Code (repository root)

```text
cloud-spot-vm/
├── cmd/
│   ├── main.go                          # 早期入口（已弃用，使用旧版 SpotService）
│   └── spot-manager/
│       └── main.go                      # 主入口：加载配置 → 创建 TCC → 启动 AutoManager → 启动 Gin 服务
├── internal/
│   ├── config/
│   │   └── config.go                    # 配置加载（.env + 环境变量）
│   ├── models/
│   │   └── models.go                    # 数据模型（Zone, InstanceType, Price, SSL 响应等）
│   ├── handlers/
│   │   ├── spot_vm.go                   # Spot VM API 处理器（CRUD + 自动管理器控制）
│   │   ├── simple_auth.go               # 认证处理器（健康检查 + API Key 验证）
│   │   └── simple_user.go               # 用户处理器（预留）
│   ├── middleware/
│   │   ├── middleware.go                # 通用中间件（JWT、CORS、日志、速率限制、审计）
│   │   └── simple_auth.go              # API Key 认证中间件
│   ├── routes/
│   │   └── routes.go                    # Gin 路由注册
│   ├── service/
│   │   ├── spot.go                      # SpotService 编排服务（早期版本，已被 AutoManager 替代）
│   │   ├── nginx.go                     # Nginx 服务（Docker 容器部署 + SSL 配置）
│   │   └── docker.go                    # Docker 安装（支持 apt-get/yum/dnf）
│   └── tcc/
│       ├── tcc.go                       # TCC 统一入口（聚合 SpotVMManager + AutoManager + SSL + DNS）
│       ├── spot_vm/
│       │   ├── manager.go              # SpotVMManager：腾讯云 CVM API 封装（查询/创建/删除实例）
│       │   ├── simple_auto_manager.go  # SimpleAutoManager：自动监控 + 替换核心逻辑
│       │   └── vm.go                   # SpotVM：metadata API 状态采集 + 终止检测
│       └── network/
│           ├── dns.go                   # DNSManager：DNSPod API 封装（创建 A 记录）
│           └── ssl.go                   # SSLManager：SSL 证书 API 封装（获取证书详情）
├── .env                                 # 环境变量配置（不提交到 Git）
├── go.mod                               # Go 模块定义
└── go.sum                               # 依赖校验
```

**Structure Decision**: 项目采用标准 Go 项目布局，`cmd/` 存放入口，`internal/` 存放内部包。当前有两个入口点：`cmd/main.go`（早期版本，使用 SpotService）和 `cmd/spot-manager/main.go`（当前版本，使用 SimpleAutoManager + REST API）。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Region/Domain/CertificateId 硬编码在 main.go | 快速原型开发 | 应迁移到 config.go 通过环境变量加载 |
| 实例密码硬编码在 manager.go | 快速原型开发 | 应迁移到配置或使用 SSH Key 登录 |
| ImageId 硬编码 `img-hdt9xxkt` | 特定 Region 镜像 | 应配置化，不同 Region 可能需要不同镜像 |
| 无重试机制 | 简化实现 | 生产环境需要指数退避重试 |
| terminationCh 缓冲区为 1 | 简化实现 | 可能丢失信号，应增加去重机制 |
| SpotService (cmd/main.go) 与 SimpleAutoManager 并存 | 历史遗留 | 应清理旧版入口 |
