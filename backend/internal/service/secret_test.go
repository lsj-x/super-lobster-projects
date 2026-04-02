package service

import (
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// 设置测试用的加密密钥 (32 字节)
	testKey := "thisis32bytekey12345678901234ab"
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGhpc2lzMzJieXRla2V5MTIzNDU2Nzg5MDEyM2Fi") // base64 编码

	// 创建 SecretManager
	mgr, err := NewSecretManager()
	if err != nil {
		t.Fatalf("创建 SecretManager 失败：%v", err)
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
	// 设置测试用的加密密钥
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGhpc2lzMzJieXRla2V5MTIzNDU2Nzg5MDEyM2Fi")

	mgr, err := NewSecretManager()
	if err != nil {
		t.Fatalf("创建 SecretManager 失败：%v", err)
	}

	testCases := []string{
		"简单文本",
		"包含特殊字符：!@#$%^&*()",
		"包含中文：你好，世界！",
		"包含 Emoji: 😀🎉🚀",
		"长文本：" + string(make([]byte, 1000)),
	}

	for i, tc := range testCases {
		t.Run(string(rune(i)), func(t *testing.T) {
			// 填充长文本测试
			if len(tc) > 10 {
				tc = tc[:10] + "..."
			}

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
	// 测试无效的密钥长度
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGk=") // 只有 2 字节

	_, err := NewSecretManager()
	if err == nil {
		t.Fatal("期望密钥长度错误，但得到了 nil")
	}

	t.Logf("正确捕获了错误：%v", err)
}

func TestSecretDataStructure(t *testing.T) {
	os.Setenv("SECRET_ENCRYPTION_KEY", "dGhpc2lzMzJieXRla2V5MTIzNDU2Nzg5MDEyM2Fi")

	mgr, err := NewSecretManager()
	if err != nil {
		t.Fatalf("创建 SecretManager 失败：%v", err)
	}

	// 测试创建秘密
	data := map[string]string{
		"username": "admin",
		"password": "secret123",
		"api_key":  "sk-1234567890abcdef",
	}

	secret, err := mgr.CreateSecret("test-secret", data, "测试秘密")
	if err != nil {
		t.Fatalf("创建秘密失败：%v", err)
	}

	if secret.Name != "test-secret" {
		t.Errorf("秘密名称不匹配：期望 test-secret, 得到 %s", secret.Name)
	}

	if secret.Description != "测试秘密" {
		t.Errorf("描述不匹配：期望 测试秘密，得到 %s", secret.Description)
	}

	if len(secret.Data) != 3 {
		t.Errorf("数据长度不匹配：期望 3, 得到 %d", len(secret.Data))
	}

	t.Logf("创建的秘密：%+v", secret)
}
