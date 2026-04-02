package service

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ConfigMapManager ConfigMap 管理器
type ConfigMapManager struct {
	clientset *kubernetes.Clientset
	namespace string
}

// ConfigMapData ConfigMap 数据结构
type ConfigMapData struct {
	Name        string            `json:"name"`
	Data        map[string]string `json:"data"`
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewConfigMapManager 创建新的 ConfigMap 管理器
func NewConfigMapManager() (*ConfigMapManager, error) {
	// 获取 Kubernetes 客户端配置
	config, err := rest.InClusterConfig()
	if err != nil {
		// 如果不在集群中，尝试使用本地 kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = rest.InClusterConfig()
		if err != nil {
			// 创建模拟管理器用于测试
			return &ConfigMapManager{
				clientset: nil,
				namespace: "default",
			}, nil
		}
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 客户端失败：%w", err)
	}

	// 获取当前命名空间
	namespace := "default"
	if nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		namespace = string(nsBytes)
	}

	return &ConfigMapManager{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// CreateConfigMap 创建 ConfigMap
func (m *ConfigMapManager) CreateConfigMap(name string, data map[string]string, description string) (*ConfigMapData, error) {
	// 转换为 Kubernetes ConfigMap
	k8sData := make(map[string]string)
	for key, value := range data {
		k8sData[key] = value
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":      "modelmagic",
				"managed":  "true",
			},
			Annotations: map[string]string{
				"description": description,
			},
		},
		Data: k8sData,
	}

	if m.clientset != nil {
		_, err := m.clientset.CoreV1().ConfigMaps(m.namespace).Create(nil, configMap, metav1.CreateOptions{})
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil, fmt.Errorf("ConfigMap 已存在：%s", name)
			}
			return nil, fmt.Errorf("创建 Kubernetes ConfigMap 失败：%w", err)
		}
	}

	now := time.Now()
	return &ConfigMapData{
		Name:        name,
		Data:        data,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetConfigMap 获取 ConfigMap
func (m *ConfigMapManager) GetConfigMap(name string) (*ConfigMapData, error) {
	var configMap *corev1.ConfigMap
	var err error

	if m.clientset != nil {
		configMap, err = m.clientset.CoreV1().ConfigMaps(m.namespace).Get(nil, name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("ConfigMap 不存在：%s", name)
			}
			return nil, fmt.Errorf("获取 Kubernetes ConfigMap 失败：%w", err)
		}
	} else {
		// 模拟数据用于测试
		return nil, fmt.Errorf("ConfigMap 不存在：%s", name)
	}

	// 从元数据获取描述
	description := ""
	if desc, ok := configMap.Annotations["description"]; ok {
		description = desc
	}

	// 创建时间从元数据获取
	createdAt := time.Now()
	if configMap.CreationTimestamp.Time.Unix() > 0 {
		createdAt = configMap.CreationTimestamp.Time
	}

	return &ConfigMapData{
		Name:        name,
		Data:        configMap.Data,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}, nil
}

// ListConfigMaps 列出所有 ConfigMap
func (m *ConfigMapManager) ListConfigMaps() ([]string, error) {
	if m.clientset == nil {
		return []string{}, nil
	}

	configMaps, err := m.clientset.CoreV1().ConfigMaps(m.namespace).List(nil, metav1.ListOptions{
		LabelSelector: "managed=true",
	})
	if err != nil {
		return nil, fmt.Errorf("列出 Kubernetes ConfigMaps 失败：%w", err)
	}

	names := make([]string, len(configMaps.Items))
	for i, cm := range configMaps.Items {
		names[i] = cm.Name
	}
	return names, nil
}

// UpdateConfigMap 更新 ConfigMap
func (m *ConfigMapManager) UpdateConfigMap(name string, data map[string]string, description string) (*ConfigMapData, error) {
	var configMap *corev1.ConfigMap
	var err error

	if m.clientset != nil {
		configMap, err = m.clientset.CoreV1().ConfigMaps(m.namespace).Get(nil, name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("ConfigMap 不存在：%s", name)
			}
			return nil, fmt.Errorf("获取 Kubernetes ConfigMap 失败：%w", err)
		}
	} else {
		return nil, fmt.Errorf("ConfigMap 不存在：%s", name)
	}

	// 更新数据
	configMap.Data = data
	if configMap.Annotations == nil {
		configMap.Annotations = make(map[string]string)
	}
	configMap.Annotations["description"] = description

	if m.clientset != nil {
		_, err = m.clientset.CoreV1().ConfigMaps(m.namespace).Update(nil, configMap, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("更新 Kubernetes ConfigMap 失败：%w", err)
		}
	}

	now := time.Now()
	return &ConfigMapData{
		Name:        name,
		Data:        data,
		Description: description,
		CreatedAt:   configMap.CreationTimestamp.Time,
		UpdatedAt:   now,
	}, nil
}

// DeleteConfigMap 删除 ConfigMap
func (m *ConfigMapManager) DeleteConfigMap(name string) error {
	if m.clientset == nil {
		return nil
	}

	err := m.clientset.CoreV1().ConfigMaps(m.namespace).Delete(nil, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("ConfigMap 不存在：%s", name)
		}
		return fmt.Errorf("删除 Kubernetes ConfigMap 失败：%w", err)
	}
	return nil
}
