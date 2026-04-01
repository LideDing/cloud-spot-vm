package handlers

import (
	"log"
	"net/http"
	"sort"
	"strconv"

	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/models"
	"gitee.com/dinglide/spot-vm/internal/tcc"
	"github.com/gin-gonic/gin"
)

// SpotVMHandler Spot VM处理器
type SpotVMHandler struct {
	cfg *config.Config
	tcc *tcc.TCC
}

// NewSpotVMHandler 创建Spot VM处理器
func NewSpotVMHandler(cfg *config.Config, tcc *tcc.TCC) *SpotVMHandler {
	return &SpotVMHandler{
		cfg: cfg,
		tcc: tcc,
	}
}

// CreateCheapestSpotVMRequest 创建最便宜Spot VM请求
type CreateCheapestSpotVMRequest struct {
	DryRun bool `json:"dry_run"` // 是否只预检，不实际创建
}

// CreateCheapestSpotVMResponse 创建最便宜Spot VM响应
type CreateCheapestSpotVMResponse struct {
	Status       string               `json:"status"`
	Message      string               `json:"message"`
	InstanceInfo *models.InstanceType `json:"instance_info,omitempty"`
	InstanceIds  []string             `json:"instance_ids,omitempty"`
	DryRun       bool                 `json:"dry_run"`
}

// GetCheapestInstancesResponse 获取最便宜实例响应
type GetCheapestInstancesResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Instances []*models.InstanceType `json:"instances"`
	Total     int                    `json:"total"`
}

// CreateCheapestSpotVM 创建最便宜的Spot VM
func (h *SpotVMHandler) CreateCheapestSpotVM(c *gin.Context) {
	var req CreateCheapestSpotVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"error_code": "INVALID_REQUEST",
		})
		return
	}

	// 获取所有可用区
	zones, err := h.tcc.SpotVMManager.GetAvailableZones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get available zones: " + err.Error(),
			"error_code": "ZONE_QUERY_FAILED",
		})
		return
	}

	// 收集所有实例信息
	allInstances := []*models.InstanceType{}
	for _, zone := range zones {
		instances, err := h.tcc.SpotVMManager.GetAvailableInstanceTypes(zone.Zone)
		if err != nil {
			continue // 跳过失败的可用区
		}
		allInstances = append(allInstances, instances...)
	}

	if len(allInstances) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "No available spot instances found",
			"error_code": "NO_INSTANCES_AVAILABLE",
		})
		return
	}

	// 按价格排序，找到最便宜的实例
	sort.Slice(allInstances, func(i, j int) bool {
		return allInstances[i].Price.UnitPriceDiscount < allInstances[j].Price.UnitPriceDiscount
	})

	cheapestInstance := allInstances[0]

	// 创建实例
	instanceIds, err := h.tcc.SpotVMManager.CreateSpotInstance(
		cheapestInstance.Zone,
		cheapestInstance.InstanceType,
		req.DryRun,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create spot instance: " + err.Error(),
			"error_code": "CREATE_INSTANCE_FAILED",
		})
		return
	}

	response := CreateCheapestSpotVMResponse{
		Status:       "success",
		Message:      "Spot VM created successfully",
		InstanceInfo: cheapestInstance,
		InstanceIds:  instanceIds,
		DryRun:       req.DryRun,
	}

	if req.DryRun {
		response.Message = "Dry run completed successfully"
	}

	c.JSON(http.StatusOK, response)
}

// GetCheapestInstances 获取最便宜的实例列表
func (h *SpotVMHandler) GetCheapestInstances(c *gin.Context) {
	// T014: 支持 limit 查询参数，默认返回前10个
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 获取所有可用区
	zones, err := h.tcc.SpotVMManager.GetAvailableZones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get available zones: " + err.Error(),
		})
		return
	}

	// 收集所有实例信息
	allInstances := []*models.InstanceType{}
	for _, zone := range zones {
		instances, err := h.tcc.SpotVMManager.GetAvailableInstanceTypes(zone.Zone)
		if err != nil {
			continue // 跳过失败的可用区
		}
		allInstances = append(allInstances, instances...)
	}

	if len(allInstances) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No available spot instances found",
		})
		return
	}

	// 按价格排序
	sort.Slice(allInstances, func(i, j int) bool {
		return allInstances[i].Price.UnitPriceDiscount < allInstances[j].Price.UnitPriceDiscount
	})

	// 返回前 limit 个最便宜的实例
	topInstances := allInstances
	if len(allInstances) > limit {
		topInstances = allInstances[:limit]
	}

	response := GetCheapestInstancesResponse{
		Status:    "success",
		Message:   "Retrieved cheapest instances successfully",
		Instances: topInstances,
		Total:     len(allInstances),
	}

	c.JSON(http.StatusOK, response)
}

// GetAvailableZones 获取可用区列表
func (h *SpotVMHandler) GetAvailableZones(c *gin.Context) {
	zones, err := h.tcc.SpotVMManager.GetAvailableZones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get available zones: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Retrieved available zones successfully",
		"zones":   zones,
		"total":   len(zones),
	})
}

