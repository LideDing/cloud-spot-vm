package spot_vm

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// SpotVM Spot VM
type SpotVM struct {
	InstanceState *instanceState
	Terminated    chan struct{}
}

// NewSpotVM 创建Spot VM
func NewSpotVM() *SpotVM {
	return &SpotVM{
		Terminated:    make(chan struct{}),
		InstanceState: &instanceState{},
	}
}

// instanceState 实例状态
type instanceState struct {
	PublicIp     *string
	PrivateIp    *string
	InstanceId   *string
	InstanceType *string
	Zone         *string
}

// UpdateInstanceState 持续更新实例状态
func (s *SpotVM) UpdateInstanceState() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		if s.InstanceState.InstanceId == nil {
			instanceId, err := s.getMetadataInstanceId()
			if err != nil {
				fmt.Println("fail to get instance id:", err.Error())
			}
			s.InstanceState.InstanceId = &instanceId
		}
		if s.InstanceState.PublicIp == nil {
			publicIp, err := s.getMetadataPublicIp()
			if err != nil {
				fmt.Println("fail to get public ip:", err.Error())
			}
			s.InstanceState.PublicIp = &publicIp
		}
		if s.InstanceState.PrivateIp == nil {
			privateIp, err := s.getMetadataPrivateIp()
			if err != nil {
				fmt.Println("fail to get private ip:", err.Error())
			}
			s.InstanceState.PrivateIp = &privateIp
		}
		if s.InstanceState.InstanceType == nil {
			instanceType, err := s.getMetadataInstanceType()
			if err != nil {
				fmt.Println("fail to get instance type:", err.Error())
			}
			s.InstanceState.InstanceType = &instanceType
		}
		if s.InstanceState.Zone == nil {
			zone, err := s.getMetadataZone()
			if err != nil {
				fmt.Println("fail to get zone:", err.Error())
			}
			s.InstanceState.Zone = &zone
		}
		if s.getInstanceTerminated() {
			s.Terminated <- struct{}{}
		}
		<-ticker.C
	}
}

// getMetadataInstanceId 获取实例ID
func (s *SpotVM) getMetadataInstanceId() (string, error) {
	response, err := http.Get(metadataUrl + "instance-id")
	if err != nil {
		return "", err
	}
	instanceId, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(instanceId), nil
}

// getMetadataPublicIp 获取实例公网IP
func (s *SpotVM) getMetadataPublicIp() (string, error) {
	response, err := http.Get(metadataUrl + "public-ipv4")
	if err != nil {
		return "", err
	}
	publicIp, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(publicIp), nil
}

// getMetadataPrivateIp 获取实例私有IP
func (s *SpotVM) getMetadataPrivateIp() (string, error) {
	response, err := http.Get(metadataUrl + "local-ipv4")
	if err != nil {
		return "", err
	}
	privateIp, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(privateIp), nil
}

// getMetadataInstanceType 获取实例类型
func (s *SpotVM) getMetadataInstanceType() (string, error) {
	response, err := http.Get(metadataUrl + "instance-id")
	if err != nil {
		return "", err
	}
	instanceType, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(instanceType), nil
}

// getMetadataZone 获取实例所在可用区
func (s *SpotVM) getMetadataZone() (string, error) {
	response, err := http.Get(metadataUrl + "placement/zone")
	if err != nil {
		return "", err
	}
	zone, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(zone), nil
}

// getInstanceTerminated 获取实例是否被终止
func (s *SpotVM) getInstanceTerminated() bool {
	// 404表示不会被终止，这是正常情况
	resp, err := http.Get(metadataUrl + "spot/termination-time")
	if err != nil {
		return false
	}
	if resp.StatusCode == 404 {
		return false
	}
	return true
}

// GetInstanceInfo 获取实例信息
func (s *SpotVM) GetInstanceInfo() map[string]interface{} {
	info := make(map[string]interface{})

	if s.InstanceState.InstanceId != nil {
		info["InstanceId"] = *s.InstanceState.InstanceId
	}
	if s.InstanceState.InstanceType != nil {
		info["InstanceType"] = *s.InstanceState.InstanceType
	}
	if s.InstanceState.Zone != nil {
		info["Zone"] = *s.InstanceState.Zone
	}
	if s.InstanceState.PublicIp != nil {
		info["PublicIpAddress"] = *s.InstanceState.PublicIp
	}
	if s.InstanceState.PrivateIp != nil {
		info["PrivateIpAddress"] = *s.InstanceState.PrivateIp
	}

	// 添加一些默认值
	info["Status"] = "RUNNING"
	info["CreatedTime"] = time.Now().Format(time.RFC3339)
	info["ExpiredTime"] = "N/A"

	return info
}
