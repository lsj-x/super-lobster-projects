package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrSecretKeyMissing  = errors.New("SECRET_ENCRYPTION_KEY environment variable is not set")
	ErrSecretNotFound    = errors.New("secret not found")
	ErrInvalidSecretData = errors.New("invalid secret data")
)

// SecretManager 管理 Kubernetes Secret 的加密和解密
type SecretManager struct {
	clientset     *kubernetes.Clientset
	encryptionKey []byte
}

// NewSecretManager 创建新的 SecretManager 实例
func NewSecretManager(clientset *kubernetes.Clientset) (*SecretManager, error) {
	key := os.Getenv("SECRET_ENCRYPTION_KEY")
	if key == "" {
		// 开发模式下允许无加密密钥（不推荐生产环境）
		if os.Getenv("ENV") == "development" {
			key = "dev-secret-key-32-bytes-long!!"
		} else {
			return nil, ErrSecretKeyMissing
		}
	}

	return &SecretManager{
		clientset:     clientset,
		encryptionKey: []byte(key),
	}, nil
}

// Encrypt 使用 AES-256-GCM 加密数据
func (m *SecretManager) Encrypt(data string) (string, error) {
	// 准备数据
	plaintext := []byte(data)

	// 创建 cipher block
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建 GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 编码为 base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (m *SecretManager) Decrypt(encrypted string) (string, error) {
	// 解码 base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 创建 cipher block
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建 GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// 分离 nonce 和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// SecretData 表示一个 Secret 的数据
type SecretData struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// CreateSecret 创建新的 Secret
func (m *SecretManager) CreateSecret(ctx context.Context, namespace, name string, data map[string]string) error {
	// 加密所有数据值
	encryptedData := make(map[string][]byte)
	for key, value := range data {
		encrypted, err := m.Encrypt(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt value for key '%s': %w", key, err)
		}
		encryptedData[key] = []byte(encrypted)
	}

	// 创建 Kubernetes Secret 对象
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "modelmagic-deploy-console",
			},
		},
		Data: encryptedData,
		Type: corev1.SecretTypeOpaque,
	}

	// 调用 K8s API 创建 Secret
	_, err := m.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return fmt.Errorf("secret '%s' already exists in namespace '%s'", name, namespace)
		}
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetSecret 获取 Secret（解密后的值）
func (m *SecretManager) GetSecret(ctx context.Context, namespace, name string) (*SecretData, error) {
	// 从 K8s 获取 Secret
	k8sSecret, err := m.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// 解密数据
	decryptedData := make(map[string]string)
	for key, value := range k8sSecret.Data {
		decrypted, err := m.Decrypt(string(value))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt value for key '%s': %w", key, err)
		}
		decryptedData[key] = decrypted
	}

	return &SecretData{
		Name:      k8sSecret.Name,
		Namespace: k8sSecret.Namespace,
		Data:      decryptedData,
		CreatedAt: k8sSecret.CreationTimestamp.String(),
		UpdatedAt: k8sSecret.CreationTimestamp.String(), // 简化处理
	}, nil
}

// UpdateSecret 更新 Secret
func (m *SecretManager) UpdateSecret(ctx context.Context, namespace, name string, data map[string]string) error {
	// 先获取现有 Secret
	existing, err := m.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return ErrSecretNotFound
		}
		return fmt.Errorf("failed to get existing secret: %w", err)
	}

	// 加密新数据
	encryptedData := make(map[string][]byte)
	for key, value := range data {
		encrypted, err := m.Encrypt(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt value for key '%s': %w", key, err)
		}
		encryptedData[key] = []byte(encrypted)
	}

	// 更新 Secret
	existing.Data = encryptedData
	_, err = m.clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteSecret 删除 Secret
func (m *SecretManager) DeleteSecret(ctx context.Context, namespace, name string) error {
	err := m.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return ErrSecretNotFound
		}
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// ListSecrets 列出命名空间中的所有 Secret（仅名称）
func (m *SecretManager) ListSecrets(ctx context.Context, namespace string) ([]string, error) {
	secrets, err := m.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "managed-by=modelmagic-deploy-console",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	names := make([]string, len(secrets.Items))
	for i, secret := range secrets.Items {
		names[i] = secret.Name
	}

	return names, nil
}
