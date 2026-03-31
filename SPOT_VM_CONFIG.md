# Spot VM 配置指南

## 概述

本指南将帮助您配置Spot VM API，以便能够成功创建腾讯云Spot实例。

## 前置条件

1. **腾讯云账户**：确保您有有效的腾讯云账户
2. **API密钥**：配置好腾讯云API密钥（SecretId和SecretKey）
3. **权限**：确保API密钥有CVM相关权限

## 环境变量配置

在 `.env` 文件中设置以下环境变量：

```bash
# 腾讯云API密钥
TENCENTCLOUD_SECRET_ID=your-secret-id
TENCENTCLOUD_SECRET_KEY=your-secret-key

# API服务器配置
API_KEY=your-super-secret-api-key-change-in-production
PORT=8080
```

## 腾讯云资源配置

### 1. 获取VPC和子网信息

```bash
# 使用腾讯云CLI获取VPC列表
tccli vpc DescribeVpcs

# 获取子网列表
tccli vpc DescribeSubnets --vpc-id vpc-xxxxxxxx
```

### 2. 获取可用镜像

```bash
# 获取公共镜像列表
tccli cvm DescribeImages --ImageType PUBLIC_IMAGE
```

### 3. 修改配置文件

编辑 `internal/tcc/spot_vm/manager.go` 文件，修改以下配置：

```go
// 在CreateSpotInstance函数中修改以下参数
params := map[string]any{
    "InstanceChargeType": "SPOTPAID",
    "Placement":          map[string]any{"Zone": zone},
    "InstanceType":       instanceType,
    "ImageId":            "img-xxxxxxxx", // 替换为你的镜像ID
    "SystemDisk": map[string]any{
        "DiskType": "CLOUD_BSSD",
        "DiskSize": 20,
    },
    "VirtualPrivateCloud": map[string]any{
        "VpcId":    "vpc-xxxxxxxx",    // 替换为你的VPC ID
        "SubnetId": "subnet-xxxxxxxx", // 替换为你的子网ID
    },
    "InternetAccessible": map[string]any{
        "InternetChargeType":      "TRAFFIC_POSTPAID_BY_HOUR",
        "InternetMaxBandwidthOut": 10,
        "PublicIpAssigned":        true,
    },
    "InstanceCount": 1,
    "LoginSettings": map[string]any{"Password": "YourPassword123!"}, // 修改密码
    "EnhancedService": map[string]any{
        "SecurityService":   map[string]any{"Enabled": true},
        "MonitorService":    map[string]any{"Enabled": true},
        "AutomationService": map[string]any{"Enabled": true},
    },
    "TagSpecification": []map[string]any{
        {"ResourceType": "instance", "Tags": []map[string]any{
            {"Key": "owner", "Value": "your-name"},
            {"Key": "env", "Value": "lab"},
            {"Key": "projid", "Value": "spot-vm"},
            {"Key": "service", "Value": "infra"},
        }},
    },
    "DryRun": dryRun,
}
```

## 测试配置

### 1. 编译项目

```bash
go build -o spot-vm cmd/main.go
```

### 2. 启动服务器

```bash
./spot-vm
```

### 3. 测试API

```bash
# 测试健康检查
curl http://localhost:8080/api/v1/health

# 测试获取可用区
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/spot-vm/zones

# 测试预检模式创建
curl -X POST -H "X-API-Key: your-api-key" -H "Content-Type: application/json" \
  -d '{"dry_run": true}' http://localhost:8080/api/v1/spot-vm/create-cheapest
```

## 常见问题

### 1. VPC不存在错误

```
Code=InvalidParameterValue.VpcIdNotExist
Message=The vpcId [`vpc-xxxxxxxx`] you specified does not exist.
```

**解决方案**：
- 检查VPC ID是否正确
- 确保VPC在指定的Region中存在
- 使用腾讯云控制台或CLI确认VPC ID

### 2. 子网不存在错误

```
Code=InvalidParameterValue.SubnetIdNotExist
Message=The subnetId [`subnet-xxxxxxxx`] you specified does not exist.
```

**解决方案**：
- 检查子网ID是否正确
- 确保子网在指定的VPC中存在
- 确保子网在指定的可用区中存在

### 3. 镜像不存在错误

```
Code=InvalidParameterValue.ImageIdNotExist
Message=The imageId [`img-xxxxxxxx`] you specified does not exist.
```

**解决方案**：
- 检查镜像ID是否正确
- 确保镜像在指定的Region中存在
- 使用公共镜像或确保有权限使用私有镜像

### 4. 权限不足错误

```
Code=UnauthorizedOperation
Message=You are not authorized to perform this operation.
```

**解决方案**：
- 检查API密钥权限
- 确保有CVM相关权限
- 联系腾讯云管理员分配权限

## 安全建议

1. **密码安全**：使用强密码，避免在代码中硬编码
2. **API密钥**：定期轮换API密钥
3. **网络配置**：配置安全组规则，限制访问
4. **监控**：启用云监控，监控实例状态
5. **备份**：定期备份重要数据

## 成本优化

1. **使用Spot实例**：Spot实例比按量计费实例便宜
2. **监控使用情况**：定期检查实例使用情况
3. **自动清理**：设置自动清理机制，避免闲置实例
4. **资源标签**：使用标签管理资源，便于成本分析

## 支持

如果遇到问题，请：

1. 检查腾讯云控制台日志
2. 查看API响应错误信息
3. 参考腾讯云官方文档
4. 联系腾讯云技术支持
