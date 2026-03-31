package spot_vm

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gitee.com/dinglide/spot-vm/internal/models"
)

// SimpleAutoManager 简化的自动管理器
type SimpleAutoManager struct {
	manager       *SpotVMManager
	region        string
	targetRegion  string // 目标Region，用于创建新实例
	terminationCh chan struct{}
	mu            sync.RWMutex
	isRunning     bool
	stopCh        chan struct{}
}

// NewSimpleAutoManager 创建简化的自动管理器
func NewSimpleAutoManager(manager *SpotVMManager, region string) *SimpleAutoManager {
	return &SimpleAutoManager{
		manager:       manager,
		region:        region,
		targetRegion:  region, // 默认目标Region与当前Region相同
		terminationCh: make(chan struct{}, 1),
		stopCh:        make(chan struct{}),
	}
}

// Start 启动自动管理器
func (sam *SimpleAutoManager) Start() {
	sam.mu.Lock()
	if sam.isRunning {
		sam.mu.Unlock()
		return
	}
	sam.isRunning = true
	sam.mu.Unlock()

	log.Println("🚀 启动Spot VM自动管理器...")

	// 启动终止检测
	go sam.monitorTermination()

	// 启动替换处理
	go sam.handleReplacement()

	log.Println("✅ Spot VM自动管理器启动成功")
}

// Stop 停止自动管理器
func (sam *SimpleAutoManager) Stop() {
	sam.mu.Lock()
	if !sam.isRunning {
		sam.mu.Unlock()
		return
	}
	sam.isRunning = false
	sam.mu.Unlock()

	close(sam.stopCh)
	log.Println("🛑 Spot VM自动管理器已停止")
}

// IsRunning 检查是否正在运行
func (sam *SimpleAutoManager) IsRunning() bool {
	sam.mu.RLock()
	defer sam.mu.RUnlock()
	return sam.isRunning
}

// monitorTermination 监控实例终止
func (sam *SimpleAutoManager) monitorTermination() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-sam.stopCh:
			return
		case <-ticker.C:
			if sam.manager.SpotVM.getInstanceTerminated() {
				log.Println("⚠️  检测到Spot实例即将被回收！")
				sam.terminationCh <- struct{}{}
			}
		}
	}
}

// handleReplacement 处理实例替换
func (sam *SimpleAutoManager) handleReplacement() {
	for {
		select {
		case <-sam.stopCh:
			return
		case <-sam.terminationCh:
			sam.createReplacementInstance()
		}
	}
}

// createReplacementInstance 创建替换实例
func (sam *SimpleAutoManager) createReplacementInstance() {
	log.Println("🔄 开始创建替换实例...")

	// 获取目标Region
	sam.mu.RLock()
	targetRegion := sam.targetRegion
	sam.mu.RUnlock()

	log.Printf("🎯 目标Region: %s", targetRegion)

	// 获取最便宜的实例
	cheapestInstance, err := sam.getCheapestInstanceInRegion(targetRegion)
	if err != nil {
		log.Printf("❌ 获取最便宜实例失败: %v", err)
		return
	}

	log.Printf("💰 选择最便宜实例: %s (¥%.4f/小时)",
		cheapestInstance.InstanceType, cheapestInstance.Price.UnitPriceDiscount)

	// 根据目标Region选择正确的SpotVMManager
	var manager *SpotVMManager
	if targetRegion == sam.region {
		manager = sam.manager
	} else {
		// 创建目标Region的SpotVMManager
		manager = NewSpotVMManager(targetRegion)
		log.Printf("🔄 使用目标Region的SpotVMManager: %s", targetRegion)
	}

	// 创建新实例
	instanceIds, err := manager.CreateSpotInstance(
		cheapestInstance.Zone,
		cheapestInstance.InstanceType,
		false, // 实际创建
	)
	if err != nil {
		log.Printf("❌ 创建替换实例失败: %v", err)
		return
	}

	if len(instanceIds) > 0 {
		log.Printf("✅ 替换实例创建成功: %s", instanceIds[0])
		log.Printf("🎉 新实例已启动，当前实例即将被回收")
	}
}

