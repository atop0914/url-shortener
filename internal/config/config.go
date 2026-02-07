package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置结构
type Config struct {
	Port            int              // 服务端口
	DatabaseURL     string           // 数据库连接字符串
	BaseURL         string           // 基础URL，用于生成短链接
	Debug           bool             // 调试模式
	RateLimitConfig *RateLimitConfig // 限流配置
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerMinute int      // 每分钟请求数限制
	Enabled           bool     // 是否启用限流
	ExcludePaths      []string // 排除限流的路径列表
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	rateLimitConfig := &RateLimitConfig{
		RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
		Enabled:           getEnvAsBool("RATE_LIMIT_ENABLED", true),
		ExcludePaths:      parseExcludePaths(os.Getenv("RATE_LIMIT_EXCLUDE_PATHS")),
	}

	config := &Config{
		Port:            getEnvAsInt("PORT", 8080),
		DatabaseURL:     getEnv("DATABASE_URL", "./urls.db"),
		BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),
		Debug:           getEnvAsBool("DEBUG", false),
		RateLimitConfig: rateLimitConfig,
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		switch value {
		case "1", "t", "T", "true", "TRUE", "True":
			return true
		case "0", "f", "F", "false", "FALSE", "False":
			return false
		}
	}
	return defaultValue
}

// parseExcludePaths 解析排除路径列表
func parseExcludePaths(env string) []string {
	if env == "" {
		return []string{"/health", "/docs", "/swagger"}
	}
	paths := strings.Split(env, ",")
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d, port must be between 1 and 65535", c.Port)
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	if c.RateLimitConfig.RequestsPerMinute <= 0 {
		return fmt.Errorf("invalid rate limit: %d, requests per minute must be greater than 0",
			c.RateLimitConfig.RequestsPerMinute)
	}

	if c.RateLimitConfig.RequestsPerMinute > 10000 {
		return fmt.Errorf("invalid rate limit: %d, requests per minute cannot exceed 10000",
			c.RateLimitConfig.RequestsPerMinute)
	}

	return nil
}
