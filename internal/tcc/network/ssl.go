package network

import (
	"encoding/json"
	"log"
	"os"

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
	params := map[string]any{
		"CertificateId": s.CertificateId,
	}
	request := tchttp.NewCommonRequest("ssl", "2019-12-05", "DescribeCertificateDetail")
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	err = request.SetActionParameters(jsonParams)
	if err != nil {
		return nil, err
	}
	response := tchttp.NewCommonResponse()
	err = s.SSLClient.Send(request, response)
	if err != nil {
		return nil, err
	}
	var getSSLCertificateResponse models.GetSSLCertificateResponse
	err = json.Unmarshal(response.GetBody(), &getSSLCertificateResponse)
	if err != nil {
		return nil, err
	}
	return &getSSLCertificateResponse, nil
}
