package migration

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gitee.com/dinglide/spot-vm/internal/config"
	"golang.org/x/crypto/ssh"
)

// Migrator SSH/SCP 迁移引擎
type Migrator struct {
	cfg    *config.Config
	client *ssh.Client // 当前活跃的 SSH 连接
}

// NewMigrator 创建迁移引擎实例
func NewMigrator(cfg *config.Config) *Migrator {
	return &Migrator{
		cfg: cfg,
	}
}

// WaitForSSH 等待目标实例 SSH 端口就绪
// 轮询 TCP 连接目标 IP 的 SSH 端口，每 5 秒重试，超时时间从配置读取
func (m *Migrator) WaitForSSH(targetIP string) error {
	addr := net.JoinHostPort(targetIP, fmt.Sprintf("%d", m.cfg.SSHPort))
	timeout := time.Duration(m.cfg.SSHWaitTimeout) * time.Second
	log.Printf("⏳ 等待 SSH 端口就绪: %s（超时: %v）...", addr, timeout)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			conn.Close()
			log.Printf("✅ SSH 端口已就绪: %s", addr)
			return nil
		}
		log.Printf("🔄 SSH 端口未就绪: %s, 5秒后重试...", addr)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("等待 SSH 端口 %s 就绪超时（%v）", addr, timeout)
}

// Connect 建立 SSH 连接（密码认证）
func (m *Migrator) Connect(targetIP string) error {
	addr := net.JoinHostPort(targetIP, fmt.Sprintf("%d", m.cfg.SSHPort))
	log.Printf("🔗 建立 SSH 连接: %s", addr)

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(m.cfg.InstancePassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(m.cfg.SSHTimeout) * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH 连接失败 (%s): %w", addr, err)
	}

	m.client = client
	log.Printf("✅ SSH 连接成功: %s", addr)
	return nil
}

// Close 关闭 SSH 连接
func (m *Migrator) Close() {
	if m.client != nil {
		m.client.Close()
		m.client = nil
		log.Println("🔌 SSH 连接已关闭")
	}
}

// ExecuteCommand 通过 SSH Session 执行远程命令并返回输出
func (m *Migrator) ExecuteCommand(command string) (string, error) {
	if m.client == nil {
		return "", fmt.Errorf("SSH 未连接")
	}

	session, err := m.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH Session 失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("执行远程命令失败 [%s]: %w, 输出: %s", command, err, string(output))
	}

	return string(output), nil
}

// TransferFile 通过 SCP 协议传输文件到远程实例
func (m *Migrator) TransferFile(localPath, remotePath string, mode os.FileMode) error {
	if m.client == nil {
		return fmt.Errorf("SSH 未连接")
	}

	// 读取本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败 (%s): %w", localPath, err)
	}
	defer localFile.Close()

	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败 (%s): %w", localPath, err)
	}

	// 创建 SSH Session
	session, err := m.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH Session 失败: %w", err)
	}
	defer session.Close()

	// 通过 SCP 协议传输
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", mode, stat.Size(), filepath.Base(remotePath))
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	remoteDir := filepath.Dir(remotePath)
	err = session.Run(fmt.Sprintf("scp -t %s", remoteDir))
	if err != nil {
		return fmt.Errorf("SCP 传输失败 (%s -> %s): %w", localPath, remotePath, err)
	}

	log.Printf("✅ 文件传输成功: %s -> %s (%d bytes)", localPath, remotePath, stat.Size())
	return nil
}

