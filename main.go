package main

import (
	"log"
	"net/http"

	"github.com/yourorg/mmctl-backend/cmd"
	"github.com/yourorg/mmctl-backend/internal/server"
)

func main() {
	// 启动 Cobra 命令行工具（mmctl.sh）
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	// 启动 Gin HTTP 服务（独立运行时）
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := server.StartHTTPServer(); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// 阻塞主 goroutine
	select {}
}

