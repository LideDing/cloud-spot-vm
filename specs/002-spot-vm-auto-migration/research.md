# Research: Spot CVM 自动迁移与自愈

**Feature**: 002-spot-vm-auto-migration
**Date**: 2026-04-01
**Status**: Complete

## Research Task 1: SSH/SCP 程序自迁移方案

### Decision
使用 `golang.org/x/crypto/ssh` 包实现 SSH 客户端，通过 SCP 协议传输二进制文件和配置文件，通过 SSH 远程执行命令启动程序。

### Rationale
- Go 标准库不提供 SSH 客户端功能，`golang.org/x/crypto/ssh` 是 Go 官方维护的扩展库，稳定可靠
- SCP 基于 SSH 协议，不需要额外的服务端组件（如 FTP），腾讯云 CVM 默认开放 SSH
- 相比 rsync，SCP 更简单，且本场景只需传输少量文件（二进制 + .env + 证书）
- 密码认证方式与腾讯云 CVM 创建时设置的密码一致，无需额外配置 SSH Key

### Alternatives Considered
1. **rsync over SSH**: 功能更强大（增量传输），但本场景文件少且小，SCP 足够
2. **HTTP 文件下载**: 需要额外的文件服务器，增加复杂度
3. **腾讯云 COS 中转**: 上传到对象存储再下载，增加延迟和成本
4. **Cloud-init / User Data**: 腾讯云支持创建实例时注入 UserData 脚本，但有大小限制（16KB），且无法传输二进制文件

## Research Task 2: 新实例 SSH 就绪等待策略

### Decision
采用轮询 + 指数退避策略等待新实例 SSH 端口就绪。创建实例后，每隔 5 秒尝试 TCP 连接 22 端口，最多等待 3 分钟。

### Rationale
- 腾讯云 CVM 实例从创建到 SSH 可用通常需要 30-90 秒
- 轮询方式简单可靠，不依赖额外的通知机制
- 3 分钟超时足够覆盖大多数情况，超时后记录错误日志

### Alternatives Considered
1. **腾讯云 DescribeInstances API 轮询**: 只能确认实例状态为 RUNNING，但 SSH 服务可能尚未启动
2. **固定等待时间**: 不够灵活，可能等待过长或过短
3. **WebSocket 通知**: 过度工程化，不适合单实例场景

## Research Task 3: 获取新实例公网 IP 的方式

### Decision
创建实例后，通过腾讯云 `DescribeInstances` API 查询实例详情获取公网 IP。需要等待实例进入 RUNNING 状态后才能获取到 IP。

### Rationale
- `RunInstances` API 返回的只有 InstanceId，不包含 IP 信息
- `DescribeInstances` 是获取实例详情的标准方式
- 需要轮询等待实例状态变为 RUNNING

### Alternatives Considered
1. **Metadata API**: 只能在实例内部访问，无法从外部获取新实例的 IP
2. **弹性 IP**: 增加成本和复杂度，Spot 实例不适合绑定 EIP

## Research Task 4: 迁移文件清单

### Decision
迁移以下文件到新实例：
1. **程序二进制文件**: 当前运行的可执行文件（通过 `os.Executable()` 获取路径）
2. **.env 配置文件**: 包含所有环境变量配置
3. **SSL 证书文件**（如存在）: `/etc/nginx/ssl/cert.pem` 和 `/etc/nginx/ssl/cert.key`

### Rationale
- 程序为单一 Go 二进制文件，无需传输源码或依赖
- .env 文件包含所有运行时配置（API Key、腾讯云凭证、Region 等）
- SSL 证书文件用于 Nginx HTTPS 配置

### Alternatives Considered
1. **传输整个项目目录**: 不必要，Go 编译后为单一二进制
2. **在新实例上重新编译**: 需要安装 Go 环境，增加复杂度和时间
3. **Docker 镜像方式**: 将程序打包为 Docker 镜像，但增加了镜像管理的复杂度

## Research Task 5: Zone 自动选择策略

### Decision
沿用现有的 `getCheapestInstanceInRegion` 逻辑：遍历目标 Region 下所有可用 Zone，查询每个 Zone 的 Spot 实例价格，过滤掉低配实例（CPU < 1 或 Memory < 1024MB），选择价格最低的实例类型和对应的 Zone。

### Rationale
- 现有代码已实现了完整的 Zone 遍历和价格比较逻辑
- 用户明确表示不关心 Zone，只需指定 Region
- 按价格排序自然实现了"选择有可用容量的 Zone"（无容量的 Zone 不会返回可售卖实例）

### Alternatives Considered
1. **随机选择 Zone**: 无法保证成本最优
2. **用户指定 Zone**: 用户明确表示不想关心 Zone
3. **按延迟选择 Zone**: 增加复杂度，且 Spot 实例场景下成本优先于延迟

## Research Task 6: 新实例上程序启动方式

### Decision
通过 SSH 远程执行 `nohup` 命令在后台启动程序，确保 SSH 断开后程序继续运行。启动命令格式：
```
nohup /path/to/spot-manager > /var/log/spot-manager.log 2>&1 &
```

### Rationale
- `nohup` 确保进程不会因 SSH 会话结束而终止
- 输出重定向到日志文件便于排查问题
- 后台运行（`&`）使 SSH 命令立即返回

### Alternatives Considered
1. **systemd service**: 更规范，但需要额外创建 service 文件，增加迁移复杂度
2. **screen/tmux**: 需要额外安装，不如 nohup 简单
3. **Docker 容器化程序本身**: 增加 Docker 镜像管理的复杂度

## Research Task 7: DescribeInstances API 集成

### Decision
在 `SpotVMManager` 中新增 `GetInstanceDetails` 方法，通过 `DescribeInstances` API 查询实例详情（包括公网 IP、状态等）。

### Rationale
- 创建实例后需要获取公网 IP 用于 SSH 连接和 DNS 更新
- 现有代码中没有 `DescribeInstances` 的封装
- 该 API 是腾讯云 CVM 的标准接口

### Alternatives Considered
- 无其他合理替代方案，`DescribeInstances` 是获取实例详情的唯一标准 API
