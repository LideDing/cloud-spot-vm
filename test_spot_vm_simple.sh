#!/bin/bash

echo "🎯 简单测试Spot VM API..."

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
go run cmd/main.go > server.log 2>&1 &
SERVER_PID=$!

# 等待服务器启动
sleep 5

API_KEY="your-super-secret-api-key-change-in-production"
BASE_URL="http://localhost:8080"

# 测试健康检查
echo "📋 测试1: 健康检查"
curl -s "$BASE_URL/api/v1/health" | jq .
echo ""

# 测试获取可用区
echo "📋 测试2: 获取可用区"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/zones" | jq .
echo ""

# 测试直接创建Spot VM
echo "📋 测试3: 直接创建Spot VM"
echo "⚠️  注意：这将实际创建Spot VM实例，可能会产生费用！"
curl -s -X POST -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"dry_run": false}' "$BASE_URL/api/v1/spot-vm/create-cheapest" | jq .
echo ""

# 停止服务器
echo "🛑 停止服务器..."
kill $SERVER_PID

echo "✅ 测试完成！"
