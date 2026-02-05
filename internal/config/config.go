package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置结构
type Config struct {
	Port        int    // 服务端口
	DatabaseURL string // 数据库连接字符串
	BaseURL     string // 基础URL，用于生成短链接
	Debug       bool   // 调试模式
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	config := &Config{
		Port:        getEnvAsInt("PORT", 8080),
		DatabaseURL: getEnv("DATABASE_URL", "./urls.db"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		Debug:       getEnvAsBool("DEBUG", false),
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

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d, port must be between 1 and 65535", c.Port)
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	return nil
}