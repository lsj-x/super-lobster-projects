package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"modelmagic-deploy-console/backend/internal/config"
	"modelmagic-deploy-console/backend/internal/logger"
	"modelmagic-deploy-console/backend/internal/service"
	"modelmagic-deploy-console/backend/internal/mmctl"

	"go.uber.org/zap"
)

var (
	mmctlSvc *mmctl.MmctlService
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应限制
		},
	}
)

type API struct {
	cfg *config.Config
}

func NewAPI(cfg *config.Config) *API {
	mmctlSvc = mmctl.NewMmctlService(
		cfg.Mmctl.BaseDir,
		cfg.Mmctl.ScriptPath,
		cfg.Mmctl.WorkDir,
	)
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
		
		// 安装/卸载
		api.POST("/install", a.startInstall)
		api.POST("/uninstall", a.startUninstall)
		api.GET("/access/:name", a.getAccess)
		
		// 实时日志 WebSocket
		api.GET("/logs/ws", a.websocketLogs)
	}
}

func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

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
	
	// 生成默认配置
	if err := mmctlSvc.NewNamespace(req.Namespace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "命名空间配置已创建", "namespace": req.Namespace})
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
	
	c.JSON(http.StatusOK, gin.H{"message": "配置已更新", "namespace": name})
}

func (a *API) deleteNamespace(c *gin.Context) {
	name := c.Param("name")
	// 这里应该先检查是否已安装，如果已安装则提示用户
	c.JSON(http.StatusOK, gin.H{"message": "命名空间删除请求已接收", "namespace": name})
}

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
			// 这里可以将日志推送到 WebSocket 客户端
			logger.Info("Install log", zap.String("line", line))
		})
		if err != nil {
			logger.Error("安装失败", zap.Error(err))
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{"message": "安装任务已启动", "namespace": req.Namespace, "step": req.Step})
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
	
	if err := mmctlSvc.Uninstall(req.Namespace, req.Step); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "卸载完成", "namespace": req.Namespace})
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

func (a *API) websocketLogs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}
	defer conn.Close()
	
	// 这里可以将实时日志推送到客户端
	// 实际实现需要维护一个日志通道
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// 发送日志
		// conn.WriteMessage(websocket.TextMessage, []byte(logLine))
	}
}
