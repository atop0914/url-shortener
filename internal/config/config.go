package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Config 应用配置结构
type Config struct {
	Port             int           // 服务端口
	DatabaseURL      string        // 数据库连接字符串
	BaseURL          string        // 基础URL，用于生成短链接
	Debug            bool          // 调试模式
	ReadTimeout      time.Duration // HTTP读取超时
	WriteTimeout     time.Duration // HTTP写入超时
	IdleTimeout      time.Duration // HTTP空闲超时
	MaxConnections   int           // 数据库最大连接数
	MinConnections   int           // 数据库最小连接数
	ConnectionLife   time.Duration // 数据库连接生命周期
	RequestTimeout   time.Duration // 请求超时时间
	RateLimitWindow  time.Duration // 速率限制窗口
	RateLimitCount   int           // 速率限制请求数
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	config := &Config{
		Port:            getEnvAsInt("PORT", 8080),
		DatabaseURL:     getEnv("DATABASE_URL", "./urls.db"),
		BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),
		Debug:           getEnvAsBool("DEBUG", false),
		ReadTimeout:     time.Duration(getEnvAsInt("READ_TIMEOUT", 15)) * time.Second,
		WriteTimeout:    time.Duration(getEnvAsInt("WRITE_TIMEOUT", 15)) * time.Second,
		IdleTimeout:     time.Duration(getEnvAsInt("IDLE_TIMEOUT", 60)) * time.Second,
		MaxConnections:  getEnvAsInt("MAX_CONNECTIONS", 25),
		MinConnections:  getEnvAsInt("MIN_CONNECTIONS", 5),
		ConnectionLife:  time.Duration(getEnvAsInt("CONNECTION_LIFETIME", 300)) * time.Second,
		RequestTimeout:  time.Duration(getEnvAsInt("REQUEST_TIMEOUT", 30)) * time.Second,
		RateLimitWindow: time.Duration(getEnvAsInt("RATE_LIMIT_WINDOW", 1)) * time.Minute,
		RateLimitCount:  getEnvAsInt("RATE_LIMIT_COUNT", 100),
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
		log.Printf("Warning: Invalid value for %s: %s, using default: %d", key, value, defaultValue)
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
		default:
			log.Printf("Warning: Invalid boolean value for %s: %s, using default: %t", key, value, defaultValue)
		}
	}
	return defaultValue
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d, port must be between 1 and 65535", c.Port)
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if c.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}

	if c.MinConnections < 0 || c.MinConnections > c.MaxConnections {
		return fmt.Errorf("min connections must be between 0 and max connections")
	}

	return nil
}