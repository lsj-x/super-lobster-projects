package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"modelmagic-deploy-console/backend/internal/api"
	"modelmagic-deploy-console/backend/internal/config"
	"modelmagic-deploy-console/backend/internal/logger"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("加载配置文件失败：%v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Output); err != nil {
		fmt.Printf("初始化日志失败：%v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("模法师部署控制台启动", zap.String("config", *configPath))

	// 设置 Gin 模式
	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()

	// 设置 API 路由
	apiSvc := api.NewAPI(cfg)
	apiSvc.SetupRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("服务器启动", zap.String("addr", addr))

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			logger.Error("服务器启动失败", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("服务器关闭中...")
}
