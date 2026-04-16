package service

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
	"modelmagic-deploy-console/backend/internal/logger"
)

// MmctlService 提供对 mmctl 脚本的封装
type MmctlService struct {
	baseDir    string
	scriptPath string
	workDir    string
}

// NewMmctlService 创建新的 MmctlService 实例
func NewMmctlService(baseDir, scriptPath, workDir string) *MmctlService {
	return &MmctlService{
		baseDir:    baseDir,
		scriptPath: scriptPath,
		workDir:    workDir,
	}
}

// LogCallback 日志回调函数类型
type LogCallback func(string)

// ListNamespaces 列出所有命名空间
func (s *MmctlService) ListNamespaces() ([]string, error) {
	// 简化实现：从工作目录读取
	namespaces := []string{}
	// 实际实现应该扫描 workspaces 目录
	return namespaces, nil
}

// GetNamespaceConfig 获取命名空间配置
func (s *MmctlService) GetNamespaceConfig(name string) (map[string]string, error) {
	// 简化实现
	return map[string]string{"name": name}, nil
}

// NewNamespace 创建新的命名空间配置
func (s *MmctlService) NewNamespace(namespace string) error {
	cmd := exec.Command("bash", s.scriptPath, "02", namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w, output: %s", err, string(output))
	}
	return nil
}

// UpdateNamespaceConfig 更新命名空间配置
func (s *MmctlService) UpdateNamespaceConfig(name string, config map[string]string) error {
	// 简化实现
	return nil
}

// Install 安装部署
func (s *MmctlService) Install(namespace string, step int, logCallback LogCallback) error {
	cmd := exec.Command("bash", s.scriptPath, "04", namespace, fmt.Sprintf("%d", step))
	
	// 实时日志输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	
	if err := cmd.Start(); err != nil {
		return err
	}
	
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			logger.Info("Install log", zap.String("line", line))
			if logCallback != nil {
				logCallback(line)
			}
		}
	}()
	
	return cmd.Wait()
}

// Uninstall 卸载
func (s *MmctlService) Uninstall(namespace string, step int, logCallback LogCallback) error {
	cmd := exec.Command("bash", s.scriptPath, "88", namespace, fmt.Sprintf("%d", step))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall: %w, output: %s", err, string(output))
	}
	return nil
}

// Rollback 回滚
func (s *MmctlService) Rollback(namespace, version string, logCallback LogCallback) error {
	cmd := exec.Command("bash", s.scriptPath, "90", namespace, version)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	
	if err := cmd.Start(); err != nil {
		return err
	}
	
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			logger.Info("Rollback log", zap.String("line", line))
			if logCallback != nil {
				logCallback(line)
			}
		}
	}()
	
	return cmd.Wait()
}

// Upgrade 升级
func (s *MmctlService) Upgrade(namespace, version string, logCallback LogCallback) error {
	cmd := exec.Command("bash", s.scriptPath, "89", namespace, version)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	
	if err := cmd.Start(); err != nil {
		return err
	}
	
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			logger.Info("Upgrade log", zap.String("line", line))
			if logCallback != nil {
				logCallback(line)
			}
		}
	}()
	
	return cmd.Wait()
}

// GetSystemAccess 获取系统访问地址
func (s *MmctlService) GetSystemAccess(name string) (string, error) {
	cmd := exec.Command("bash", s.scriptPath, "05", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get access: %w, output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// Scale 扩缩容
func (s *MmctlService) Scale(ctx context.Context, namespace string, replicas int) error {
	cmd := exec.Command("bash", s.scriptPath, "91", namespace, fmt.Sprintf("%d", replicas))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to scale: %w, output: %s", err, string(output))
	}
	return nil
}

// ValidateNamespace 验证命名空间名称格式
func ValidateNamespace(ns string) error {
	if ns == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	// 简单验证：只允许小写字母、数字和连字符
	if !isValidNamespaceName(ns) {
		return fmt.Errorf("invalid namespace name format")
	}
	return nil
}

// isValidNamespaceName 检查命名空间名称是否符合 K8s 规范
func isValidNamespaceName(name string) bool {
	// 简化验证：只允许小写字母、数字和连字符
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// StreamLogs 实时流式获取命名空间或应用的日志
// namespace: 命名空间名称
// appName: 应用名称（可选，为空则获取该命名空间下所有 Pod 日志）
// tailLines: 日志行数
// follow: 是否持续跟踪
func (s *MmctlService) StreamLogs(ctx context.Context, namespace, appName string, tailLines int, follow bool, callback LogCallback) error {
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	// 构建 kubectl logs 命令
	args := []string{"logs"}
	
	if appName != "" {
		// 如果是 Deployment/StatefulSet，需要加 -l app=appName
		args = append(args, "-l", "app="+appName)
	} else {
		// 获取该命名空间下所有 Pod 的日志
		args = append(args, "--all-namespaces=false")
	}
	
	args = append(args, "--namespace", namespace)
	
	if tailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tailLines))
	}
	
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Dir = s.workDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// 读取日志并回调
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Log.Debug("Log line", zap.String("line", line))
		if callback != nil {
			callback(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return cmd.Wait()
}
