package network

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

// SSLManager SSL管理器
type SSLManager struct {
	SSLClient     *common.Client
	CertificateId string
}

// NewSSLManager 创建SSL管理器
func NewSSLManager(region, certificateId string) *SSLManager {
	credential := common.NewCredential(os.Getenv("TENCENTCLOUD_SECRET_ID"), os.Getenv("TENCENTCLOUD_SECRET_KEY"))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	sslClient := common.NewCommonClient(credential, region, cpf).WithLogger(log.Default())
	return &SSLManager{
		SSLClient:     sslClient,
		CertificateId: certificateId,
	}
}

// GetSSLCertificate 获取SSL证书
func (s *SSLManager) GetSSLCertificate() (*models.GetSSLCertificateResponse, error) {
	log.Printf("🔐 获取SSL证书: certificateId=%s", s.CertificateId)
	params := map[string]any{
		"CertificateId": s.CertificateId,
	}
	request := tchttp.NewCommonRequest("ssl", "2019-12-05", "DescribeCertificateDetail")
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("序列化SSL请求参数失败: %w", err)
	}
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return nil, fmt.Errorf("设置SSL请求参数失败: %w", err)
	}
	response := tchttp.NewCommonResponse()
	err = s.SSLClient.Send(request, response)
	if err != nil {
		return nil, fmt.Errorf("获取SSL证书失败: %w", err)
	}
	var getSSLCertificateResponse models.GetSSLCertificateResponse
	err = json.Unmarshal(response.GetBody(), &getSSLCertificateResponse)
	if err != nil {
		return nil, fmt.Errorf("解析SSL证书响应失败: %w", err)
	}

	// T027: 证书有效期检查
	s.checkCertificateExpiry(&getSSLCertificateResponse)

	log.Printf("✅ SSL证书获取成功: certificateId=%s", s.CertificateId)
	return &getSSLCertificateResponse, nil
}

// checkCertificateExpiry 检查证书有效期，即将过期时输出警告
func (s *SSLManager) checkCertificateExpiry(resp *models.GetSSLCertificateResponse) {
	if resp == nil || resp.Response.CertEndTime == "" {
		return
	}
	// 尝试解析证书过期时间（腾讯云格式: 2024-12-31 23:59:59）
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if endTime, err := time.Parse(layout, resp.Response.CertEndTime); err == nil {
			daysUntilExpiry := int(time.Until(endTime).Hours() / 24)
			if daysUntilExpiry < 30 {
				log.Printf("⚠️  SSL证书即将过期！剩余 %d 天 (过期时间: %s)", daysUntilExpiry, resp.Response.CertEndTime)
			} else {
				log.Printf("🔐 SSL证书有效期: 剩余 %d 天", daysUntilExpiry)
			}
			return
		}
	}
}
