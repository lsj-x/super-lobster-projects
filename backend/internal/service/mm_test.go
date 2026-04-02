package service

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewMMService(t *testing.T) {
	// 测试默认配置
	os.Setenv("MM_API_BASE_URL", "http://test-api.example.com")
	os.Setenv("MM_API_KEY", "test-api-key")

	service := NewMMService()

	if service.baseURL != "http://test-api.example.com" {
		t.Errorf("Expected baseURL to be 'http://test-api.example.com', got '%s'", service.baseURL)
	}

	if service.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got '%s'", service.apiKey)
	}

	if service.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", service.httpClient.Timeout)
	}
}

func TestNewMMServiceDefaults(t *testing.T) {
	// 测试默认值
	os.Unsetenv("MM_API_BASE_URL")
	os.Unsetenv("MM_API_KEY")

	service := NewMMService()

	if service.baseURL != "https://api.modelmagic.example.com" {
		t.Errorf("Expected default baseURL, got '%s'", service.baseURL)
	}

	if service.apiKey != "" {
		t.Errorf("Expected empty apiKey by default, got '%s'", service.apiKey)
	}
}

func TestListNamespaces(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	namespaces, err := service.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}

	if len(namespaces) == 0 {
		t.Error("Expected at least one namespace")
	}

	// 验证命名空间结构
	for _, ns := range namespaces {
		if ns.Name == "" {
			t.Error("Namespace name should not be empty")
		}
		if ns.CreatedAt.IsZero() {
			t.Error("Namespace CreatedAt should not be zero")
		}
	}
}

func TestDeploy(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	deployment, err := service.Deploy(ctx, "production", "test-model", "v1.0")
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	if deployment.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got '%s'", deployment.Namespace)
	}

	if deployment.ModelName != "test-model" {
		t.Errorf("Expected model_name 'test-model', got '%s'", deployment.ModelName)
	}

	if deployment.Version != "v1.0" {
		t.Errorf("Expected version 'v1.0', got '%s'", deployment.Version)
	}

	if deployment.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", deployment.Status)
	}
}

func TestGetStatusNotFound(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	_, err := service.GetStatus(ctx, "nonexistent", "model")
	if err == nil {
		t.Error("Expected error for non-existent deployment")
	}
}

func TestListDeployments(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	// 先创建一个部署
	service.Deploy(ctx, "production", "test-model", "v1.0")

	// 列出所有部署
	deployments, err := service.ListDeployments(ctx, "")
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}

	if len(deployments) == 0 {
		t.Error("Expected at least one deployment")
	}

	// 按命名空间过滤
	prodDeployments, err := service.ListDeployments(ctx, "production")
	if err != nil {
		t.Fatalf("ListDeployments with namespace failed: %v", err)
	}

	for _, dep := range prodDeployments {
		if dep.Namespace != "production" {
			t.Errorf("Expected all deployments to be in 'production', got '%s'", dep.Namespace)
		}
	}
}

func TestScale(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	// 先创建一个部署
	service.Deploy(ctx, "production", "test-model", "v1.0")

	// 扩容
	err := service.Scale(ctx, "production", "test-model", 3)
	if err != nil {
		t.Fatalf("Scale failed: %v", err)
	}

	// 验证扩容结果
	status, err := service.GetStatus(ctx, "production", "test-model")
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.Replicas != 3 {
		t.Errorf("Expected replicas to be 3, got %d", status.Replicas)
	}
}

func TestScaleNotFound(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	err := service.Scale(ctx, "nonexistent", "model", 2)
	if err == nil {
		t.Error("Expected error for scaling non-existent deployment")
	}
}

func TestDelete(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	// 先创建一个部署
	service.Deploy(ctx, "production", "test-model", "v1.0")

	// 删除
	err := service.Delete(ctx, "production", "test-model")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证删除结果
	_, err = service.GetStatus(ctx, "production", "test-model")
	if err == nil {
		t.Error("Expected error after deletion")
	}
}

func TestDeleteNotFound(t *testing.T) {
	service := NewMMService()
	ctx := context.Background()

	err := service.Delete(ctx, "nonexistent", "model")
	if err == nil {
		t.Error("Expected error for deleting non-existent deployment")
	}
}
