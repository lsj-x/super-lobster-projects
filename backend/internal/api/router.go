package api

import (
	"net/http"
	"modelmagic-deploy-console/backend/internal/config"
	"modelmagic-deploy-console/backend/internal/logger"
	"modelmagic-deploy-console/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	mmctlSvc    *service.MmctlService
	secretSvc   *service.SecretManager
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应限制
		},
	}
)

type API struct {
	cfg *config.Config
}

func NewAPI(cfg *config.Config) *API {
	mmctlSvc = service.NewMmctlService(
		cfg.Mmctl.BaseDir,
		cfg.Mmctl.ScriptPath,
		cfg.Mmctl.WorkDir,
	)
	
	// 初始化 SecretManager（可能失败）
	var err error
	secretSvc, err = service.NewSecretManager(nil)
	if err != nil {
		logger.Warn("Failed to initialize SecretManager", zap.Error(err))
	}
	
	return &API{cfg: cfg}
}

// SetupRoutes 设置路由
func (a *API) SetupRoutes(r *gin.Engine) {
	r.GET("/health", a.healthCheck)
	
	api := r.Group("/api/v1")
	{
		// 命名空间管理
		api.GET("/namespaces", a.listNamespaces)
		api.GET("/namespaces/:name", a.getNamespace)
		api.POST("/namespaces", a.createNamespace)
		api.PUT("/namespaces/:name", a.updateNamespace)
		api.DELETE("/namespaces/:name", a.deleteNamespace)
		
		// 安装/卸载/升级/回滚
		api.POST("/install", a.startInstall)
		api.POST("/uninstall", a.startUninstall)
		api.POST("/rollback", a.startRollback)
		api.POST("/upgrade", a.startUpgrade)
		api.GET("/access/:name", a.getAccess)
		
		// Secret 管理
		api.GET("/secrets", a.listSecretsHandler)
		api.GET("/secrets/:namespace/:name", a.getSecretHandler)
		api.POST("/secrets", a.createSecretHandler)
		api.PUT("/secrets/:namespace/:name", a.updateSecretHandler)
		api.DELETE("/secrets/:namespace/:name", a.deleteSecretHandler)
		
		// 实时日志 WebSocket
		api.GET("/logs/ws", a.websocketLogs)
	}
}

func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============ 命名空间管理 ============

func (a *API) listNamespaces(c *gin.Context) {
	namespaces, err := mmctlSvc.ListNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
}

func (a *API) getNamespace(c *gin.Context) {
	name := c.Param("name")
	config, err := mmctlSvc.GetNamespaceConfig(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": config})
}

func (a *API) createNamespace(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := mmctlSvc.NewNamespace(req.Namespace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message":   "命名空间配置已创建",
		"namespace": req.Namespace,
	})
}

func (a *API) updateNamespace(c *gin.Context) {
	name := c.Param("name")
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := mmctlSvc.UpdateNamespaceConfig(name, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "配置已更新",
		"namespace": name,
	})
}

func (a *API) deleteNamespace(c *gin.Context) {
	name := c.Param("name")
	c.JSON(http.StatusOK, gin.H{
		"message":   "命名空间删除请求已接收",
		"namespace": name,
	})
}

// ============ 安装/卸载/升级/回滚 ============

func (a *API) startInstall(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		Step      int    `json:"step"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 异步执行安装
	go func() {
		err := mmctlSvc.Install(req.Namespace, req.Step, func(line string) {
			logger.Info("Install log", zap.String("line", line))
		})
		if err != nil {
			logger.Error("安装失败", zap.Error(err))
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "安装任务已启动",
		"namespace": req.Namespace,
		"step":      req.Step,
	})
}

func (a *API) startUninstall(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		Step      int    `json:"step"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := mmctlSvc.Uninstall(req.Namespace, req.Step, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "卸载完成",
		"namespace": req.Namespace,
	})
}

func (a *API) startRollback(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		Version   string `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 异步执行回滚
	go func() {
		err := mmctlSvc.Rollback(req.Namespace, req.Version, func(line string) {
			logger.Info("Rollback log", zap.String("line", line))
		})
		if err != nil {
			logger.Error("回滚失败", zap.Error(err))
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "回滚已启动",
		"namespace": req.Namespace,
		"version":   req.Version,
	})
}

func (a *API) startUpgrade(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		Version   string `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 异步执行升级
	go func() {
		err := mmctlSvc.Upgrade(req.Namespace, req.Version, func(line string) {
			logger.Info("Upgrade log", zap.String("line", line))
		})
		if err != nil {
			logger.Error("升级失败", zap.Error(err))
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "升级已启动",
		"namespace": req.Namespace,
		"version":   req.Version,
	})
}

func (a *API) getAccess(c *gin.Context) {
	name := c.Param("name")
	url, err := mmctlSvc.GetSystemAccess(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// ============ Secret 管理 ============

func (a *API) listSecretsHandler(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace query parameter is required"})
		return
	}
	
	if secretSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SecretManager not initialized"})
		return
	}
	
	secrets, err := secretSvc.ListSecrets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"namespace": namespace,
		"secrets":   secrets,
		"count":     len(secrets),
	})
}

func (a *API) getSecretHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	
	if secretSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SecretManager not initialized"})
		return
	}
	
	secret, err := secretSvc.GetSecret(c.Request.Context(), namespace, name)
	if err != nil {
		if err == service.ErrSecretNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, secret)
}

func (a *API) createSecretHandler(c *gin.Context) {
	var req struct {
		Name      string            `json:"name" binding:"required"`
		Namespace string            `json:"namespace" binding:"required"`
		Data      map[string]string `json:"data" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if secretSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SecretManager not initialized"})
		return
	}
	
	if err := service.ValidateNamespace(req.Namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid namespace: " + err.Error()})
		return
	}
	
	err := secretSvc.CreateSecret(c.Request.Context(), req.Namespace, req.Name, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Secret created successfully",
		"name":      req.Name,
		"namespace": req.Namespace,
	})
}

func (a *API) updateSecretHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	
	var req struct {
		Data map[string]string `json:"data" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if secretSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SecretManager not initialized"})
		return
	}
	
	err := secretSvc.UpdateSecret(c.Request.Context(), namespace, name, req.Data)
	if err != nil {
		if err == service.ErrSecretNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "Secret updated successfully",
		"name":      name,
		"namespace": namespace,
	})
}

func (a *API) deleteSecretHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	
	if secretSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SecretManager not initialized"})
		return
	}
	
	err := secretSvc.DeleteSecret(c.Request.Context(), namespace, name)
	if err != nil {
		if err == service.ErrSecretNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "Secret deleted successfully",
		"name":      name,
		"namespace": namespace,
	})
}

// ============ WebSocket 日志 ============

func (a *API) websocketLogs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}
	defer conn.Close()
	
	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// 实际实现需要订阅日志通道并推送
	}
}
