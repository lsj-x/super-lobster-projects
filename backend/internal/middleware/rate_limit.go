package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	MaxRequests   int           // 最大请求数
	TimeWindow    time.Duration // 时间窗口
}

// DefaultRateLimitConfig 默认速率限制配置 (60 次/分钟)
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests:  60,
		TimeWindow:   time.Minute,
	}
}

// clientRateLimit 单个客户端的速率限制状态
type clientRateLimit struct {
	count     int
	resetTime time.Time
}

// RateLimiter 速率限制器
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientRateLimit
	config   RateLimitConfig
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	limiter := &RateLimiter{
		clients: make(map[string]*clientRateLimit),
		config:  config,
	}
	// 启动清理协程
	go limiter.cleanup()
	return limiter
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// 获取或创建客户端限制状态
	client, exists := rl.clients[clientIP]
	if !exists {
		rl.clients[clientIP] = &clientRateLimit{
			count:     1,
			resetTime: now.Add(rl.config.TimeWindow),
		}
		return true
	}

	// 检查时间窗口是否过期
	if now.After(client.resetTime) {
		client.count = 1
		client.resetTime = now.Add(rl.config.TimeWindow)
		return true
	}

	// 检查请求次数是否超过限制
	if client.count >= rl.config.MaxRequests {
		return false
	}

	client.count++
	return true
}

// cleanup 定期清理过期的客户端记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.TimeWindow)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			if now.After(client.resetTime) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit 速率限制中间件
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(429, gin.H{
				"error":       "请求过于频繁，请稍后再试",
				"retry_after": int(config.TimeWindow.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithConfig 带配置的速率限制中间件
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(429, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
