package service

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// detectOSAndPackageManager 检测操作系统和包管理器
func detectOSAndPackageManager() (string, string) {
	// 检查是否是Linux
	if runtime.GOOS != "linux" {
		return runtime.GOOS, "unsupported"
	}

	// 检查/etc/os-release文件
	if _, err := os.Stat("/etc/os-release"); err == nil {
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "ubuntu") || strings.Contains(content, "debian") {
				return "ubuntu", "apt-get"
			} else if strings.Contains(content, "centos") || strings.Contains(content, "rhel") || strings.Contains(content, "fedora") {
				return "centos", "yum"
			} else if strings.Contains(content, "amazon") {
				return "amazon", "yum"
			}
		}
	}

	// 检查包管理器是否存在
	if _, err := exec.LookPath("apt-get"); err == nil {
		return "debian", "apt-get"
	}
	if _, err := exec.LookPath("yum"); err == nil {
		return "rhel", "yum"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "fedora", "dnf"
	}

	return "unknown", "unknown"
}

// InstallDocker 在Linux上安装Docker
func (n *NginxService) InstallDocker() error {
	log.Println("开始安装Docker...")
	log.Printf("检测到操作系统: %s, 包管理器: %s", n.OS, n.PackageManager)

	// 检查Docker是否已安装
	if n.isDockerInstalled() {
		log.Println("Docker已经安装")
		return nil
	}

	// 根据不同的包管理器执行不同的安装命令
	var commands []string

	switch n.PackageManager {
	case "apt-get":
		commands = n.GetAptGetCommands()
	case "yum":
		commands = n.GetYumCommands()
	case "dnf":
		commands = n.GetDnfCommands()
	default:
		return fmt.Errorf("不支持的包管理器: %s", n.PackageManager)
	}

	for _, cmd := range commands {
		log.Printf("执行命令: %s", cmd)
		if err := n.executeCommand(cmd); err != nil {
			return fmt.Errorf("安装Docker失败: %v", err)
		}
	}

	// T024: 安装结果验证
	if !n.isDockerInstalled() {
		return fmt.Errorf("Docker安装完成但验证失败: docker --version 执行失败，请检查安装日志")
	}

	// 获取并记录Docker版本
	versionCmd := exec.Command("docker", "--version")
	if output, err := versionCmd.Output(); err == nil {
		log.Printf("Docker安装完成: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// GetAptGetCommands 获取apt-get安装Docker的命令
func (n *NginxService) GetAptGetCommands() []string {
	return []string{
		"apt-get update",
		"apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg",
		"echo \"deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" | tee /etc/apt/sources.list.d/docker.list > /dev/null",
		"apt-get update",
		"apt-get install -y docker-ce docker-ce-cli containerd.io",
		"systemctl start docker",
		"systemctl enable docker",
		"usermod -aG docker $USER",
	}
}

// GetYumCommands 获取yum安装Docker的命令
func (n *NginxService) GetYumCommands() []string {
	return []string{
		"yum update -y",
		"yum install -y yum-utils",
		"yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo",
		"yum install -y docker-ce docker-ce-cli containerd.io",
		"systemctl start docker",
		"systemctl enable docker",
		"usermod -aG docker $USER",
	}
}

// GetDnfCommands 获取dnf安装Docker的命令
func (n *NginxService) GetDnfCommands() []string {
	return []string{
		"dnf update -y",
		"dnf install -y dnf-utils",
		"dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo",
		"dnf install -y docker-ce docker-ce-cli containerd.io",
		"systemctl start docker",
		"systemctl enable docker",
		"usermod -aG docker $USER",
	}
}

// isDockerInstalled 检查Docker是否已安装
func (n *NginxService) isDockerInstalled() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}
