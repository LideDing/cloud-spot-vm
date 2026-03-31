#!/bin/bash

# Spot VM API测试脚本
# 使用方法: ./test_spot_vm_api.sh [API_KEY]

# 设置默认值
API_KEY=${1:-"your-super-secret-api-key-change-in-production"}
BASE_URL="http://localhost:8080"

echo "🎯 开始测试Spot VM API..."
echo "🔑 使用的API Key: $API_KEY"
echo "🌐 基础URL: $BASE_URL"
echo ""

# 测试1: 健康检查
echo "📋 测试1: 健康检查"
curl -s -X GET "$BASE_URL/api/v1/health" | jq .
echo ""

# 测试2: 获取可用区列表
echo "📋 测试2: 获取可用区列表"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/zones" | jq .
echo ""

# 测试3: 获取最便宜的实例列表
echo "📋 测试3: 获取最便宜的实例列表"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/cheapest" | jq .
echo ""

# 测试4: 获取指定可用区的实例类型（使用第一个可用区）
echo "📋 测试4: 获取指定可用区的实例类型"
ZONE=$(curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/zones" | jq -r '.zones[0].Zone // "sa-saopaulo-1"')
echo "使用可用区: $ZONE"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/instance-types?zone=$ZONE" | jq .
echo ""

# 测试5: 创建最便宜的Spot VM（预检模式）
echo "📋 测试5: 创建最便宜的Spot VM（预检模式）"
curl -s -X POST -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"dry_run": true}' "$BASE_URL/api/v1/spot-vm/create-cheapest" | jq .
echo ""

# 测试6: 创建最便宜的Spot VM（实际创建）
echo "📋 测试6: 创建最便宜的Spot VM（实际创建）"
echo "⚠️  注意：这将实际创建Spot VM实例，可能会产生费用！"
read -p "是否继续？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    curl -s -X POST -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
      -d '{"dry_run": false}' "$BASE_URL/api/v1/spot-vm/create-cheapest" | jq .
else
    echo "跳过实际创建测试"
fi
echo ""

# 测试7: 无认证访问（应该失败）
echo "📋 测试7: 无认证访问（应该失败）"
curl -s "$BASE_URL/api/v1/spot-vm/cheapest" | jq .
echo ""

# 测试8: 错误的API key（应该失败）
echo "📋 测试8: 错误的API key（应该失败）"
curl -s -H "X-API-Key: wrong-api-key" "$BASE_URL/api/v1/spot-vm/cheapest" | jq .
echo ""

echo "✅ Spot VM API测试完成！"
echo ""
echo "📝 主要功能:"
echo "   - 获取可用区列表"
echo "   - 获取最便宜的实例列表"
echo "   - 获取指定可用区的实例类型"
echo "   - 创建最便宜的Spot VM（支持预检模式）"
echo "   - 删除实例"
echo ""
echo "🔧 使用建议:"
echo "   1. 先使用预检模式测试创建参数"
echo "   2. 确认无误后再进行实际创建"
echo "   3. 定期检查实例状态，避免意外费用"
