package handlers

import (
	"net/http"
	"modelmagic-deploy-console/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// SecretHandler 处理 Secret 相关的 API 请求
type SecretHandler struct {
	SecretManager *service.SecretManager
}

// NewSecretHandler 创建新的 SecretHandler 实例
func NewSecretHandler(sm *service.SecretManager) *SecretHandler {
	return &SecretHandler{
		SecretManager: sm,
	}
}

// CreateSecretRequest 创建 Secret 的请求体
type CreateSecretRequest struct {
	Name      string            `json:"name" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Data      map[string]string `json:"data" binding:"required,min=1"`
}

// UpdateSecretRequest 更新 Secret 的请求体
type UpdateSecretRequest struct {
	Name      string            `json:"name" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Data      map[string]string `json:"data" binding:"required,min=1"`
}

// CreateSecret 创建新的 Secret
// POST /api/v1/secrets
func (h *SecretHandler) CreateSecret(c *gin.Context) {
	var req CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// 验证命名空间格式
	if err := service.ValidateNamespace(req.Namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid namespace: " + err.Error(),
		})
		return
	}

	// 验证 Secret 名称
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Secret name is required",
		})
		return
	}

	// 创建 Secret
	err := h.SecretManager.CreateSecret(c.Request.Context(), req.Namespace, req.Name, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create secret: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Secret created successfully",
		"name":      req.Name,
		"namespace": req.Namespace,
	})
}

// GetSecret 获取单个 Secret
// GET /api/v1/secrets/:namespace/:name
func (h *SecretHandler) GetSecret(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and name are required",
		})
		return
	}

	secret, err := h.SecretManager.GetSecret(c.Request.Context(), namespace, name)
	if err != nil {
		if err == service.ErrSecretNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Secret not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get secret: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// UpdateSecret 更新现有 Secret
// PUT /api/v1/secrets/:namespace/:name
func (h *SecretHandler) UpdateSecret(c *gin.Context) {
	var req UpdateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// 使用 URL 路径参数优先
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace != "" && name != "" {
		req.Namespace = namespace
		req.Name = name
	}

	// 验证
	if req.Namespace == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and name are required",
		})
		return
	}

	// 更新 Secret
	err := h.SecretManager.UpdateSecret(c.Request.Context(), req.Namespace, req.Name, req.Data)
	if err != nil {
		if err == service.ErrSecretNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Secret not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update secret: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Secret updated successfully",
		"name":      req.Name,
		"namespace": req.Namespace,
	})
}

// DeleteSecret 删除 Secret
// DELETE /api/v1/secrets/:namespace/:name
func (h *SecretHandler) DeleteSecret(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and name are required",
		})
		return
	}

	err := h.SecretManager.DeleteSecret(c.Request.Context(), namespace, name)
	if err != nil {
		if err == service.ErrSecretNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Secret not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete secret: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Secret deleted successfully",
		"name":      name,
		"namespace": namespace,
	})
}

// ListSecrets 列出命名空间中的所有 Secret
// GET /api/v1/secrets?namespace=xxx
func (h *SecretHandler) ListSecrets(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace query parameter is required",
		})
		return
	}

	secrets, err := h.SecretManager.ListSecrets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list secrets: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace": namespace,
		"secrets":   secrets,
		"count":     len(secrets),
	})
}
