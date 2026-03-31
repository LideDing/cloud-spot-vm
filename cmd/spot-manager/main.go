package main

import (
	"fmt"
	"log"

	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/routes"
	"gitee.com/dinglide/spot-vm/internal/tcc"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建TCC实例
	// ap-singapore
	// ap-hongkong
	// ap-seoul
	// ap-tokyo
	// ap-bangkok
	// ap-jakarta
	// na-siliconvalley
	// eu-frankfurt
	// na-ashburn
	// sa-saopaulo
	region := "sa-saopaulo" // 可以根据需要修改
	domain := "oitcep.com"
	certificateId := "Qi1S1ItN"
	tccClient, err := tcc.NewTCC(region, certificateId, domain)
	if err != nil {
		fmt.Printf("创建TCC客户端失败: %v\n", err)
		return
	}

	// 启动自动管理器
	tccClient.AutoManager.Start()

	// 设置路由
	router := routes.SetupRoutes(cfg, tccClient)

	// 启动服务器
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 启动Spot VM自动管理服务...\n")
	fmt.Printf("📍 监听端口: %s\n", port)
	fmt.Printf("🔑 API Key: %s\n", cfg.APIKey)
	fmt.Printf("🌐 访问地址: http://localhost:%s\n", port)
	fmt.Printf("🤖 自动管理器: 已启动\n")
	fmt.Printf("📋 API端点:\n")
	fmt.Printf("   GET  /api/v1/health - 健康检查\n")
	fmt.Printf("   POST /api/v1/auth/validate - 验证API key\n")
	fmt.Printf("   GET  /api/v1/spot-vm/current/status - 获取当前Spot机器状态\n")
	fmt.Printf("   GET  /api/v1/spot-vm/current/region - 查看当前Region\n")
	fmt.Printf("   PUT  /api/v1/spot-vm/target-region - 修改目标Region\n")
	fmt.Printf("   POST /api/v1/spot-vm/trigger-termination - 手动触发回收\n")
	fmt.Printf("   GET  /api/v1/spot-vm/auto-manager/status - 获取自动管理器状态\n")
	fmt.Printf("   POST /api/v1/spot-vm/auto-manager/start - 启动自动管理器\n")
	fmt.Printf("   POST /api/v1/spot-vm/auto-manager/stop - 停止自动管理器\n")
	fmt.Printf("   POST /api/v1/spot-vm/auto-manager/simulate-termination - 模拟实例终止\n")
	fmt.Printf("\n📝 使用示例:\n")
	fmt.Printf("   # 健康检查\n")
	fmt.Printf("   curl http://localhost:%s/api/v1/health\n", port)
	fmt.Printf("\n   # 获取当前Spot机器状态\n")
	fmt.Printf("   curl -H 'X-API-Key: %s' http://localhost:%s/api/v1/spot-vm/current/status\n", cfg.APIKey, port)
	fmt.Printf("\n   # 查看当前Region\n")
	fmt.Printf("   curl -H 'X-API-Key: %s' http://localhost:%s/api/v1/spot-vm/current/region\n", cfg.APIKey, port)
	fmt.Printf("\n   # 修改目标Region\n")
	fmt.Printf("   curl -X PUT -H 'X-API-Key: %s' -H 'Content-Type: application/json' -d '{\"region\":\"ap-beijing\"}' http://localhost:%s/api/v1/spot-vm/target-region\n", cfg.APIKey, port)
	fmt.Printf("\n   # 手动触发回收\n")
	fmt.Printf("   curl -X POST -H 'X-API-Key: %s' http://localhost:%s/api/v1/spot-vm/trigger-termination\n", cfg.APIKey, port)
	fmt.Printf("\n🎯 功能说明:\n")
	fmt.Printf("   - 每10秒检查一次实例是否即将被回收\n")
	fmt.Printf("   - 检测到回收信号时自动创建新的最便宜实例\n")
	fmt.Printf("   - 支持模拟终止进行测试\n")
	fmt.Printf("   - 提供状态查询和控制接口\n")

	log.Fatal(router.Run(":" + port))
}
