#!/bin/bash

# API测试脚本
# 使用方法: ./test_api.sh [API_KEY]

# 设置默认值
API_KEY=${1:-"your-super-secret-api-key-change-in-production"}
BASE_URL="http://localhost:8080"

echo "🧪 开始测试API key认证系统..."
echo "🔑 使用的API Key: $API_KEY"
echo "🌐 基础URL: $BASE_URL"
echo ""

# 测试1: 健康检查（无需认证）
echo "📋 测试1: 健康检查（无需认证）"
curl -s -X GET "$BASE_URL/api/v1/health" | jq .
echo ""

# 测试2: 验证API key
echo "📋 测试2: 验证API key"
curl -s -X POST "$BASE_URL/api/v1/auth/validate" \
  -H "Content-Type: application/json" \
  -d "{\"api_key\": \"$API_KEY\"}" | jq .
echo ""

# 测试3: 使用错误的API key验证
echo "📋 测试3: 使用错误的API key验证"
curl -s -X POST "$BASE_URL/api/v1/auth/validate" \
  -H "Content-Type: application/json" \
  -d "{\"api_key\": \"wrong-api-key\"}" | jq .
echo ""

# 测试4: 访问受保护的端点（使用X-API-Key头）
echo "📋 测试4: 访问受保护的端点（使用X-API-Key头）"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/protected" | jq .
echo ""

# 测试5: 访问受保护的端点（使用Authorization头）
echo "📋 测试5: 访问受保护的端点（使用Authorization头）"
curl -s -H "Authorization: Bearer $API_KEY" "$BASE_URL/api/v1/protected" | jq .
echo ""

# 测试6: 访问受保护的端点（使用查询参数）
echo "📋 测试6: 访问受保护的端点（使用查询参数）"
curl -s "$BASE_URL/api/v1/protected?api_key=$API_KEY" | jq .
echo ""

# 测试7: 访问受保护的端点（无认证）
echo "📋 测试7: 访问受保护的端点（无认证）"
curl -s "$BASE_URL/api/v1/protected" | jq .
echo ""

# 测试8: 获取实例信息（需要认证）
echo "📋 测试8: 获取实例信息（需要认证）"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/instances" | jq .
echo ""

# 测试9: 创建实例（需要认证）
echo "📋 测试9: 创建实例（需要认证）"
curl -s -X POST -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/instances" | jq .
echo ""

echo "✅ API测试完成！"
