#!/bin/bash

echo "🧪 测试API key认证系统..."

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
./spot-vm > server.log 2>&1 &
SERVER_PID=$!

# 等待服务器启动
sleep 3

# 测试健康检查
echo "📋 测试1: 健康检查"
curl -s http://localhost:8080/api/v1/health | jq .

# 测试API key验证
echo -e "\n📋 测试2: API key验证"
curl -s -X POST http://localhost:8080/api/v1/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"api_key": "your-super-secret-api-key-change-in-production"}' | jq .

# 测试受保护的端点
echo -e "\n📋 测试3: 受保护的端点"
curl -s -H "X-API-Key: your-super-secret-api-key-change-in-production" \
  http://localhost:8080/api/v1/protected | jq .

# 测试无认证访问
echo -e "\n📋 测试4: 无认证访问（应该失败）"
curl -s http://localhost:8080/api/v1/protected | jq .

# 停止服务器
echo -e "\n🛑 停止服务器..."
kill $SERVER_PID

echo "✅ 测试完成！"
