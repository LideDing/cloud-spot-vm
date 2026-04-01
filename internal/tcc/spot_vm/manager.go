// Package spot_vm 管理腾讯云Spot VM
package spot_vm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gitee.com/dinglide/spot-vm/internal/models"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const (
	metadataUrl = "http://metadata.tencentyun.com/latest/meta-data/"
)

// VMConfig 实例创建配置
type VMConfig struct {
	ImageId          string
	InstancePassword string
	DiskType         string
	DiskSize         int
	Bandwidth        int
}

// SpotVMManager Spot VM管理器
type SpotVMManager struct {
	cvmClient  *common.Client // 腾讯云CVM客户端
	Terminated chan struct{}  // 实例终止信号
	SpotVM     *SpotVM        // Spot VM实例
	config     VMConfig       // 实例创建配置
}

// NewSpotVMManager 创建Spot VM管理器
func NewSpotVMManager(region string, vmCfg VMConfig) *SpotVMManager {
	credential := common.NewCredential(os.Getenv("TENCENTCLOUD_SECRET_ID"), os.Getenv("TENCENTCLOUD_SECRET_KEY"))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	cvmClient := common.NewCommonClient(credential, region, cpf).WithLogger(log.Default())
	return &SpotVMManager{
		cvmClient:  cvmClient,
		Terminated: make(chan struct{}),
		SpotVM:     NewSpotVM(),
		config:     vmCfg,
	}
}

// GetAvailableInstanceTypes 获取指定可用区下所有可售卖的实例类型
func (t *SpotVMManager) GetAvailableInstanceTypes(zone string) ([]*models.InstanceType, error) {
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "DescribeUserAvailableInstanceTypes")
	params := map[string]any{
		"Filters": []map[string]any{
			{"Name": "zone", "Values": []string{zone}},
			{"Name": "instance-charge-type", "Values": []string{"SPOTPAID"}},
		},
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return nil, err
	}
	response := tchttp.NewCommonResponse()
	err = t.cvmClient.Send(request, response)
	if err != nil {
		return nil, err
	}
	var instanceResponse models.GetInstanceResponse
	err = json.Unmarshal(response.GetBody(), &instanceResponse)
	if err != nil {
		return nil, err
	}

	// 过滤出可售卖的实例类型
	availableInstances := []*models.InstanceType{}
	for _, instance := range instanceResponse.Response.InstanceTypeQuotaSet {
		if instance.Status == "SELL" {
			// 设置可用区信息
			instance.Zone = zone
			availableInstances = append(availableInstances, &instance)
		}
	}
	return availableInstances, nil
}

// CreateSpotInstance 创建Spot实例
func (t *SpotVMManager) CreateSpotInstance(zone string, instanceType string, dryRun bool) ([]string, error) {
	log.Printf("📦 创建Spot实例: zone=%s, type=%s, image=%s, disk=%s/%dGB, bandwidth=%dMbps, dryRun=%v",
		zone, instanceType, t.config.ImageId, t.config.DiskType, t.config.DiskSize, t.config.Bandwidth, dryRun)
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "RunInstances")
	params := map[string]any{
		"InstanceChargeType": "SPOTPAID",
		"Placement":          map[string]any{"Zone": zone},
		"InstanceType":       instanceType,
		"ImageId":            t.config.ImageId,
		"SystemDisk": map[string]any{
			"DiskType": t.config.DiskType,
			"DiskSize": t.config.DiskSize,
		},
		// "VirtualPrivateCloud": map[string]any{
		// 	"VpcId":    "vpc-enajr3z3",
		// 	"SubnetId": "subnet-g7te823k",
		// },
		"InternetAccessible": map[string]any{
			"InternetChargeType":      "TRAFFIC_POSTPAID_BY_HOUR",
			"InternetMaxBandwidthOut": t.config.Bandwidth,
			"PublicIpAssigned":        true,
		},
		"InstanceCount": 1,
		"LoginSettings": map[string]any{"Password": t.config.InstancePassword},
		"EnhancedService": map[string]any{
			"SecurityService":   map[string]any{"Enabled": true},
			"MonitorService":    map[string]any{"Enabled": true},
			"AutomationService": map[string]any{"Enabled": true},
		},
		"TagSpecification": []map[string]any{
			{"ResourceType": "instance", "Tags": []map[string]any{
				{"Key": "owner", "Value": "lideding"},
				{"Key": "env", "Value": "lab"},
				{"Key": "projid", "Value": "mcnp"},
				{"Key": "service", "Value": "infra"},
				{"Key": "sod", "Value": "infra-network"},
			}}},
		"DryRun": dryRun, // 是否只预检此次请求
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("序列化创建实例参数失败 (zone=%s, type=%s): %w", zone, instanceType, err)
	}
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return nil, fmt.Errorf("设置创建实例请求参数失败 (zone=%s, type=%s): %w", zone, instanceType, err)
	}
	response := tchttp.NewCommonResponse()
	err = t.cvmClient.Send(request, response)
	if err != nil {
		return nil, fmt.Errorf("调用腾讯云RunInstances API失败 (zone=%s, type=%s): %w", zone, instanceType, err)
	}
	if dryRun {
		log.Printf("✅ DryRun成功 (zone=%s, type=%s): %s", zone, instanceType, string(response.GetBody()))
		return nil, nil
	}

	var createInstanceResponse models.CreateInstanceResponse
	err = json.Unmarshal(response.GetBody(), &createInstanceResponse)
	if err != nil {
		return nil, fmt.Errorf("解析创建实例响应失败 (zone=%s, type=%s): %w", zone, instanceType, err)
	}

	log.Printf("✅ Spot实例创建成功: zone=%s, type=%s, instanceIds=%v", zone, instanceType, createInstanceResponse.Response.InstanceIdSet)
	return createInstanceResponse.Response.InstanceIdSet, nil
}

