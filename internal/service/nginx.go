package service

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Nginx 开放 80和443端口，443 端口配置 SSL 证书，80 端口重定向到 443 端口

// NginxService 管理Nginx服务的结构体
type NginxService struct {
	ContainerName  string
	SSLEnabled     bool
	SSLCertPath    string
	SSLKeyPath     string
	Port80         int
	Port443        int
	OS             string
	PackageManager string
}

// NewNginxService 创建新的Nginx服务实例
func NewNginxService(containerName string) *NginxService {
	os, pkgManager := detectOSAndPackageManager()
	sslCertPath := "/etc/nginx/ssl"

	return &NginxService{
		ContainerName:  containerName,
		SSLEnabled:     true,
		SSLCertPath:    sslCertPath,
		Port80:         80,
		Port443:        443,
		OS:             os,
		PackageManager: pkgManager,
	}
}

// CreateSSLCertificates 创建自签名SSL证书
func (n *NginxService) CreateSSLCertificates(publicKey, privateKey string) error {
	if !n.SSLEnabled {
		return nil
	}

	log.Println("创建SSL证书...")
	if err := os.MkdirAll(n.SSLCertPath, 0o755); err != nil {
		return err
	}
	publicKeyPath := filepath.Join(n.SSLCertPath, "cert.pem")
	err := os.WriteFile(publicKeyPath, []byte(publicKey), 0600)
	if err != nil {
		return err
	}
	privateKeyPath := filepath.Join(n.SSLCertPath, "cert.key")
	err = os.WriteFile(privateKeyPath, []byte(privateKey), 0600)
	if err != nil {
		return err
	}
	log.Println("SSL证书创建完成")
	return nil
}

// CreateNginxConfig 创建Nginx配置文件
func (n *NginxService) CreateNginxConfig() (string, error) {
	log.Println("创建Nginx配置文件...")

	configDir := "/tmp/nginx-config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %v", err)
	}

	configPath := filepath.Join(configDir, "nginx.conf")
	var configContent string

	if n.SSLEnabled {
		configContent = n.generateSSLConfig()
	} else {
		configContent = n.generateHTTPConfig()
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %v", err)
	}

	log.Println("Nginx配置文件创建完成")
	return configPath, nil
}

// generateSSLConfig 生成支持SSL的Nginx配置
func (n *NginxService) generateSSLConfig() string {
	return fmt.Sprintf(`events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    # 重定向HTTP到HTTPS
    server {
        listen 80;
        server_name _;
        return 301 https://$host$request_uri;
    }

    # HTTPS服务器
    server {
        listen 443 ssl;
        server_name _;

        ssl_certificate %s;
        ssl_certificate_key %s;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        location / {
            root   /usr/share/nginx/html;
            index  index.html index.htm;
        }

        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}`, fmt.Sprintf("%s/cert.pem", n.SSLCertPath), fmt.Sprintf("%s/cert.key", n.SSLCertPath))
}

// generateHTTPConfig 生成HTTP的Nginx配置
func (n *NginxService) generateHTTPConfig() string {
	return `events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    server {
        listen 80;
        server_name _;

        location / {
            root   /usr/share/nginx/html;
            index  index.html index.htm;
        }

        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}`
}

