package spot_vm

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/migration"
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
	replacing     atomic.Bool // 去重标志：防止同时创建多个替换实例

	// 配置
	cfg *config.Config

	// 迁移引擎
	migrator *migration.Migrator

	// 当前活跃的迁移任务
	currentMigration *models.MigrationTask

	// 实例创建成功后的回调函数（用于网络配置等后续操作）
	OnInstanceCreated func(instanceId string, publicIP string)
}

// NewSimpleAutoManager 创建简化的自动管理器
func NewSimpleAutoManager(manager *SpotVMManager, region string, cfg *config.Config, migrator *migration.Migrator) *SimpleAutoManager {
	return &SimpleAutoManager{
		manager:       manager,
		region:        region,
		targetRegion:  region, // 默认目标Region与当前Region相同
		terminationCh: make(chan struct{}, 1),
		stopCh:        make(chan struct{}),
		cfg:           cfg,
		migrator:      migrator,
	}
}

// Start 启动自动管理器
func (sam *SimpleAutoManager) Start() {
	sam.mu.Lock()
	if sam.isRunning {
		sam.mu.Unlock()
		log.Println("⚠️  自动管理器已在运行中，跳过启动")
		return
	}
	sam.isRunning = true
	sam.mu.Unlock()

	log.Printf("🚀 启动Spot VM自动管理器 [region=%s, targetRegion=%s]", sam.region, sam.targetRegion)

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
		log.Println("⚠️  自动管理器未在运行，跳过停止")
		return
	}
	sam.isRunning = false
	sam.mu.Unlock()

	close(sam.stopCh)
	log.Printf("🛑 Spot VM自动管理器已停止 [region=%s, targetRegion=%s]", sam.region, sam.targetRegion)
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
				// T010: 非阻塞 select 模式，防止通道满时阻塞 goroutine
				select {
				case sam.terminationCh <- struct{}{}:
				default:
					log.Println("⚠️  终止信号通道已满，跳过本次信号")
				}
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

// createReplacementInstance 创建替换实例（带去重和重试机制）
func (sam *SimpleAutoManager) createReplacementInstance() {
	// T013: 去重机制，防止同时创建多个替换实例
	if !sam.replacing.CompareAndSwap(false, true) {
		log.Println("⚠️  已有替换任务在进行中，跳过本次请求")
		return
	}
	defer sam.replacing.Store(false)

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

	log.Printf("💰 选择最便宜实例: %s (CPU=%d, Mem=%dMB, ¥%.4f/小时) [zone=%s]",
		cheapestInstance.InstanceType, cheapestInstance.Cpu, cheapestInstance.Memory,
		cheapestInstance.Price.UnitPriceDiscount, cheapestInstance.Zone)

	// 根据目标Region选择正确的SpotVMManager
	var manager *SpotVMManager
	if targetRegion == sam.region {
		manager = sam.manager
	} else {
		manager = NewSpotVMManager(targetRegion, sam.manager.config)
		log.Printf("🔄 使用目标Region的SpotVMManager: %s", targetRegion)
	}

	// T011: 指数退避重试机制（最多3次，间隔5s/15s/45s）
	retryDelays := []time.Duration{5 * time.Second, 15 * time.Second, 45 * time.Second}
	var instanceIds []string

	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		instanceIds, err = manager.CreateSpotInstance(
			cheapestInstance.Zone,
			cheapestInstance.InstanceType,
			false,
		)
		if err == nil {
			break
		}

		if attempt < len(retryDelays) {
			log.Printf("❌ 创建替换实例失败 (第%d次尝试): %v，%v后重试...", attempt+1, err, retryDelays[attempt])
			time.Sleep(retryDelays[attempt])
		} else {
			log.Printf("❌ 创建替换实例失败，已耗尽所有重试次数: %v", err)
			return
		}
	}

	if len(instanceIds) > 0 {
		instanceId := instanceIds[0]
		log.Printf("✅ 替换实例创建成功: %s", instanceId)

		// 创建迁移任务
		migrationTask := &models.MigrationTask{
			TargetInstanceId: instanceId,
			TargetRegion:     targetRegion,
			TargetZone:       cheapestInstance.Zone,
			Status:           models.MigrationStatusPending,
			RetryCount:       0,
			MaxRetries:       sam.cfg.MigrationMaxRetries,
			StartTime:        time.Now(),
		}

		sam.mu.Lock()
		sam.currentMigration = migrationTask
		sam.mu.Unlock()

		// 在后台执行完整的迁移流程
		go sam.executeMigrationFlow(manager, migrationTask)
	}
}

