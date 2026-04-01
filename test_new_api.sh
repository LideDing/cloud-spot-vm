#!/bin/bash

echo "🚀 测试新的Spot VM管理API..."

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
go run cmd/spot-manager/main.go > server.log 2>&1 &
SERVER_PID=$!

# 等待服务器启动
sleep 5

API_KEY="your-super-secret-api-key-change-in-production"
BASE_URL="http://localhost:8080"

echo "📋 测试新的API功能..."

# 测试1: 健康检查
echo "1️⃣ 健康检查"
curl -s "$BASE_URL/api/v1/health" | jq .
echo ""

# 测试2: 获取当前Spot机器状态
echo "2️⃣ 获取当前Spot机器状态"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/current/status" | jq .
echo ""

# 测试3: 查看当前Region
echo "3️⃣ 查看当前Region"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/current/region" | jq .
echo ""

# 测试4: 修改目标Region
echo "4️⃣ 修改目标Region为 ap-beijing"
curl -s -X PUT -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"region":"ap-beijing"}' \
  "$BASE_URL/api/v1/spot-vm/target-region" | jq .
echo ""

# 测试5: 再次查看当前Region（应该没有变化）
echo "5️⃣ 再次查看当前Region（应该没有变化）"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/current/region" | jq .
echo ""

# 测试6: 获取自动管理器状态（查看目标Region是否已更新）
echo "6️⃣ 获取自动管理器状态（查看目标Region是否已更新）"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/auto-manager/status" | jq .
echo ""

# 测试7: 手动触发回收
echo "7️⃣ 手动触发Spot机器回收"
echo "⚠️  这将触发自动创建新的Spot VM实例！"
curl -s -X POST -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/trigger-termination" | jq .
echo ""

# 等待一段时间让自动管理器处理
echo "⏳ 等待自动管理器处理..."
sleep 10

# 测试8: 再次检查状态
echo "8️⃣ 检查自动管理器状态"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/spot-vm/auto-manager/status" | jq .
echo ""

# 测试9: 修改回原来的Region
echo "9️⃣ 修改目标Region回 sa-saopaulo"
curl -s -X PUT -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"region":"sa-saopaulo"}' \
  "$BASE_URL/api/v1/spot-vm/target-region" | jq .
echo ""

# 停止服务器
echo "🛑 停止服务器..."
kill $SERVER_PID

echo "✅ 新API功能测试完成！"
echo ""
echo "📝 新API功能总结:"
echo "   ✅ GET  /api/v1/spot-vm/current/status - 获取当前Spot机器状态"
echo "   ✅ GET  /api/v1/spot-vm/current/region - 查看当前Region"
echo "   ✅ PUT  /api/v1/spot-vm/target-region - 修改目标Region"
echo "   ✅ POST /api/v1/spot-vm/trigger-termination - 手动触发回收"
echo ""
echo "🎯 功能说明:"
echo "   - 可以查看当前Spot实例的详细状态"
echo "   - 可以查看当前实例所在的Region"
echo "   - 可以修改目标Region，下次创建实例时会在新Region中创建"
echo "   - 可以手动触发实例回收，测试自动替换功能"
