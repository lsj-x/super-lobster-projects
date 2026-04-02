package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"modelmagic-deploy-console/backend/internal/logger"
	"go.uber.org/zap"
)

// Logging 请求日志中间件
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 脱敏敏感信息
		clientIP := c.ClientIP()
		
		// 记录日志
		logFields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error("请求失败", logFields...)
		} else if statusCode >= 400 {
			logger.Warn("请求错误", logFields...)
		} else {
			logger.Info("请求完成", logFields...)
		}
	}
}

// LoggingWithSkip 可跳过特定路径的日志中间件
func LoggingWithSkip(skipPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过
		skip := false
		for _, path := range skipPaths {
			if c.Request.URL.Path == path {
				skip = true
				break
			}
		}

		if skip {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logFields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
		}

		if statusCode >= 500 {
			logger.Error("请求失败", logFields...)
		} else if statusCode >= 400 {
			logger.Warn("请求错误", logFields...)
		} else {
			logger.Info("请求完成", logFields...)
		}
	}
}
