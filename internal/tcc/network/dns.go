package network

import (
	"encoding/json"
	"log"
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// DNSManager DNS管理器
type DNSManager struct {
	DNSClient *common.Client
	DNSDomain string
}

// NewDNSManager 创建DNS管理器
func NewDNSManager(region, domain string) *DNSManager {
	credential := common.NewCredential(os.Getenv("TENCENTCLOUD_SECRET_ID"), os.Getenv("TENCENTCLOUD_SECRET_KEY"))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "dns.tencentcloudapi.com"
	dnspodClient := common.NewCommonClient(credential, region, cpf).WithLogger(log.Default())
	return &DNSManager{
		DNSClient: dnspodClient,
		DNSDomain: domain,
	}
}

// CreateDNSHost 创建DNS主机
func (d *DNSManager) CreateDNSHost(publicIP string) error {
	params := map[string]any{
		"Domain":     d.DNSDomain,
		"RecordType": "A",
		"RecordLine": "默认",
		"Value":      publicIP,
		"SubDomain":  "www",
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return err
	}
	request := tchttp.NewCommonRequest("dnspod", "2021-03-23", "CreateRecord")
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return err
	}
	response := tchttp.NewCommonResponse()
	err = d.DNSClient.Send(request, response)
	if err != nil {
		return err
	}
	return nil
}