// getCheapestInstance 获取最便宜的实例（当前Region）
func (sam *SimpleAutoManager) getCheapestInstance() (*models.InstanceType, error) {
	return sam.getCheapestInstanceInRegion(sam.region)
}

// getCheapestInstanceInRegion 获取指定Region的最便宜实例
func (sam *SimpleAutoManager) getCheapestInstanceInRegion(region string) (*models.InstanceType, error) {
	// 如果目标Region与当前Region不同，需要创建新的SpotVMManager
	var manager *SpotVMManager
	if region == sam.region {
		manager = sam.manager
	} else {
		// 创建目标Region的SpotVMManager
		manager = NewSpotVMManager(region)
	}

	// 获取所有可用区
	zones, err := manager.GetAvailableZones()
	if err != nil {
		return nil, fmt.Errorf("获取可用区失败: %v", err)
	}

	// 收集所有实例信息
	var allInstances []*models.InstanceType
	for _, zone := range zones {
		instances, err := manager.GetAvailableInstanceTypes(zone.Zone)
		if err != nil {
			continue // 跳过失败的可用区
		}
		allInstances = append(allInstances, instances...)
	}

	if len(allInstances) == 0 {
		return nil, fmt.Errorf("在Region %s 中没有可用的Spot实例", region)
	}

	// 按价格排序，找到最便宜的实例
	cheapest := allInstances[0]
	for _, instance := range allInstances[1:] {
		if instance.Price.UnitPriceDiscount < cheapest.Price.UnitPriceDiscount {
			cheapest = instance
		}
	}

	return cheapest, nil
}

// SimulateTermination 模拟实例终止（用于测试）
func (sam *SimpleAutoManager) SimulateTermination() {
	log.Println("🧪 模拟Spot实例终止...")
	sam.terminationCh <- struct{}{}
}

// GetStatus 获取管理器状态
func (sam *SimpleAutoManager) GetStatus() map[string]interface{} {
	sam.mu.RLock()
	defer sam.mu.RUnlock()

	return map[string]interface{}{
		"is_running":    sam.isRunning,
		"region":        sam.region,
		"target_region": sam.targetRegion,
		"last_check":    time.Now().Format(time.RFC3339),
	}
}

// GetCurrentInstanceStatus 获取当前Spot机器状态
func (sam *SimpleAutoManager) GetCurrentInstanceStatus() map[string]interface{} {
	// 获取当前实例信息
	instanceInfo := sam.manager.SpotVM.GetInstanceInfo()

	// 检查是否即将被回收
	isTerminated := sam.manager.SpotVM.getInstanceTerminated()

	// 从map中获取值，如果不存在则使用默认值
	getString := func(key string) string {
		if val, ok := instanceInfo[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return "N/A"
	}

	return map[string]interface{}{
		"instance_id":   getString("InstanceId"),
		"instance_type": getString("InstanceType"),
		"zone":          getString("Zone"),
		"region":        sam.region,
		"status":        getString("Status"),
		"is_terminated": isTerminated,
		"private_ip":    getString("PrivateIpAddress"),
		"public_ip":     getString("PublicIpAddress"),
		"created_time":  getString("CreatedTime"),
		"expired_time":  getString("ExpiredTime"),
	}
}

// GetCurrentRegion 获取当前Region
func (sam *SimpleAutoManager) GetCurrentRegion() string {
	sam.mu.RLock()
	defer sam.mu.RUnlock()
	return sam.region
}

// GetTargetRegion 获取目标Region
func (sam *SimpleAutoManager) GetTargetRegion() string {
	sam.mu.RLock()
	defer sam.mu.RUnlock()
	return sam.targetRegion
}

// SetTargetRegion 设置目标Region
func (sam *SimpleAutoManager) SetTargetRegion(region string) {
	sam.mu.Lock()
	defer sam.mu.Unlock()
	sam.targetRegion = region
	log.Printf("🎯 目标Region已更新: %s -> %s", sam.region, region)
}

// TriggerTermination 手动触发实例回收
func (sam *SimpleAutoManager) TriggerTermination() {
	log.Println("🧪 手动触发Spot实例回收...")
	sam.terminationCh <- struct{}{}
}