// GetInstanceTypes 获取指定可用区的实例类型
func (h *SpotVMHandler) GetInstanceTypes(c *gin.Context) {
	zone := c.Query("zone")
	if zone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "zone parameter is required",
		})
		return
	}

	instances, err := h.tcc.SpotVMManager.GetAvailableInstanceTypes(zone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get instance types: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Retrieved instance types successfully",
		"instances": instances,
		"zone":      zone,
		"total":     len(instances),
	})
}

// DeleteInstance 删除实例
func (h *SpotVMHandler) DeleteInstance(c *gin.Context) {
	instanceID := c.Param("id")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "instance ID is required",
		})
		return
	}

	err := h.tcc.SpotVMManager.DeleteInstance(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete instance: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Instance deleted successfully",
		"instance_id": instanceID,
	})
}

// GetAutoManagerStatus 获取自动管理器状态
func (h *SpotVMHandler) GetAutoManagerStatus(c *gin.Context) {
	status := h.tcc.AutoManager.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Retrieved auto manager status successfully",
		"data":    status,
	})
}

// StartAutoManager 启动自动管理器
func (h *SpotVMHandler) StartAutoManager(c *gin.Context) {
	if h.tcc.AutoManager.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Auto manager is already running",
		})
		return
	}

	h.tcc.AutoManager.Start()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Auto manager started successfully",
	})
}

// StopAutoManager 停止自动管理器
func (h *SpotVMHandler) StopAutoManager(c *gin.Context) {
	if !h.tcc.AutoManager.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Auto manager is not running",
		})
		return
	}

	h.tcc.AutoManager.Stop()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Auto manager stopped successfully",
	})
}

// SimulateTermination 模拟实例终止（用于测试）
func (h *SpotVMHandler) SimulateTermination(c *gin.Context) {
	if !h.tcc.AutoManager.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Auto manager is not running. Please start it first.",
		})
		return
	}

	h.tcc.AutoManager.SimulateTermination()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Termination simulation triggered successfully",
	})
}

// GetCurrentInstanceStatus 获取当前Spot机器状态
func (h *SpotVMHandler) GetCurrentInstanceStatus(c *gin.Context) {
	status := h.tcc.AutoManager.GetCurrentInstanceStatus()

	// T016: 检查是否获取到有效数据
	instanceId, _ := status["instance_id"].(string)
	if instanceId == "" || instanceId == "N/A" {
		log.Println("⚠️  获取当前实例状态失败: 可能不在腾讯云实例上运行，或 metadata 服务不可达")
		c.JSON(http.StatusOK, gin.H{
			"status":  "warning",
			"message": "无法获取实例信息，可能不在腾讯云实例上运行，或 metadata 服务不可达",
			"data":    status,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Retrieved current instance status successfully",
		"data":    status,
	})
}

// GetCurrentRegion 查看当前spot机器的Region
func (h *SpotVMHandler) GetCurrentRegion(c *gin.Context) {
	region := h.tcc.AutoManager.GetCurrentRegion()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Retrieved current region successfully",
		"data": gin.H{
			"current_region": region,
		},
	})
}

// SetTargetRegion 修改Region配置
func (h *SpotVMHandler) SetTargetRegion(c *gin.Context) {
	var req struct {
		Region string `json:"region" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request format: " + err.Error(),
			"error_code": "INVALID_REQUEST",
		})
		return
	}

	// 验证Region格式
	if req.Region == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Region cannot be empty",
			"error_code": "EMPTY_REGION",
		})
		return
	}

	// T021: 记录切换前的状态
	oldRegion := h.tcc.AutoManager.GetTargetRegion()

	// SetTargetRegion 会验证 Region 有效性
	if err := h.tcc.AutoManager.SetTargetRegion(req.Region); err != nil {
		// 获取可用 Region 列表，帮助用户选择
		var availableRegions []string
		if regions, regErr := h.tcc.SpotVMManager.GetAvailableRegions(); regErr == nil {
			for _, r := range regions {
				availableRegions = append(availableRegions, r.Region)
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "无效的Region: " + req.Region,
			"error_code":        "INVALID_REGION",
			"available_regions": availableRegions,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Target region updated successfully",
		"data": gin.H{
			"current_region": h.tcc.AutoManager.GetCurrentRegion(),
			"old_target":     oldRegion,
			"new_target":     req.Region,
		},
	})
}

// TriggerTermination 手动触发spot机器回收
func (h *SpotVMHandler) TriggerTermination(c *gin.Context) {
	if !h.tcc.AutoManager.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Auto manager is not running. Please start it first.",
		})
		return
	}

	h.tcc.AutoManager.TriggerTermination()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Manual termination triggered successfully",
	})
}

// GetAvailableRegions 获取所有可用的腾讯云 Region 列表
func (h *SpotVMHandler) GetAvailableRegions(c *gin.Context) {
	regions, err := h.tcc.SpotVMManager.GetAvailableRegions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get available regions: " + err.Error(),
			"error_code": "REGION_QUERY_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"message":        "Retrieved available regions successfully",
		"regions":        regions,
		"total":          len(regions),
		"current_region": h.tcc.AutoManager.GetCurrentRegion(),
		"target_region":  h.tcc.AutoManager.GetTargetRegion(),
	})
}
