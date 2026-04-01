package models

import "time"

// MigrationStatus 迁移状态枚举
type MigrationStatus string

const (
	MigrationStatusPending      MigrationStatus = "PENDING"      // 迁移任务已创建，等待新实例创建完成
	MigrationStatusWaitingSSH   MigrationStatus = "WAITING_SSH"  // 等待新实例 SSH 端口就绪
	MigrationStatusTransferring MigrationStatus = "TRANSFERRING" // 正在通过 SCP 传输文件
	MigrationStatusStarting     MigrationStatus = "STARTING"     // 正在通过 SSH 远程启动程序
	MigrationStatusVerifying    MigrationStatus = "VERIFYING"    // 正在验证新实例上的程序是否正常运行
	MigrationStatusCompleted    MigrationStatus = "COMPLETED"    // 迁移完成
	MigrationStatusFailed       MigrationStatus = "FAILED"       // 迁移失败
)

// MigrationTask 代表一次程序迁移操作
type MigrationTask struct {
	TargetInstanceId string          `json:"target_instance_id"` // 目标实例 ID
	TargetIP         string          `json:"target_ip"`          // 目标实例公网 IP
	TargetRegion     string          `json:"target_region"`      // 目标 Region
	TargetZone       string          `json:"target_zone"`        // 目标 Zone（系统自动选择）
	Status           MigrationStatus `json:"status"`             // 迁移状态
	RetryCount       int             `json:"retry_count"`        // 已重试次数
	MaxRetries       int             `json:"max_retries"`        // 最大重试次数
	StartTime        time.Time       `json:"start_time"`         // 迁移开始时间
	Error            string          `json:"error,omitempty"`    // 最后一次错误信息
}

// GetSSLCertificateResponse 获取SSL证书响应
type GetSSLCertificateResponse struct {
	Response struct {
		CertificateId         string `json:"CertificateId"`
		CertificatePrivateKey string `json:"CertificatePrivateKey"`
		CertificatePublicKey  string `json:"CertificatePublicKey"`
		CertEndTime           string `json:"CertEndTime"`
		RequestId             string `json:"RequestId"`
	} `json:"Response"`
}

// Zone 可用区
type Zone struct {
	Zone      string `json:"Zone"`
	ZoneId    string `json:"ZoneId"`
	ZoneName  string `json:"ZoneName"`
	ZoneState string `json:"ZoneState"`
}

// ZoneResponse 可用区响应
type ZoneResponse struct {
	Response struct {
		ZoneSet    []*Zone `json:"ZoneSet"`
		TotalCount int     `json:"TotalCount"`
		RequestId  string  `json:"RequestId"`
	} `json:"Response"`
}

// 定义响应结构体
type GetInstanceResponse struct {
	Response struct {
		InstanceTypeQuotaSet []InstanceType `json:"InstanceTypeQuotaSet"`
		RequestId            string         `json:"RequestId"`
	} `json:"Response"`
}

// InstanceType 实例类型
type InstanceType struct {
	InstanceType      string  `json:"InstanceType"`
	InstanceFamily    string  `json:"InstanceFamily"`
	TypeName          string  `json:"TypeName"`
	Cpu               int     `json:"Cpu"`
	Memory            int     `json:"Memory"`
	CpuType           string  `json:"CpuType"`
	Frequency         string  `json:"Frequency"`
	Gpu               int     `json:"Gpu"`
	GpuCount          int     `json:"GpuCount"`
	InstanceBandwidth float64 `json:"InstanceBandwidth"`
	InstancePps       int     `json:"InstancePps"`
	Status            string  `json:"Status"`
	StatusCategory    string  `json:"StatusCategory"`
	Price             Price   `json:"Price"`
	Zone              string  `json:"Zone"`
}

// CreateInstanceResponse 创建实例响应
type CreateInstanceResponse struct {
	Response struct {
		InstanceIdSet []string `json:"InstanceIdSet"`
		RequestId     string   `json:"RequestId"`
	} `json:"Response"`
}

// Price 价格
type Price struct {
	ChargeUnit        string  `json:"ChargeUnit"`
	UnitPrice         float64 `json:"UnitPrice"`
	UnitPriceDiscount float64 `json:"UnitPriceDiscount"`
	Discount          float64 `json:"Discount"`
}