// DeleteInstance 删除实例
func (t *SpotVMManager) DeleteInstance(instanceId string) error {
	params := map[string]any{
		"InstanceIds":    []string{instanceId},
		"ReleaseAddress": true,
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return err
	}
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "TerminateInstances")
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return err
	}
	response := tchttp.NewCommonResponse()
	err = t.cvmClient.Send(request, response)
	if err != nil {
		return err
	}
	return err
}

// GetAvailableZones 获取指定 Region 下所有可用区
func (t *SpotVMManager) GetAvailableZones() ([]*models.Zone, error) {
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "DescribeZones")
	response := tchttp.NewCommonResponse()
	err := t.cvmClient.Send(request, response)
	if err != nil {
		log.Printf("❌ 获取可用区失败: %v", err)
		return nil, fmt.Errorf("获取可用区失败: %w", err)
	}
	var availableZonesResponse models.ZoneResponse
	err = json.Unmarshal(response.GetBody(), &availableZonesResponse)
	if err != nil {
		log.Printf("❌ 解析可用区响应失败: %v", err)
		return nil, fmt.Errorf("解析可用区响应失败: %w", err)
	}
	availableZones := []*models.Zone{}
	for _, zone := range availableZonesResponse.Response.ZoneSet {
		if zone.ZoneState == "AVAILABLE" {
			availableZones = append(availableZones, zone)
		}
	}
	return availableZones, nil
}

// NewSpotVMManagerForRegion 为指定 Region 创建独立的 SpotVMManager 实例
func NewSpotVMManagerForRegion(region string, vmCfg VMConfig) *SpotVMManager {
	return NewSpotVMManager(region, vmCfg)
}

// RegionInfo Region 信息
type RegionInfo struct {
	Region      string `json:"Region"`
	RegionName  string `json:"RegionName"`
	RegionState string `json:"RegionState"`
}

// RegionResponse DescribeRegions API 响应
type RegionResponse struct {
	Response struct {
		RegionSet []*RegionInfo `json:"RegionSet"`
		RequestId string        `json:"RequestId"`
	} `json:"Response"`
}

// GetAvailableRegions 获取所有可用的腾讯云 Region 列表
func (t *SpotVMManager) GetAvailableRegions() ([]*RegionInfo, error) {
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "DescribeRegions")
	response := tchttp.NewCommonResponse()
	err := t.cvmClient.Send(request, response)
	if err != nil {
		return nil, fmt.Errorf("获取Region列表失败: %w", err)
	}
	var regionResponse RegionResponse
	err = json.Unmarshal(response.GetBody(), &regionResponse)
	if err != nil {
		return nil, fmt.Errorf("解析Region响应失败: %w", err)
	}
	// 仅返回可用的 Region
	var availableRegions []*RegionInfo
	for _, r := range regionResponse.Response.RegionSet {
		if r.RegionState == "AVAILABLE" {
			availableRegions = append(availableRegions, r)
		}
	}
	return availableRegions, nil
}

