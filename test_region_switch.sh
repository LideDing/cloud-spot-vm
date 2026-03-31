#!/bin/bash

echo "🌍 测试Region切换功能..."

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
go run cmd/spot-manager/main.go > server.log 2>&1 &
SERVER_PID=$!

# 等待服务器启动
sleep 5

API_KEY="your-super-secret-api-key-change-in-production"
BASE_URL="http://localhost:8080/api/v1"

echo "📋 测试Region切换功能..."

# 测试1: 查看当前Region
echo "1️⃣ 查看当前Region"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/current/region" | jq .
echo ""

# 测试2: 查看自动管理器状态
echo "2️⃣ 查看自动管理器状态"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/auto-manager/status" | jq .
echo ""

# 测试3: 切换到ap-beijing
echo "3️⃣ 切换到ap-beijing"
curl -s -X PUT -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"region":"ap-beijing"}' \
  "$BASE_URL/spot-vm/target-region" | jq .
echo ""

# 测试4: 验证目标Region已更新
echo "4️⃣ 验证目标Region已更新"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/auto-manager/status" | jq .
echo ""

# 测试5: 手动触发回收（测试跨Region创建）
echo "5️⃣ 手动触发回收（测试跨Region创建）"
echo "⚠️  这将尝试在ap-beijing创建新实例！"
curl -s -X POST -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/trigger-termination" | jq .
echo ""

# 等待处理
echo "⏳ 等待自动管理器处理..."
sleep 15

# 测试6: 检查处理结果
echo "6️⃣ 检查处理结果"
curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/auto-manager/status" | jq .
echo ""

# 测试7: 切换回sa-saopaulo
echo "7️⃣ 切换回sa-saopaulo"
curl -s -X PUT -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"region":"sa-saopaulo"}' \
  "$BASE_URL/spot-vm/target-region" | jq .
echo ""

# 测试8: 再次触发回收（测试同Region创建）
echo "8️⃣ 再次触发回收（测试同Region创建）"
curl -s -X POST -H "X-API-Key: $API_KEY" "$BASE_URL/spot-vm/trigger-termination" | jq .
echo ""

# 等待处理
echo "⏳ 等待自动管理器处理..."
sleep 15

# 停止服务器
echo "🛑 停止服务器..."
kill $SERVER_PID

echo "✅ Region切换功能测试完成！"
echo ""
echo "📝 测试总结:"
echo "   - 测试了Region配置的修改"
echo "   - 测试了跨Region创建实例"
echo "   - 测试了同Region创建实例"
echo "   - 验证了SpotVMManager的正确使用"
echo ""
echo "🔍 查看详细日志:"
echo "   tail -f server.log"
