package config

import (
	"os"
)

type Config struct {
	Port       string
	DBPath     string
	BaseURL    string
}

func LoadConfig() *Config {
	return &Config{
		Port:    getEnv("PORT", "8080"),
		DBPath:  getEnv("DB_PATH", "./urls.db"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}