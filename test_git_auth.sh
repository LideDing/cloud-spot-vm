#!/bin/bash

echo "🔐 测试Git SSH认证..."

# 检查SSH密钥
echo "📋 检查SSH密钥..."
if [ -f ~/.ssh/id_ed25519 ]; then
    echo "✅ SSH密钥存在: ~/.ssh/id_ed25519"
else
    echo "❌ SSH密钥不存在"
    exit 1
fi

# 检查远程仓库URL
echo "📋 检查远程仓库URL..."
REMOTE_URL=$(git remote get-url origin)
echo "当前远程仓库URL: $REMOTE_URL"

if [[ $REMOTE_URL == git@* ]]; then
    echo "✅ 使用SSH协议"
else
    echo "❌ 仍在使用HTTPS协议"
    exit 1
fi

# 测试SSH连接
echo "📋 测试SSH连接..."
SSH_OUTPUT=$(ssh -T git@gitee.com 2>&1)
if echo "$SSH_OUTPUT" | grep -q "dinglide"; then
    echo "✅ SSH连接成功"   
else
    echo "❌ SSH连接失败"
    echo "SSH输出: $SSH_OUTPUT"
    exit 1
fi

# 测试Git操作
echo "📋 测试Git操作..."
git fetch --quiet
if [ $? -eq 0 ]; then
    echo "✅ Git fetch成功"
else
    echo "❌ Git fetch失败"
    exit 1
fi

echo "🎉 Git SSH认证配置成功！"
echo ""
echo "📝 使用说明:"
echo "   git pull origin master    # 拉取最新代码"
echo "   git push origin master    # 推送代码"
echo "   git fetch                 # 获取远程更新"
