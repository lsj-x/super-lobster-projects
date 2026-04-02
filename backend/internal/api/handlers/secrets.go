package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"modelmagic-deploy-console/backend/internal/service"
)

// SecretHandlers Secret API 处理器
type SecretHandlers struct {
	secretMgr *service.SecretManager
}

// NewSecretHandlers 创建 Secret 处理器
func NewSecretHandlers(secretMgr *service.SecretManager) *SecretHandlers {
	return &SecretHandlers{
		secretMgr: secretMgr,
	}
}

// CreateSecretRequest 创建秘密请求
type CreateSecretRequest struct {
	Name        string            `json:"name" binding:"required"`
	Data        map[string]string `json:"data" binding:"required"`
	Description string            `json:"description"`
}

// UpdateSecretRequest 更新秘密请求
type UpdateSecretRequest struct {
	Data        map[string]string `json:"data" binding:"required"`
	Description string            `json:"description"`
}

// CreateSecret 创建秘密
// @Summary 创建秘密
// @Description 创建一个加密的秘密
// @Tags secrets
// @Accept json
// @Produce json
// @Param secret body CreateSecretRequest true "秘密数据"
// @Success 201 {object} service.SecretData
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/secrets [post]
func (h *SecretHandlers) CreateSecret(c *gin.Context) {
	var req CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证名称
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}

	if len(req.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据不能为空"})
		return
	}

	secret, err := h.secretMgr.CreateSecret(req.Name, req.Data, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, secret)
}

// GetSecret 获取秘密
// @Summary 获取秘密
// @Description 获取指定名称的秘密
// @Tags secrets
// @Produce json
// @Param name path string true "秘密名称"
// @Success 200 {object} service.SecretData
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/secrets/{name} [get]
func (h *SecretHandlers) GetSecret(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}

	secret, err := h.secretMgr.GetSecret(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// ListSecrets 列出所有秘密
// @Summary 列出所有秘密
// @Description 获取所有秘密的列表
// @Tags secrets
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]string
// @Router /api/v1/secrets [get]
func (h *SecretHandlers) ListSecrets(c *gin.Context) {
	names, err := h.secretMgr.ListSecrets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secrets": names})
}

// UpdateSecret 更新秘密
// @Summary 更新秘密
// @Description 更新指定名称的秘密
// @Tags secrets
// @Accept json
// @Produce json
// @Param name path string true "秘密名称"
// @Param secret body UpdateSecretRequest true "更新数据"
// @Success 200 {object} service.SecretData
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/secrets/{name} [put]
func (h *SecretHandlers) UpdateSecret(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}

	var req UpdateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据不能为空"})
		return
	}

	secret, err := h.secretMgr.UpdateSecret(name, req.Data, req.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// DeleteSecret 删除秘密
// @Summary 删除秘密
// @Description 删除指定名称的秘密
// @Tags secrets
// @Produce json
// @Param name path string true "秘密名称"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/secrets/{name} [delete]
func (h *SecretHandlers) DeleteSecret(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}

	err := h.secretMgr.DeleteSecret(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "秘密已删除", "name": name})
}
