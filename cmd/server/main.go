package main

import (
	"log"
	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	repo, err := repository.NewURLRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close() // 确保在程序退出时关闭数据库连接

	// 初始化分析数据库
	analyticsRepo, err := repository.NewAnalyticsRepository(repo.DB())
	if err != nil {
		log.Fatalf("Failed to initialize analytics database: %v", err)
	}

	// 初始化增强服务
	shortenerService := service.NewEnhancedShortenerService(repo, analyticsRepo, cfg.BaseURL)

	// 初始化增强处理器
	h := handler.NewEnhancedHandler(shortenerService)

	// 设置路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", h.HealthCheck)

	// API 路由
	api := r.Group("/api")
	{
		api.POST("/shorten", h.CreateShortURL)
		api.GET("/stats/:code", h.GetStats)
		
		// 高级分析API
		api.GET("/analytics/:code", h.GetAdvancedAnalytics)
		api.GET("/visits/:code", h.GetRecentVisits)
		
		api.GET("/urls", h.ListURLs)
		api.DELETE("/links/:code", h.DeleteURL)
		api.POST("/cleanup", h.CleanupExpiredURLs) // 新增清理过期链接API
	}

	// 重定向路由
	r.GET("/:code", h.Redirect)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Base URL: %s", cfg.BaseURL)
	log.Println("Features: Custom short codes, Expiration control, Statistics tracking, Advanced Analytics")
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}