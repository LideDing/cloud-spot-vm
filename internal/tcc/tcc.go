package tcc

import (
	"log"

	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/migration"
	"gitee.com/dinglide/spot-vm/internal/tcc/network"
	"gitee.com/dinglide/spot-vm/internal/tcc/spot_vm"
)

// TCC 管理器
type TCC struct {
	Region        string
	SpotVMManager *spot_vm.SpotVMManager
	AutoManager   *spot_vm.SimpleAutoManager
	SSLManager    *network.SSLManager
	DNSManager    *network.DNSManager
	Migrator      *migration.Migrator
}

// NewTCC 创建 TCC 实例
func NewTCC(region, certificateId, domain string, cfg *config.Config) (*TCC, error) {
	vmCfg := spot_vm.VMConfig{
		ImageId:          cfg.ImageId,
		InstancePassword: cfg.InstancePassword,
		DiskType:         cfg.DiskType,
		DiskSize:         cfg.DiskSize,
		Bandwidth:        cfg.Bandwidth,
	}
	spotVMManager := spot_vm.NewSpotVMManager(region, vmCfg)
	sslManager := network.NewSSLManager(region, certificateId)
	dnsManager := network.NewDNSManager(region, domain)
	migrator := migration.NewMigrator(cfg)
	autoManager := spot_vm.NewSimpleAutoManager(spotVMManager, region, cfg, migrator)

	// 注册实例创建成功后的网络配置回调（DNS + SSL + Nginx）
	autoManager.OnInstanceCreated = func(instanceId string, publicIP string) {
		log.Printf("🌐 开始执行网络配置: instanceId=%s, publicIP=%s", instanceId, publicIP)

		// 1. 更新 DNS 记录
		if publicIP != "" && domain != "" {
			if err := dnsManager.CreateDNSHost(publicIP); err != nil {
				log.Printf("❌ DNS记录更新失败: %v", err)
			} else {
				log.Printf("✅ DNS记录更新成功: www.%s -> %s", domain, publicIP)
			}
		}

		// 2. 获取 SSL 证书并远程部署 Nginx
		var sslCertContent, sslKeyContent string
		if certificateId != "" {
			certResp, err := sslManager.GetSSLCertificate()
			if err != nil {
				log.Printf("❌ 获取SSL证书失败: %v", err)
			} else {
				sslCertContent = certResp.Response.CertificatePublicKey
				sslKeyContent = certResp.Response.CertificatePrivateKey
			}
		}

		// 3. 远程部署 Nginx（通过 SSH）
		if publicIP != "" {
			if err := migrator.DeployNginxRemotely(publicIP, sslCertContent, sslKeyContent); err != nil {
				log.Printf("❌ 远程部署 Nginx 失败: %v", err)
			} else {
				log.Printf("✅ Nginx 远程部署成功")
			}
		}

		log.Printf("🎉 网络配置完成: instanceId=%s, publicIP=%s", instanceId, publicIP)
	}

	return &TCC{
		Region:        region,
		SpotVMManager: spotVMManager,
		AutoManager:   autoManager,
		SSLManager:    sslManager,
		DNSManager:    dnsManager,
		Migrator:      migrator,
	}, nil
}
