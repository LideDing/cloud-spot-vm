package tcc

import (
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
}

// NewTCC 创建 TCC 实例
func NewTCC(region, certificateId, domain string) (*TCC, error) {
	spotVMManager := spot_vm.NewSpotVMManager(region)
	autoManager := spot_vm.NewSimpleAutoManager(spotVMManager, region)

	return &TCC{
		Region:        region,
		SpotVMManager: spotVMManager,
		AutoManager:   autoManager,
		SSLManager:    network.NewSSLManager(region, certificateId),
		DNSManager:    network.NewDNSManager(region, domain),
	}, nil
}
