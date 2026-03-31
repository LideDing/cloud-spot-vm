package service

import (
	"log"
	"sort"
	"time"

	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/models"
	"gitee.com/dinglide/spot-vm/internal/tcc"
)

type SpotService struct {
	Region        string
	Domain        string
	CertificateId string
	tccClient     *tcc.TCC
	allInstances  []*models.InstanceType
}

func NewSpotService(cfg *config.Config) (*SpotService, error) {
	tccClient, err := tcc.NewTCC(cfg.TENCENTCLOUD_SECRET_ID, cfg.TENCENTCLOUD_SECRET_KEY, cfg.Region)
	if err != nil {
		return nil, err
	}
	return &SpotService{
		Region:        cfg.Region,
		Domain:        cfg.Domain,
		CertificateId: cfg.CertificateId,
		tccClient:     tccClient,
	}, nil
}

// 获取所有实例类型
func (s *SpotService) getAllInstances() error {
	zones, err := s.tccClient.SpotVMManager.GetAvailableZones()
	if err != nil {
		return err
	}
	// 收集所有实例信息
	allInstances := []*models.InstanceType{}
	for _, zone := range zones {
		// 获取该可用区的所有实例类型
		instances, err := s.tccClient.SpotVMManager.GetAvailableInstanceTypes(zone.Zone)
		if err != nil {
			continue
		}
		allInstances = append(allInstances, instances...)
	}
	// 按价格排序
	sort.Slice(allInstances, func(i, j int) bool {
		return allInstances[i].Price.UnitPriceDiscount < allInstances[j].Price.UnitPriceDiscount
	})
	s.allInstances = allInstances
	return nil
}

func (s *SpotService) CreateSpotInstance() error {
	instanceType := s.allInstances[0]
	instanceId, err := s.tccClient.SpotVMManager.CreateSpotInstance(instanceType.Zone, instanceType.InstanceType, false)
	if err != nil {
		return err
	}
	log.Printf("创建实例: %s", instanceId)
	return nil
}

func (s *SpotService) Run() {
	// 每 10 秒获取一次实例类型
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				s.tccClient.SpotVMManager.SpotVM.UpdateInstanceState() // 更新实例状态
				err := s.getAllInstances()
				if err != nil {
					log.Fatalf("获取实例类型失败: %v", err)
				}
			case <-s.tccClient.SpotVMManager.SpotVM.Terminated:
				err := s.getAllInstances()
				if err != nil {
					log.Fatalf("获取实例类型失败，使用上一次的实例类型: %v", err)
				}
				err = s.CreateSpotInstance()
				if err != nil {
					log.Fatalf("创建实例失败: %v", err)
				}
				return
			}
		}
	}()
}
