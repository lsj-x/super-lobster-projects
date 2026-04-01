package api

import (
	"net/http"
	"modelmagic-deploy-console/internal/config"
	"modelmagic-deploy-console/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler 存储 API 处理器依赖
type Handler struct {
	Config    *config.Config
	MMService *service.MMService
}

// NewHandler 创建新的 API 处理器
func NewHandler(cfg *config.Config, mm *service.MMService) *Handler {
	return &Handler{
		Config:    cfg,
		MMService: mm,
	}
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"version": h.Config.Version,
	})
}

// ListNamespaces 列出所有命名空间
func (h *Handler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.MMService.ListNamespaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, namespaces)
}

// DeployModel 部署模型
func (h *Handler) DeployModel(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		ModelName string `json:"model_name" binding:"required"`
		Version   string `json:"version"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.MMService.Deploy(c.Request.Context(), req.Namespace, req.ModelName, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDeploymentStatus 获取部署状态
func (h *Handler) GetDeploymentStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	modelName := c.Param("model")

	status, err := h.MMService.GetStatus(c.Request.Context(), namespace, modelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListDeployments 列出所有部署
func (h *Handler) ListDeployments(c *gin.Context) {
	namespace := c.Query("namespace")
	
	deployments, err := h.MMService.ListDeployments(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// ScaleDeployment 扩缩容部署
func (h *Handler) ScaleDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	modelName := c.Param("model")
	
	var req struct {
		Replicas int `json:"replicas" binding:"required,min=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.MMService.Scale(c.Request.Context(), namespace, modelName, req.Replicas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "scale operation initiated",
		"namespace": namespace,
		"model": modelName,
		"replicas": req.Replicas,
	})
}

// DeleteDeployment 删除部署
func (h *Handler) DeleteDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	modelName := c.Param("model")

	err := h.MMService.Delete(c.Request.Context(), namespace, modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "deletion initiated",
		"namespace": namespace,
		"model": modelName,
	})
}