// HealthCheck 通过 HTTP GET 请求验证新实例上的程序是否正常运行
func (m *Migrator) HealthCheck(targetIP string, port string) error {
	url := fmt.Sprintf("http://%s:%s/api/v1/health", targetIP, port)
	log.Printf("🏥 开始健康检查: %s", url)

	client := &http.Client{Timeout: 5 * time.Second}
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Printf("✅ 健康检查通过: %s (第%d次尝试)", url, i+1)
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		log.Printf("🔄 健康检查失败 (第%d/%d次): %v, 5秒后重试...", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("健康检查失败: %s, 已重试 %d 次", url, maxRetries)
}

// Migrate 编排完整迁移流程
// 等待 SSH → 创建远程目录 → SCP 传输二进制文件 → SCP 传输 .env → SCP 传输 SSL 证书 → 远程启动程序 → 健康检查
func (m *Migrator) Migrate(targetIP string) error {
	log.Printf("🚀 开始迁移流程: 目标IP=%s", targetIP)

	// 1. 等待 SSH 端口就绪
	if err := m.WaitForSSH(targetIP); err != nil {
		return fmt.Errorf("等待 SSH 失败: %w", err)
	}

	// 2. 建立 SSH 连接
	if err := m.Connect(targetIP); err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer m.Close()

	// 3. 创建远程目录
	remoteDir := filepath.Dir(m.cfg.RemoteBinaryPath)
	log.Printf("📁 创建远程目录: %s", remoteDir)
	if _, err := m.ExecuteCommand(fmt.Sprintf("mkdir -p %s", remoteDir)); err != nil {
		return fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 4. 传输二进制文件
	localBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前可执行文件路径失败: %w", err)
	}
	log.Printf("📦 传输二进制文件: %s -> %s", localBinary, m.cfg.RemoteBinaryPath)
	if err := m.TransferFile(localBinary, m.cfg.RemoteBinaryPath, 0755); err != nil {
		return fmt.Errorf("传输二进制文件失败: %w", err)
	}

	// 5. 传输 .env 配置文件
	localEnvPath := ".env"
	if _, err := os.Stat(localEnvPath); err == nil {
		log.Printf("📦 传输 .env 文件: %s -> %s", localEnvPath, m.cfg.RemoteEnvPath)
		if err := m.TransferFile(localEnvPath, m.cfg.RemoteEnvPath, 0600); err != nil {
			return fmt.Errorf("传输 .env 文件失败: %w", err)
		}
	} else {
		log.Println("⚠️  本地 .env 文件不存在，跳过传输")
	}

	// 6. 传输 SSL 证书文件（如存在）
	sslCertPath := "/etc/nginx/ssl/cert.pem"
	sslKeyPath := "/etc/nginx/ssl/cert.key"
	if _, err := os.Stat(sslCertPath); err == nil {
		log.Println("📦 传输 SSL 证书文件...")
		if _, err := m.ExecuteCommand("mkdir -p /etc/nginx/ssl"); err != nil {
			log.Printf("⚠️  创建远程 SSL 目录失败: %v", err)
		}
		if err := m.TransferFile(sslCertPath, "/etc/nginx/ssl/cert.pem", 0600); err != nil {
			log.Printf("⚠️  传输 SSL 证书失败: %v", err)
		}
		if err := m.TransferFile(sslKeyPath, "/etc/nginx/ssl/cert.key", 0600); err != nil {
			log.Printf("⚠️  传输 SSL 密钥失败: %v", err)
		}
	}

	// 7. 远程启动程序（nohup 后台运行）
	port := m.cfg.Port
	if port == "" {
		port = "8080"
	}
	startCmd := fmt.Sprintf("cd %s && nohup %s > /var/log/spot-manager.log 2>&1 &",
		remoteDir, m.cfg.RemoteBinaryPath)
	log.Printf("🚀 远程启动程序: %s", startCmd)
	if _, err := m.ExecuteCommand(startCmd); err != nil {
		return fmt.Errorf("远程启动程序失败: %w", err)
	}

	// 等待程序启动
	log.Println("⏳ 等待远程程序启动（10秒）...")
	time.Sleep(10 * time.Second)

	// 8. 健康检查
	if err := m.HealthCheck(targetIP, port); err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}

	log.Printf("🎉 迁移完成: 目标IP=%s", targetIP)
	return nil
}

