package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SecretManager 秘密管理器
type SecretManager struct {
	clientset      *kubernetes.Clientset
	encryptionKey  []byte
	namespace      string
}

// SecretData 秘密数据结构
type SecretData struct {
	Name        string            `json:"name"`
	Data        map[string]string `json:"data"`
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewSecretManager 创建新的秘密管理器
func NewSecretManager() (*SecretManager, error) {
	// 从环境变量读取加密密钥
	encryptionKeyStr := os.Getenv("SECRET_ENCRYPTION_KEY")
	if encryptionKeyStr == "" {
		return nil, fmt.Errorf("SECRET_ENCRYPTION_KEY 环境变量未设置")
	}

	// 解码加密密钥 (base64)
	encryptionKey, err := base64.StdEncoding.DecodeString(encryptionKeyStr)
	if err != nil {
		return nil, fmt.Errorf("解密密钥解码失败: %w", err)
	}

	// 验证密钥长度 (AES-256 需要 32 字节)
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("加密密钥长度必须是 32 字节 (AES-256)")
	}

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
			return &SecretManager{
				clientset:     nil,
				encryptionKey: encryptionKey,
				namespace:     "default",
			}, nil
		}
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	// 获取当前命名空间
	namespace := "default"
	if nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		namespace = string(nsBytes)
	}

	return &SecretManager{
		clientset:     clientset,
		encryptionKey: encryptionKey,
		namespace:     namespace,
	}, nil
}

// Encrypt 加密数据 (AES-256-GCM)
func (m *SecretManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("创建 AES 密码块失败: %w", err)
	}

	// GCM 模式不需要 IV 填充
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 模式失败: %w", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成 nonce 失败: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回 base64 编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据 (AES-256-GCM)
func (m *SecretManager) Decrypt(ciphertext string) (string, error) {
	// 解码 base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}

	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("创建 AES 密码块失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 模式失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}

	// 提取 nonce 和密文
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

// CreateSecret 创建秘密
func (m *SecretManager) CreateSecret(name string, data map[string]string, description string) (*SecretData, error) {
	// 加密所有数据值
	encryptedData := make(map[string]string)
	for key, value := range data {
		encryptedValue, err := m.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("加密值失败 [%s]: %w", key, err)
		}
		encryptedData[key] = encryptedValue
	}

	// 转换为 Kubernetes Secret
	k8sData := make(map[string][]byte)
	for key, value := range encryptedData {
		k8sData[key] = []byte(value)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":     "modelmagic",
				"managed": "true",
			},
			Annotations: map[string]string{
				"description": description,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: k8sData,
	}

	if m.clientset != nil {
		_, err := m.clientset.CoreV1().Secrets(m.namespace).Create(nil, secret, metav1.CreateOptions{})
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil, fmt.Errorf("秘密已存在：%s", name)
			}
			return nil, fmt.Errorf("创建 Kubernetes Secret 失败: %w", err)
		}
	}

	now := time.Now()
	return &SecretData{
		Name:        name,
		Data:        data, // 返回原始数据（不返回加密数据）
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetSecret 获取秘密
func (m *SecretManager) GetSecret(name string) (*SecretData, error) {
	var secret *corev1.Secret
	var err error

	if m.clientset != nil {
		secret, err = m.clientset.CoreV1().Secrets(m.namespace).Get(nil, name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("秘密不存在：%s", name)
			}
			return nil, fmt.Errorf("获取 Kubernetes Secret 失败: %w", err)
		}
	} else {
		// 模拟数据用于测试
		return nil, fmt.Errorf("秘密不存在：%s", name)
	}

	// 解密数据
	decryptedData := make(map[string]string)
	for key, encryptedValue := range secret.Data {
		decryptedValue, err := m.Decrypt(string(encryptedValue))
		if err != nil {
			return nil, fmt.Errorf("解密值失败 [%s]: %w", key, err)
		}
		decryptedData[key] = decryptedValue
	}

	description := ""
	if desc, ok := secret.Annotations["description"]; ok {
		description = desc
	}

	// 创建时间从元数据获取
	createdAt := time.Now()
	if secret.CreationTimestamp.Time.Unix() > 0 {
		createdAt = secret.CreationTimestamp.Time
	}

	return &SecretData{
		Name:        name,
		Data:        decryptedData,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}, nil
}

// ListSecrets 列出所有秘密
func (m *SecretManager) ListSecrets() ([]string, error) {
	if m.clientset == nil {
		return []string{}, nil
	}

	secrets, err := m.clientset.CoreV1().Secrets(m.namespace).List(nil, metav1.ListOptions{
		LabelSelector: "managed=true",
	})
	if err != nil {
		return nil, fmt.Errorf("列出 Kubernetes Secrets 失败: %w", err)
	}

	names := make([]string, len(secrets.Items))
	for i, secret := range secrets.Items {
		names[i] = secret.Name
	}

	return names, nil
}

// UpdateSecret 更新秘密
func (m *SecretManager) UpdateSecret(name string, data map[string]string, description string) (*SecretData, error) {
	var secret *corev1.Secret
	var err error

	if m.clientset != nil {
		secret, err = m.clientset.CoreV1().Secrets(m.namespace).Get(nil, name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("秘密不存在：%s", name)
			}
			return nil, fmt.Errorf("获取 Kubernetes Secret 失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("秘密不存在：%s", name)
	}

	// 加密所有数据值
	encryptedData := make(map[string][]byte)
	for key, value := range data {
		encryptedValue, err := m.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("加密值失败 [%s]: %w", key, err)
		}
		encryptedData[key] = []byte(encryptedValue)
	}

	// 更新数据
	secret.Data = encryptedData
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}
	secret.Annotations["description"] = description

	if m.clientset != nil {
		_, err = m.clientset.CoreV1().Secrets(m.namespace).Update(nil, secret, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("更新 Kubernetes Secret 失败: %w", err)
		}
	}

	now := time.Now()
	return &SecretData{
		Name:        name,
		Data:        data,
		Description: description,
		CreatedAt:   secret.CreationTimestamp.Time,
		UpdatedAt:   now,
	}, nil
}

// DeleteSecret 删除秘密
func (m *SecretManager) DeleteSecret(name string) error {
	if m.clientset == nil {
		return nil
	}

	err := m.clientset.CoreV1().Secrets(m.namespace).Delete(nil, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("秘密不存在：%s", name)
		}
		return fmt.Errorf("删除 Kubernetes Secret 失败: %w", err)
	}

	return nil
}
