package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/routes"
	"gitee.com/dinglide/spot-vm/internal/tcc"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建TCC实例 — 所有参数从配置读取，不硬编码
	region := cfg.Region
	if region == "" {
		log.Fatal("❌ 未配置 REGION，请在 .env 文件中设置 REGION 环境变量")
	}
	domain := cfg.Domain
	if domain == "" {
		log.Println("⚠️  未配置 DOMAIN，DNS 更新功能将不可用")
	}
	certificateId := cfg.CertificateId
	if certificateId == "" {
		log.Println("⚠️  未配置 CERTIFICATE_ID，SSL 证书功能将不可用")
	}

	tccClient, err := tcc.NewTCC(region, certificateId, domain, cfg)
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
	fmt.Printf("🌐 Region: %s\n", region)
	fmt.Printf("🌐 Domain: %s\n", domain)
	fmt.Printf("🔐 CertificateId: %s\n", certificateId)
	fmt.Printf("🔧 SSH端口: %d, 超时: %ds, 等待超时: %ds\n", cfg.SSHPort, cfg.SSHTimeout, cfg.SSHWaitTimeout)
	fmt.Printf("🔄 迁移最大重试: %d, 远程路径: %s\n", cfg.MigrationMaxRetries, cfg.RemoteBinaryPath)
	fmt.Printf("🤖 自动管理器: 已启动\n")
	fmt.Printf("📋 API端点:\n")
	fmt.Printf("   GET  /api/v1/health - 健康检查\n")
	fmt.Printf("   POST /api/v1/auth/validate - 验证API key\n")
	fmt.Printf("   GET  /api/v1/spot-vm/current/status - 获取当前Spot机器状态\n")
	fmt.Printf("   GET  /api/v1/spot-vm/current/region - 查看当前Region\n")
	fmt.Printf("   PUT  /api/v1/spot-vm/target-region - 修改目标Region\n")
	fmt.Printf("   GET  /api/v1/spot-vm/regions - 获取可用Region列表\n")
	fmt.Printf("   POST /api/v1/spot-vm/trigger-termination - 手动触发回收\n")
	fmt.Printf("   GET  /api/v1/spot-vm/auto-manager/status - 获取自动管理器状态\n")
	fmt.Printf("   POST /api/v1/spot-vm/auto-manager/start - 启动自动管理器\n")
	fmt.Printf("   POST /api/v1/spot-vm/auto-manager/stop - 停止自动管理器\n")
	fmt.Printf("   POST /api/v1/spot-vm/auto-manager/simulate-termination - 模拟实例终止\n")

	// 优雅关闭（graceful shutdown）
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 在 goroutine 中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动HTTP服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 收到关闭信号，开始优雅关闭...")

	// 先停止 AutoManager
	log.Println("🔄 停止自动管理器...")
	tccClient.AutoManager.Stop()

	// 关闭 HTTP 服务器，等待最多 10 秒
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ HTTP服务器关闭失败: %v", err)
	}

	log.Println("✅ 服务已安全关闭")
}
