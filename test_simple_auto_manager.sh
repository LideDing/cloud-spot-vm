#!/bin/bash

echo "🤖 测试简化的Spot VM自动管理功能..."

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
go run cmd/spot-manager/main.go > server.log 2>&1 &
SERVER_PID=$!

# 等待服务器启动
sleep 5

API_KEY="your-super-secret-api-key-change-in-production"
BASE_URL="http://localhost:8080"

# 测试1: 健康检查
echo "📋 测试1: 健康检查"
curl -s "$BASE_URL/api/v1/health" | jq .
echo ""

# 测试2: 获取自动管理器状态
echo "📋 测试2: 获取自动管理器状态"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/auto-manager/status" | jq .
echo ""

# 测试3: 模拟实例终止
echo "📋 测试3: 模拟实例终止（触发自动替换）"
echo "⚠️  这将触发自动创建新的Spot VM实例！"
curl -s -X POST -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/auto-manager/simulate-termination" | jq .
echo ""

# 等待一段时间让自动管理器处理
echo "⏳ 等待自动管理器处理..."
sleep 10

# 测试4: 再次检查状态
echo "📋 测试4: 检查自动管理器状态"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/auto-manager/status" | jq .
echo ""

# 停止服务器
echo "🛑 停止服务器..."
kill $SERVER_PID

echo "✅ 简化自动管理功能测试完成！"
echo ""
echo "📝 功能说明:"
echo "   - 在Spot实例上运行此程序"
echo "   - 每10秒检查一次实例是否即将被回收"
echo "   - 检测到回收信号时，自动创建新的最便宜实例"
echo "   - 支持模拟终止进行测试"
echo ""
echo "🔧 部署建议:"
echo "   1. 在Spot实例上运行: go run cmd/simple_main.go"
echo "   2. 程序会自动启动自动管理器"
echo "   3. 检测到回收时自动创建新实例"
echo "   4. 通过API监控状态"
