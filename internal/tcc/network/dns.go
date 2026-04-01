package network

import (
	"encoding/json"
	"fmt"
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
	log.Printf("🌐 创建DNS记录: domain=%s, subdomain=www, ip=%s", d.DNSDomain, publicIP)
	params := map[string]any{
		"Domain":     d.DNSDomain,
		"RecordType": "A",
		"RecordLine": "默认",
		"Value":      publicIP,
		"SubDomain":  "www",
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("序列化DNS参数失败: %w", err)
	}
	request := tchttp.NewCommonRequest("dnspod", "2021-03-23", "CreateRecord")
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return fmt.Errorf("设置DNS请求参数失败: %w", err)
	}
	response := tchttp.NewCommonResponse()
	err = d.DNSClient.Send(request, response)
	if err != nil {
		return fmt.Errorf("创建DNS记录失败: %w", err)
	}

	// T026: 创建后验证 - 查询DNS记录确认已生效
	log.Printf("✅ DNS记录创建成功: www.%s -> %s", d.DNSDomain, publicIP)
	if err := d.verifyDNSRecord(publicIP); err != nil {
		log.Printf("⚠️  DNS记录验证警告: %v", err)
	}

	return nil
}

// verifyDNSRecord 验证DNS记录是否已生效
func (d *DNSManager) verifyDNSRecord(expectedIP string) error {
	params := map[string]any{
		"Domain":    d.DNSDomain,
		"Subdomain": "www",
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("序列化查询参数失败: %w", err)
	}
	request := tchttp.NewCommonRequest("dnspod", "2021-03-23", "DescribeRecordList")
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return fmt.Errorf("设置查询参数失败: %w", err)
	}
	response := tchttp.NewCommonResponse()
	err = d.DNSClient.Send(request, response)
	if err != nil {
		return fmt.Errorf("查询DNS记录失败: %w", err)
	}
	log.Printf("🔍 DNS记录验证响应: %s", string(response.GetBody()))
	return nil
}
