// Package spot_vm 管理腾讯云Spot VM
package spot_vm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gitee.com/dinglide/spot-vm/internal/models"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const (
	metadataUrl = "http://metadata.tencentyun.com/latest/meta-data/"
)

// SpotVMManager Spot VM管理器
type SpotVMManager struct {
	cvmClient  *common.Client // 腾讯云CVM客户端
	Terminated chan struct{}  // 实例终止信号
	SpotVM     *SpotVM        // Spot VM实例
}

// NewSpotVMManager 创建Spot VM管理器
func NewSpotVMManager(region string) *SpotVMManager {
	credential := common.NewCredential(os.Getenv("TENCENTCLOUD_SECRET_ID"), os.Getenv("TENCENTCLOUD_SECRET_KEY"))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	cvmClient := common.NewCommonClient(credential, region, cpf).WithLogger(log.Default())
	return &SpotVMManager{
		cvmClient:  cvmClient,
		Terminated: make(chan struct{}),
		SpotVM:     NewSpotVM(),
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
	request := tchttp.NewCommonRequest("cvm", "2017-03-12", "RunInstances")
	params := map[string]any{
		"InstanceChargeType": "SPOTPAID",
		"Placement":          map[string]any{"Zone": zone},
		"InstanceType":       instanceType,
		"ImageId":            "img-hdt9xxkt", //"img-hdt9xxkt",
		"SystemDisk": map[string]any{
			"DiskType": "CLOUD_BSSD",
			"DiskSize": 20,
		},
		// "VirtualPrivateCloud": map[string]any{
		// 	"VpcId":    "vpc-enajr3z3",
		// 	"SubnetId": "subnet-g7te823k",
		// },
		"InternetAccessible": map[string]any{
			"InternetChargeType":      "TRAFFIC_POSTPAID_BY_HOUR",
			"InternetMaxBandwidthOut": 10,
			"PublicIpAssigned":        true,
		},
		"InstanceCount": 1,
		"LoginSettings": map[string]any{"Password": "1qazZSE$"},
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
	if dryRun {
		fmt.Println(string(response.GetBody()))
		fmt.Println("dryRun success, no error")
		return nil, nil
	}

	var createInstanceResponse models.CreateInstanceResponse
	err = json.Unmarshal(response.GetBody(), &createInstanceResponse)
	if err != nil {
		return nil, err
	}
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
		fmt.Println("fail to invoke api:", err.Error())
		return nil, err
	}
	var availableZonesResponse models.ZoneResponse
	err = json.Unmarshal(response.GetBody(), &availableZonesResponse)
	if err != nil {
		fmt.Println("fail to parse response:", err.Error())
		return nil, err
	}
	availableZones := []*models.Zone{}
	for _, zone := range availableZonesResponse.Response.ZoneSet {
		if zone.ZoneState == "AVAILABLE" {
			availableZones = append(availableZones, zone)
		}
	}
	return availableZones, nil
}
