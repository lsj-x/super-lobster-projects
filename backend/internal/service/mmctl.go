package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
	"modelmagic-deploy-console/backend/internal/logger"
)

// MmctlService 提供对 mmctl 脚本的封装
type MmctlService struct {
	baseDir  string
	scriptPath string
	workDir  string
}

// NewMmctlService 创建新的 MmctlService 实例
func NewMmctlService(baseDir, scriptPath, workDir string) *MmctlService {
	return &MmctlService{
		baseDir:  baseDir,
		scriptPath: scriptPath,
		workDir:  workDir,
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

// Rollback 回滚到指定版本
func (s *MmctlService) Rollback(namespace, version string, logCallback LogCallback) error {
	if err := ValidateNamespace(namespace); err != nil {
		return err
	}
	cmd := exec.Command("bash", s.scriptPath, "90", namespace, version)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Info("Rollback log", zap.String("line", line))
		if logCallback != nil {
			logCallback(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}
	return cmd.Wait()
}

// Upgrade 升级到指定版本
func (s *MmctlService) Upgrade(namespace, version string, logCallback LogCallback) error {
	if err := ValidateNamespace(namespace); err != nil {
		return err
	}
	cmd := exec.Command("bash", s.scriptPath, "89", namespace, version)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Info("Upgrade log", zap.String("line", line))
		if logCallback != nil {
			logCallback(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}
	return cmd.Wait()
}

// GetSystemAccess 获取系统访问地址
func (s *MmctlService) GetSystemAccess(namespace string) (string, error) {
	// 简化实现
	return fmt.Sprintf("https://%s.example.com", namespace), nil
}

// ValidateNamespace 验证命名空间格式
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	// 命名空间只能包含小写字母、数字和连字符
	for _, r := range namespace {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return fmt.Errorf("invalid character '%c' in namespace", r)
		}
	}
	return nil
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

// GetEvents 获取命名空间下的事件
// namespace: 命名空间名称
// limit: 返回事件数量限制
// typeFilter: 事件类型过滤 (Normal, Warning, 或空表示全部)
func (s *MmctlService) GetEvents(ctx context.Context, namespace string, limit int, typeFilter string) ([]map[string]interface{}, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	args := []string{"get", "events", "--namespace", namespace, "-o", "json"}
	if limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Dir = s.workDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w, output: %s", err, string(output))
	}

	// 解析 JSON 输出
	var events struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(output, &events); err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	// 过滤事件类型
	result := events.Items
	if typeFilter != "" {
		filtered := []map[string]interface{}{}
		for _, event := range events.Items {
			if reason, ok := event["reason"].(string); ok {
				// 简单过滤：Warning 类型包含错误关键字
				if typeFilter == "Warning" {
					if strings.Contains(strings.ToLower(reason), "error") ||
						strings.Contains(strings.ToLower(reason), "failed") ||
						strings.Contains(strings.ToLower(reason), "warning") {
						filtered = append(filtered, event)
					}
				} else if typeFilter == "Normal" {
					if strings.Contains(strings.ToLower(reason), "started") ||
						strings.Contains(strings.ToLower(reason), "success") ||
						strings.Contains(strings.ToLower(reason), "created") {
						filtered = append(filtered, event)
					}
				}
			}
		}
		result = filtered
	}

	return result, nil
}

// GetPodLogs 获取特定 Pod 的日志
// namespace: 命名空间名称
// podName: Pod 名称
// container: 容器名称（可选）
// tailLines: 日志行数
func (s *MmctlService) GetPodLogs(ctx context.Context, namespace, podName, container string, tailLines int) (string, error) {
	if namespace == "" || podName == "" {
		return "", fmt.Errorf("namespace and podName are required")
	}

	args := []string{"logs", podName, "--namespace", namespace}
	if container != "" {
		args = append(args, "-c", container)
	}
	if tailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tailLines))
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Dir = s.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// ListPods 列出命名空间下的所有 Pod
func (s *MmctlService) ListPods(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	args := []string{"get", "pods", "--namespace", namespace, "-o", "json"}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Dir = s.workDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w, output: %s", err, string(output))
	}

	var pods struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(output, &pods); err != nil {
		return nil, fmt.Errorf("failed to parse pods: %w", err)
	}

	return pods.Items, nil
}
