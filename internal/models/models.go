package models

// GetSSLCertificateResponse 获取SSL证书响应
type GetSSLCertificateResponse struct {
	Response struct {
		CertificateId         string `json:"CertificateId"`
		CertificatePrivateKey string `json:"CertificatePrivateKey"`
		CertificatePublicKey  string `json:"CertificatePublicKey"`
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