// executeMigrationFlow 执行完整的迁移流程：等待实例就绪 → 程序迁移 → DNS 更新 → Nginx 部署
func (sam *SimpleAutoManager) executeMigrationFlow(manager *SpotVMManager, task *models.MigrationTask) {
	log.Printf("🚀 开始执行完整迁移流程: instanceId=%s, region=%s, zone=%s",
		task.TargetInstanceId, task.TargetRegion, task.TargetZone)

	// 1. 等待实例进入 RUNNING 状态并获取公网 IP
	sam.updateMigrationStatus(task, models.MigrationStatusWaitingSSH, "")
	detail, err := manager.WaitForInstanceRunning(task.TargetInstanceId, sam.cfg.SSHWaitTimeout)
	if err != nil {
		sam.updateMigrationStatus(task, models.MigrationStatusFailed, fmt.Sprintf("等待实例就绪失败: %v", err))
		log.Printf("❌ 等待实例就绪失败: %v", err)
		return
	}

	task.TargetIP = detail.PublicIp
	log.Printf("✅ 实例已就绪: instanceId=%s, publicIP=%s", task.TargetInstanceId, task.TargetIP)

	// 2. 执行程序迁移（带重试）
	sam.updateMigrationStatus(task, models.MigrationStatusTransferring, "")
	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		err = sam.migrator.Migrate(task.TargetIP)
		if err == nil {
			break
		}
		task.RetryCount = attempt + 1
		if attempt < task.MaxRetries {
			log.Printf("❌ 迁移失败 (第%d次尝试): %v，重试中...", attempt+1, err)
			time.Sleep(10 * time.Second)
		} else {
			sam.updateMigrationStatus(task, models.MigrationStatusFailed, fmt.Sprintf("迁移失败，已耗尽所有重试: %v", err))
			log.Printf("❌ 迁移失败，已耗尽所有重试次数: %v", err)
			return
		}
	}

	log.Printf("✅ 程序迁移成功: instanceId=%s, publicIP=%s", task.TargetInstanceId, task.TargetIP)

	// 3. 触发实例创建成功回调（DNS + SSL + Nginx）
	if sam.OnInstanceCreated != nil {
		log.Printf("🌐 触发网络配置回调: instanceId=%s, publicIP=%s", task.TargetInstanceId, task.TargetIP)
		sam.OnInstanceCreated(task.TargetInstanceId, task.TargetIP)
	}

	sam.updateMigrationStatus(task, models.MigrationStatusCompleted, "")
	log.Printf("🎉 完整迁移流程完成: instanceId=%s, publicIP=%s, 耗时=%v",
		task.TargetInstanceId, task.TargetIP, time.Since(task.StartTime))
}

// updateMigrationStatus 更新迁移任务状态
func (sam *SimpleAutoManager) updateMigrationStatus(task *models.MigrationTask, status models.MigrationStatus, errMsg string) {
	sam.mu.Lock()
	defer sam.mu.Unlock()
	task.Status = status
	task.Error = errMsg
	log.Printf("📊 迁移状态更新: instanceId=%s, status=%s", task.TargetInstanceId, status)
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
		manager = NewSpotVMManager(region, sam.manager.config)
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
			log.Printf("⚠️  查询可用区 %s 实例类型失败: %v，跳过", zone.Zone, err)
			continue // 跳过失败的可用区
		}
		// T012: 最低配置过滤，跳过 CPU < 1 或 Memory < 1024MB 的实例类型
		for _, inst := range instances {
			if inst.Cpu >= 1 && inst.Memory >= 1024 {
				allInstances = append(allInstances, inst)
			}
		}
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

	status := map[string]interface{}{
		"is_running":       sam.isRunning,
		"region":           sam.region,
		"target_region":    sam.targetRegion,
		"last_check":       time.Now().Format(time.RFC3339),
		"migration_status": nil,
	}

	// 包含当前迁移任务状态
	if sam.currentMigration != nil {
		status["migration_status"] = sam.currentMigration
	}

	return status
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

// SetTargetRegion 设置目标Region（带有效性验证）
func (sam *SimpleAutoManager) SetTargetRegion(region string) error {
	// T019: 验证 Region 有效性
	valid, err := sam.manager.ValidateRegion(region)
	if err != nil {
		log.Printf("⚠️  验证Region有效性失败: %v，仍然设置目标Region", err)
	} else if !valid {
		return fmt.Errorf("无效的Region: %s", region)
	}

	sam.mu.Lock()
	defer sam.mu.Unlock()
	oldRegion := sam.targetRegion
	sam.targetRegion = region
	log.Printf("🎯 目标Region已更新: %s -> %s", oldRegion, region)
	return nil
}

// TriggerTermination 手动触发实例回收
func (sam *SimpleAutoManager) TriggerTermination() {
	log.Println("🧪 手动触发Spot实例回收...")
	sam.terminationCh <- struct{}{}
}
