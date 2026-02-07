package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"url-shortener/internal/config"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) (bool, int)
	Reset(key string)
}

// MemoryRateLimiter 基于内存的限流器
type MemoryRateLimiter struct {
	limits map[string]*rateLimit
	mu     sync.RWMutex
	config *config.RateLimitConfig
}

// rateLimit 单个键的限流状态
type rateLimit struct {
	count     int
	windowEnd int64
}

// NewMemoryRateLimiter 创建新的内存限流器
func NewMemoryRateLimiter(cfg *config.RateLimitConfig) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		limits: make(map[string]*rateLimit),
		config: cfg,
	}
}

// Allow 检查是否允许请求
func (rl *MemoryRateLimiter) Allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().Unix()
	limit := rl.config.RequestsPerMinute

	// 检查或创建限流状态
	state, exists := rl.limits[key]
	if !exists {
		rl.limits[key] = &rateLimit{
			count:     1,
			windowEnd: now + 60,
		}
		return true, limit - 1
	}

	// 如果时间窗口已过，重置计数
	if now >= state.windowEnd {
		state.count = 1
		state.windowEnd = now + 60
		return true, limit - 1
	}

	// 检查是否超过限制
	if state.count >= limit {
		remaining := 0
		return false, remaining
	}

	state.count++
	return true, limit - state.count
}

// Reset 重置某个键的限流状态
func (rl *MemoryRateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limits, key)
}

// Cleanup 清理过期的限流状态
func (rl *MemoryRateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().Unix()
	for key, state := range rl.limits {
		if now >= state.windowEnd {
			delete(rl.limits, key)
		}
	}
}

// RateLimitMiddleware 创建限流中间件
func RateLimitMiddleware(limiter RateLimiter, excludePaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否在排除列表中
		path := c.Request.URL.Path
		for _, excludedPath := range excludePaths {
			if path == excludedPath {
				c.Next()
				return
			}
		}

		// 获取客户端标识（IP 或 API Key）
		var key string
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			key = "apikey:" + apiKey
		} else {
			key = "ip:" + c.ClientIP()
		}

		// 检查是否允许请求
		allowed, remaining := limiter.Allow(key)

		// 设置响应头
		c.Header("X-RateLimit-Limit", "60")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			// 获取等待时间
			limiterData, ok := limiter.(*MemoryRateLimiter)
			if ok {
				limiterData.mu.RLock()
				if state, exists := limiterData.limits[key]; exists {
					waitTime := state.windowEnd - time.Now().Unix()
					if waitTime > 0 {
						c.Header("Retry-After", fmt.Sprintf("%d", waitTime))
					}
				}
				limiterData.mu.RUnlock()
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":        "rate limit exceeded",
				"message":      "Too many requests. Please try again later.",
				"retry_after":  60,
			})
			return
		}

		c.Next()
	}
}

// APIKeyRateLimitMiddleware API Key 级别的限流中间件
func APIKeyRateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对需要 API Key 的路由进行限流
		if c.Request.URL.Path == "/api/keys" ||
			c.Request.URL.Path == "/api/keys/validate" ||
			c.Request.URL.Path == "/api/shorten" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		key := "apikey:" + apiKey
		allowed, remaining := limiter.Allow(key)

		c.Header("X-RateLimit-Limit", "60")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"message":     "API rate limit exceeded for this key",
				"retry_after": 60,
			})
			return
		}

		c.Next()
	}
}