// ValidateRegion 验证 Region ID 是否有效
func (t *SpotVMManager) ValidateRegion(regionId string) (bool, error) {
	regions, err := t.GetAvailableRegions()
	if err != nil {
		return false, err
	}
	for _, r := range regions {
		if r.Region == regionId {
			return true, nil
		}
	}
	return false, nil
}

// InstanceDetail 实例详情（DescribeInstances 返回）
type InstanceDetail struct {
	InstanceId   string `json:"InstanceId"`
	InstanceType string `json:"InstanceType"`
	Zone         string `json:"Zone"`
	PublicIp     string `json:"PublicIpAddress"`
	PrivateIp    string `json:"PrivateIpAddress"`
	State        string `json:"InstanceState"`
}

// DescribeInstancesResponse DescribeInstances API 响应
type DescribeInstancesResponse struct {
	Response struct {
		InstanceSet []struct {
			InstanceId    string `json:"InstanceId"`
			InstanceType  string `json:"InstanceType"`
			InstanceState string `json:"InstanceState"`
			Placement     struct {
				Zone string `json:"Zone"`
			} `json:"Placement"`
			PublicIpAddressSet []string `json:"PublicIpAddressSet"`
			PrivateIpAddresses []string `json:"PrivateIpAddresses"`
		} `json:"InstanceSet"`
		TotalCount int    `json:"TotalCount"`
		RequestId  string `json:"RequestId"`
	} `json:"Response"`
}

// GetInstanceDetails 通过 DescribeInstances API 查询实例详情（公网 IP、状态等）
func (t *SpotVMManager) GetInstanceDetails(instanceId string) (*InstanceDetail, error) {
	params := map[string]any{
		"InstanceIds": []string{instanceId},
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("序列化 DescribeInstances 参数失败: %w", err)
	}
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "DescribeInstances")
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return nil, fmt.Errorf("设置 DescribeInstances 请求参数失败: %w", err)
	}
	response := tchttp.NewCommonResponse()
	err = t.cvmClient.Send(request, response)
	if err != nil {
		return nil, fmt.Errorf("调用 DescribeInstances API 失败: %w", err)
	}

	var descResp DescribeInstancesResponse
	err = json.Unmarshal(response.GetBody(), &descResp)
	if err != nil {
		return nil, fmt.Errorf("解析 DescribeInstances 响应失败: %w", err)
	}

	if len(descResp.Response.InstanceSet) == 0 {
		return nil, fmt.Errorf("未找到实例: %s", instanceId)
	}

	inst := descResp.Response.InstanceSet[0]
	detail := &InstanceDetail{
		InstanceId:   inst.InstanceId,
		InstanceType: inst.InstanceType,
		Zone:         inst.Placement.Zone,
		State:        inst.InstanceState,
	}
	if len(inst.PublicIpAddressSet) > 0 {
		detail.PublicIp = inst.PublicIpAddressSet[0]
	}
	if len(inst.PrivateIpAddresses) > 0 {
		detail.PrivateIp = inst.PrivateIpAddresses[0]
	}

	return detail, nil
}

// WaitForInstanceRunning 轮询等待实例状态变为 RUNNING 并返回公网 IP
func (t *SpotVMManager) WaitForInstanceRunning(instanceId string, timeoutSeconds int) (*InstanceDetail, error) {
	log.Printf("⏳ 等待实例 %s 进入 RUNNING 状态（超时: %ds）...", instanceId, timeoutSeconds)

	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("等待实例 %s 进入 RUNNING 状态超时（%ds）", instanceId, timeoutSeconds)
		case <-ticker.C:
			detail, err := t.GetInstanceDetails(instanceId)
			if err != nil {
				log.Printf("⚠️  查询实例 %s 状态失败: %v，继续等待...", instanceId, err)
				continue
			}
			log.Printf("🔍 实例 %s 当前状态: %s, 公网IP: %s", instanceId, detail.State, detail.PublicIp)
			if detail.State == "RUNNING" && detail.PublicIp != "" {
				log.Printf("✅ 实例 %s 已进入 RUNNING 状态，公网IP: %s", instanceId, detail.PublicIp)
				return detail, nil
			}
		}
	}
}