// CreateIndexHTML 创建默认的index.html文件
func (n *NginxService) CreateIndexHTML() (string, error) {
	log.Println("创建默认HTML页面...")

	htmlDir := "/tmp/nginx-html"
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		return "", fmt.Errorf("创建HTML目录失败: %v", err)
	}

	htmlPath := filepath.Join(htmlDir, "index.html")
	htmlContent := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nginx Server</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 10px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 500px;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .status {
            background: #4CAF50;
            color: white;
            padding: 0.5rem 1rem;
            border-radius: 5px;
            display: inline-block;
            margin-top: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Nginx Server</h1>
        <p>恭喜！您的Nginx服务器已经成功运行。</p>
        <p>这是一个由Go程序自动部署的Nginx服务器。</p>
        <div class="status">✅ 服务运行正常</div>
    </div>
</body>
</html>`

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return "", fmt.Errorf("写入HTML文件失败: %v", err)
	}

	log.Println("默认HTML页面创建完成")
	return htmlDir, nil
}

// RunNginxContainer 运行Nginx容器
func (n *NginxService) RunNginxContainer(configPath, htmlPath string) error {
	log.Println("启动Nginx容器...")

	// 停止并删除已存在的容器
	n.stopAndRemoveContainer()

	// 构建docker run命令
	var cmd string
	if n.SSLEnabled {
		cmd = fmt.Sprintf(
			"docker run -d --name %s "+
				"-p %d:80 -p %d:443 "+
				"-v %s:/etc/nginx/nginx.conf:ro "+
				"-v %s:/usr/share/nginx/html:ro "+
				"-v %s:/etc/nginx/ssl:ro "+
				"--restart unless-stopped "+
				"nginx:alpine",
			n.ContainerName, n.Port80, n.Port443,
			configPath, htmlPath, n.SSLCertPath)
	} else {
		cmd = fmt.Sprintf(
			"docker run -d --name %s "+
				"-p %d:80 "+
				"-v %s:/etc/nginx/nginx.conf:ro "+
				"-v %s:/usr/share/nginx/html:ro "+
				"--restart unless-stopped "+
				"nginx:alpine",
			n.ContainerName, n.Port80,
			configPath, htmlPath)
	}

	if err := n.executeCommand(cmd); err != nil {
		return fmt.Errorf("启动Nginx容器失败: %v", err)
	}

	log.Println("Nginx容器启动成功")
	return nil
}

// stopAndRemoveContainer 停止并删除容器
func (n *NginxService) stopAndRemoveContainer() {
	stopCmd := fmt.Sprintf("docker stop %s", n.ContainerName)
	removeCmd := fmt.Sprintf("docker rm %s", n.ContainerName)

	n.executeCommand(stopCmd)   // 忽略错误，容器可能不存在
	n.executeCommand(removeCmd) // 忽略错误，容器可能不存在
}

// CheckContainerStatus 检查容器状态
func (n *NginxService) CheckContainerStatus() error {
	log.Println("检查容器状态...")

	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", n.ContainerName), "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("检查容器状态失败: %v", err)
	}

	log.Printf("容器状态:\n%s", string(output))
	return nil
}

// GetContainerLogs 获取容器日志
func (n *NginxService) GetContainerLogs() error {
	log.Println("获取容器日志...")

	cmd := exec.Command("docker", "logs", n.ContainerName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("获取容器日志失败: %v", err)
	}

	log.Printf("容器日志:\n%s", string(output))
	return nil
}

// StartNginxService 启动完整的Nginx服务
func (n *NginxService) StartNginxService(publicKey, privateKey string) error {
	log.Println("开始部署Nginx服务...")

	// 1. 安装Docker
	if err := n.InstallDocker(); err != nil {
		return err
	}

	// 2. 创建SSL证书（如果启用）
	if err := n.CreateSSLCertificates(publicKey, privateKey); err != nil {
		return err
	}

	// 3. 创建Nginx配置
	configPath, err := n.CreateNginxConfig()
	if err != nil {
		return err
	}

	// 4. 创建HTML文件
	htmlPath, err := n.CreateIndexHTML()
	if err != nil {
		return err
	}

	// 5. 运行Nginx容器
	if err := n.RunNginxContainer(configPath, htmlPath); err != nil {
		return err
	}

	// 6. 等待容器启动
	time.Sleep(3 * time.Second)

	// 7. 检查状态
	if err := n.CheckContainerStatus(); err != nil {
		return err
	}

	log.Println("Nginx服务部署完成！")
	if n.SSLEnabled {
		log.Printf("访问地址: https://localhost:%d", n.Port443)
	} else {
		log.Printf("访问地址: http://localhost:%d", n.Port80)
	}

	return nil
}

// executeCommand 执行shell命令
func (n *NginxService) executeCommand(cmd string) error {
	log.Printf("执行命令: %s", cmd)

	// 使用bash执行命令
	execCmd := exec.Command("bash", "-c", cmd)

	// 获取输出
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return err
	}

	// 启动命令
	if err := execCmd.Start(); err != nil {
		return err
	}

	// 读取输出
	go n.readOutput(stdout, "STDOUT")
	go n.readOutput(stderr, "STDERR")

	// 等待命令完成
	return execCmd.Wait()
}

// readOutput 读取命令输出
func (n *NginxService) readOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Printf("[%s] %s", prefix, scanner.Text())
	}
}