// DeployNginxRemotely 通过 SSH 在远程实例上部署 Docker + Nginx
func (m *Migrator) DeployNginxRemotely(targetIP string, sslCertContent, sslKeyContent string) error {
	log.Printf("🐳 开始远程部署 Nginx: %s", targetIP)

	// 确保 SSH 已连接
	if m.client == nil {
		if err := m.Connect(targetIP); err != nil {
			return fmt.Errorf("SSH 连接失败: %w", err)
		}
		defer m.Close()
	}

	// 1. 安装 Docker（如未安装）
	log.Println("🐳 检查并安装 Docker...")
	installDockerCmd := `
which docker > /dev/null 2>&1 || {
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl start docker
    systemctl enable docker
}
docker --version
`
	output, err := m.ExecuteCommand(installDockerCmd)
	if err != nil {
		return fmt.Errorf("安装 Docker 失败: %w, 输出: %s", err, output)
	}
	log.Printf("✅ Docker 就绪: %s", output)

	// 2. 写入 SSL 证书（如提供）
	if sslCertContent != "" && sslKeyContent != "" {
		log.Println("🔐 写入 SSL 证书...")
		mkdirCmd := "mkdir -p /etc/nginx/ssl"
		if _, err := m.ExecuteCommand(mkdirCmd); err != nil {
			return fmt.Errorf("创建 SSL 目录失败: %w", err)
		}
		certCmd := fmt.Sprintf("cat > /etc/nginx/ssl/cert.pem << 'CERTEOF'\n%s\nCERTEOF", sslCertContent)
		if _, err := m.ExecuteCommand(certCmd); err != nil {
			return fmt.Errorf("写入 SSL 证书失败: %w", err)
		}
		keyCmd := fmt.Sprintf("cat > /etc/nginx/ssl/cert.key << 'KEYEOF'\n%s\nKEYEOF", sslKeyContent)
		if _, err := m.ExecuteCommand(keyCmd); err != nil {
			return fmt.Errorf("写入 SSL 密钥失败: %w", err)
		}
	}

	// 3. 创建 Nginx 配置
	log.Println("📝 创建 Nginx 配置...")
	var nginxConf string
	if sslCertContent != "" {
		nginxConf = `events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    server {
        listen 80;
        server_name _;
        return 301 https://$host$request_uri;
    }
    server {
        listen 443 ssl;
        server_name _;
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/cert.key;
        ssl_protocols TLSv1.2 TLSv1.3;
        location / {
            root /usr/share/nginx/html;
            index index.html;
        }
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}`
	} else {
		nginxConf = `events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    server {
        listen 80;
        server_name _;
        location / {
            root /usr/share/nginx/html;
            index index.html;
        }
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}`
	}

	confCmd := fmt.Sprintf("mkdir -p /tmp/nginx-config && cat > /tmp/nginx-config/nginx.conf << 'CONFEOF'\n%s\nCONFEOF", nginxConf)
	if _, err := m.ExecuteCommand(confCmd); err != nil {
		return fmt.Errorf("创建 Nginx 配置失败: %w", err)
	}

	// 4. 创建默认 HTML
	htmlCmd := `mkdir -p /tmp/nginx-html && cat > /tmp/nginx-html/index.html << 'HTMLEOF'
<!DOCTYPE html>
<html><head><title>Nginx Server</title></head>
<body><h1>Nginx Server Running</h1><p>Deployed by spot-manager</p></body>
</html>
HTMLEOF`
	if _, err := m.ExecuteCommand(htmlCmd); err != nil {
		return fmt.Errorf("创建 HTML 文件失败: %w", err)
	}

	// 5. 停止旧容器并启动新容器
	log.Println("🐳 启动 Nginx 容器...")
	stopCmd := "docker stop spot-nginx 2>/dev/null; docker rm spot-nginx 2>/dev/null; true"
	m.ExecuteCommand(stopCmd)

	var runCmd string
	if sslCertContent != "" {
		runCmd = "docker run -d --name spot-nginx " +
			"-p 80:80 -p 443:443 " +
			"-v /tmp/nginx-config/nginx.conf:/etc/nginx/nginx.conf:ro " +
			"-v /tmp/nginx-html:/usr/share/nginx/html:ro " +
			"-v /etc/nginx/ssl:/etc/nginx/ssl:ro " +
			"--restart unless-stopped nginx:alpine"
	} else {
		runCmd = "docker run -d --name spot-nginx " +
			"-p 80:80 " +
			"-v /tmp/nginx-config/nginx.conf:/etc/nginx/nginx.conf:ro " +
			"-v /tmp/nginx-html:/usr/share/nginx/html:ro " +
			"--restart unless-stopped nginx:alpine"
	}

	output, err = m.ExecuteCommand(runCmd)
	if err != nil {
		return fmt.Errorf("启动 Nginx 容器失败: %w, 输出: %s", err, output)
	}

	log.Printf("✅ Nginx 容器启动成功: %s", output)

	// 6. 验证容器运行状态
	time.Sleep(3 * time.Second)
	statusOutput, err := m.ExecuteCommand("docker ps --filter name=spot-nginx --format '{{.Status}}'")
	if err != nil {
		log.Printf("⚠️  检查 Nginx 容器状态失败: %v", err)
	} else {
		log.Printf("🐳 Nginx 容器状态: %s", statusOutput)
	}

	log.Printf("🎉 Nginx 远程部署完成: %s", targetIP)
	return nil
}
