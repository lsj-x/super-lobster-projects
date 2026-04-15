package service

import (
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// 设置测试用的加密密钥 (必须正好 32 字节用于 AES-256)
	testKey := "thisis32bytekey12345678901234567" // 正好 32 字符
	if len(testKey) != 32 {
		t.Fatalf("测试密钥长度必须是 32 字节，当前：%d", len(testKey))
	}
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGhpc2lzMzJieXRla2V5MTIzNDU2Nzg5MDEyMzQ1Njc=") // base64 编码

	// 创建 SecretManager (传入 nil clientset for unit test)
	mgr := &SecretManager{
		clientset:     nil,
		encryptionKey: []byte(testKey), // 使用 32 字节密钥
	}

	// 测试数据
	plaintext := "这是测试的敏感数据"

	// 测试加密
	ciphertext, err := mgr.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	if ciphertext == "" {
		t.Fatal("加密结果为空")
	}
	if ciphertext == plaintext {
		t.Fatal("加密结果与原文相同")
	}
	t.Logf("原文：%s", plaintext)
	t.Logf("密文：%s", ciphertext)

	// 测试解密
	decrypted, err := mgr.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("解密失败：%v", err)
	}
	if decrypted != plaintext {
		t.Errorf("解密结果不匹配：期望 %s, 得到 %s", plaintext, decrypted)
	}
	t.Logf("解密结果：%s", decrypted)
}

func TestEncryptDecryptMultiple(t *testing.T) {
	// 使用 32 字节密钥
	testKey := "thisis32bytekey12345678901234567"
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGhpc2lzMzJieXRla2V5MTIzNDU2Nzg5MDEyMzQ1Njc=")
	mgr := &SecretManager{
		clientset:     nil,
		encryptionKey: []byte(testKey),
	}

	testCases := []string{
		"简单文本",
		"包含特殊字符：!@#$%^&*()",
		"包含中文：你好，世界！",
		"包含 Emoji: 😀🎉🚀",
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			ciphertext, err := mgr.Encrypt(tc)
			if err != nil {
				t.Fatalf("加密失败：%v", err)
			}
			decrypted, err := mgr.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("解密失败：%v", err)
			}
			if decrypted != tc {
				t.Errorf("解密结果不匹配：期望 %s, 得到 %s", tc, decrypted)
			}
		})
	}
}

func TestInvalidKey(t *testing.T) {
	// 测试无效的密钥长度 - 当前实现不验证密钥长度，只检查是否为空
	// 在 development 模式下，如果 SECRET_ENCRYPTION_KEY 为空，会使用默认密钥
	os.Setenv("SECRET_ENCRYPTION_KEY", "")
	os.Setenv("ENV", "development")

	// 在 development 模式下，空密钥会使用默认值，不会报错
	mgr, err := NewSecretManager(nil)
	if err != nil {
		t.Fatalf("在 development 模式下，期望使用默认密钥，但得到错误：%v", err)
	}
	if mgr == nil {
		t.Fatal("期望创建 SecretManager，但得到 nil")
	}
	t.Logf("正确使用了默认密钥")

	// 在生产模式下，空密钥应该报错
	os.Unsetenv("ENV")
	_, err = NewSecretManager(nil)
	if err == nil {
		t.Fatal("期望生产模式下空密钥报错，但得到了 nil")
	}
	t.Logf("正确捕获了生产模式下的错误：%v", err)
}

func TestSecretDataStructure(t *testing.T) {
	testKey := "thisis32bytekey12345678901234567"
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGhpc2lzMzJieXRla2V5MTIzNDU2Nzg5MDEyMzQ1Njc=")
	mgr := &SecretManager{
		clientset:     nil,
		encryptionKey: []byte(testKey),
	}

	// 测试加密数据
	data := map[string]string{
		"username": "admin",
		"password": "secret123",
		"api_key":  "sk-1234567890abcdef",
	}

	// 测试加密每个字段
	for key, value := range data {
		encrypted, err := mgr.Encrypt(value)
		if err != nil {
			t.Fatalf("加密字段 %s 失败：%v", key, err)
		}
		if encrypted == value {
			t.Fatalf("字段 %s 加密后与原文相同", key)
		}

		// 测试解密
		decrypted, err := mgr.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("解密字段 %s 失败：%v", key, err)
		}
		if decrypted != value {
			t.Errorf("字段 %s 解密结果不匹配：期望 %s, 得到 %s", key, value, decrypted)
		}
	}
	t.Logf("所有字段加密解密成功")
}
