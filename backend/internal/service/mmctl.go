package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"modelmagic-deploy-console/backend/internal/logger"

	"go.uber.org/zap"
)

type MmctlService struct {
	baseDir    string
	scriptPath string
	workDir    string
}

func NewMmctlService(baseDir, scriptPath, workDir string) *MmctlService {
	return &MmctlService{
		baseDir:    baseDir,
		scriptPath: scriptPath,
		workDir:    workDir,
	}
}

// ExtractPackages 解压安装包
func (s *MmctlService) ExtractPackages() error {
	logger.Info("开始解压安装包...")
	cmd := exec.Command("bash", s.scriptPath, "01", "extract")
	cmd.Dir = s.baseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("解压失败", zap.Error(err), zap.String("output", string(output)))
		return fmt.Errorf("解压失败: %w", err)
	}
	logger.Info("解压成功", zap.String("output", string(output)))
	return nil
}

// NewNamespace 生成新命名空间配置
func (s *MmctlService) NewNamespace(namespace string) error {
	logger.Info("生成命名空间配置", zap.String("namespace", namespace))
	cmd := exec.Command("bash", s.scriptPath, "02", "new", namespace)
	cmd.Dir = s.baseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("生成配置失败", zap.Error(err), zap.String("output", string(output)))
		return fmt.Errorf("生成配置失败: %w", err)
	}
	logger.Info("配置生成成功", zap.String("output", string(output)))
	return nil
}

// CheckEnvironment 检查环境
func (s *MmctlService) CheckEnvironment(namespace string) error {
	logger.Info("检查环境", zap.String("namespace", namespace))
	cmd := exec.Command("bash", s.scriptPath, "03", "check", namespace)
	cmd.Dir = s.baseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("检查失败", zap.Error(err), zap.String("output", string(output)))
		return fmt.Errorf("检查失败: %w", err)
	}
	logger.Info("检查通过", zap.String("output", string(output)))
	return nil
}

// Install 执行安装
func (s *MmctlService) Install(namespace string, step int, logCallback func(string)) error {
	logger.Info("开始安装", zap.String("namespace", namespace), zap.Int("step", step))
	
	args := []string{s.scriptPath, "04", "install", namespace}
	if step > 0 {
		args = append(args, fmt.Sprintf("%d", step))
	}
	
	cmd := exec.Command("bash", args...)
	cmd.Dir = s.baseDir
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建 stdout pipe 失败: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr pipe 失败: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}
	
	// 读取输出
	go s.readOutput(stdout, logCallback)
	go s.readOutput(stderr, logCallback)
	
	if err := cmd.Wait(); err != nil {
		logger.Error("安装失败", zap.Error(err))
		return fmt.Errorf("安装失败: %w", err)
	}
	
	logger.Info("安装成功")
	return nil
}

func (s *MmctlService) readOutput(r io.Reader, callback func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if callback != nil {
			callback(line)
		}
		logger.Info("mmctl", zap.String("line", line))
	}
}

// Uninstall 执行卸载
func (s *MmctlService) Uninstall(namespace string, step int) error {
	logger.Info("开始卸载", zap.String("namespace", namespace), zap.Int("step", step))
	
	args := []string{s.scriptPath, "88", "uninstall", namespace}
	if step > 0 {
		args = append(args, fmt.Sprintf("%d", step))
	}
	
	cmd := exec.Command("bash", args...)
	cmd.Dir = s.baseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("卸载失败", zap.Error(err), zap.String("output", string(output)))
		return fmt.Errorf("卸载失败: %w", err)
	}
	logger.Info("卸载成功", zap.String("output", string(output)))
	return nil
}

// GetSystemAccess 获取系统访问地址
func (s *MmctlService) GetSystemAccess(namespace string) (string, error) {
	logger.Info("获取系统访问地址", zap.String("namespace", namespace))
	cmd := exec.Command("bash", s.scriptPath, "05", "access", namespace)
	cmd.Dir = s.baseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("获取地址失败: %w", err)
	}
	
	// 解析输出中的 URL
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "http://") {
			return strings.TrimSpace(line), nil
		}
	}
	
	return "", fmt.Errorf("未找到访问地址")
}

// ListNamespaces 列出所有命名空间配置
func (s *MmctlService) ListNamespaces() ([]string, error) {
	namespaces := []string{}
	
	// 扫描 workDir 下的目录
	entries, err := os.ReadDir(s.workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return namespaces, nil
		}
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否有 hosts 和 main.yml 文件
			hostsPath := filepath.Join(s.workDir, entry.Name(), "hosts")
			mainYmlPath := filepath.Join(s.workDir, entry.Name(), "main.yml")
			
			if _, err := os.Stat(hostsPath); err == nil {
				if _, err := os.Stat(mainYmlPath); err == nil {
					namespaces = append(namespaces, entry.Name())
				}
			}
		}
	}
	
	return namespaces, nil
}

// GetNamespaceConfig 获取命名空间配置详情
func (s *MmctlService) GetNamespaceConfig(namespace string) (map[string]string, error) {
	configPath := filepath.Join(s.workDir, namespace, "main.yml")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	
	// 简单解析 YAML 为 map (实际项目中建议使用 yaml 库)
	config := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				config[key] = value
			}
		}
	}
	
	return config, nil
}

// UpdateNamespaceConfig 更新命名空间配置
func (s *MmctlService) UpdateNamespaceConfig(namespace string, updates map[string]string) error {
	configPath := filepath.Join(s.workDir, namespace, "main.yml")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}
	
	lines := strings.Split(string(data), "\n")
	newLines := make([]string, 0, len(lines))
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		updated := false
		for key, value := range updates {
			if strings.HasPrefix(trimmed, key+":") {
				// 保留缩进
				indent := ""
				for _, c := range line {
					if c == ' ' || c == '\t' {
						indent += string(c)
					} else {
						break
					}
				}
				newLines = append(newLines, indent+key+": "+value)
				updated = true
				break
			}
		}
		if !updated {
			newLines = append(newLines, line)
		}
	}
	
	err = os.WriteFile(configPath, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	
	logger.Info("配置更新成功", zap.String("namespace", namespace))
	return nil
}

// Rollback 回滚命名空间到指定版本
func (s *MmctlService) Rollback(namespace, version string, logCallback func(string)) error {
	logger.Info("开始回滚", zap.String("namespace", namespace), zap.String("version", version))
	
	args := []string{s.scriptPath, "90", namespace, version}
	cmd := exec.Command("bash", args...)
	cmd.Dir = s.baseDir
	
	cmd.Stdout = &logWriter{callback: logCallback}
	cmd.Stderr = &logWriter{callback: logCallback}
	
	err := cmd.Run()
	if err != nil {
		logger.Error("回滚失败", zap.Error(err))
		return fmt.Errorf("回滚失败：%w", err)
	}
	
	logger.Info("回滚成功", zap.String("namespace", namespace), zap.String("version", version))
	return nil
}

// logWriter 实现 io.Writer 用于日志回调
type logWriter struct {
	callback func(string)
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	if w.callback != nil {
		w.callback(string(p))
	}
	return len(p), nil
}
