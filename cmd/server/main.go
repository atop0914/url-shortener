package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"url-shortener/internal/config"
	"url-shortener/internal/database/gormdb"
	"url-shortener/internal/handler"
	"url-shortener/internal/middleware"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := gorm.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 创建仓储
	urlRepo := repository.NewURLRepository(db.GetDB())
	apiKeyRepo := repository.NewAPIKeyRepository(db.GetDB())
	analyticsRepo := repository.NewAnalyticsRepository(db.GetDB())

	// 初始化服务
	shortenerService := service.NewEnhancedShortenerService(urlRepo, analyticsRepo, cfg.BaseURL)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo)

	// 初始化处理器
	enhancedHandler := handler.NewEnhancedHandler(shortenerService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)

	// 初始化中间件
	apiKeyMiddleware := middleware.NewAPIKeyAuthMiddleware(apiKeyService)

	// 设置 Gin 模式
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"message":   "URL shortener service is running",
			"timestamp": time.Now().Unix(),
		})
	})

	// 公开路由
	router.GET("/:code", enhancedHandler.Redirect)

	// API 路由
	api := router.Group("/api")
	{
		// 公开 API
		api.POST("/keys", apiKeyHandler.CreateKey)
		api.GET("/keys/validate", apiKeyHandler.ValidateKey)

		// 需要 API Key 的路由
		api.POST("/shorten", apiKeyMiddleware.RequireAPIKey(), enhancedHandler.CreateShortURL)

		// 受保护的路由
		protected := api.Group("")
		protected.Use(apiKeyMiddleware.RequireAPIKey())
		{
			protected.GET("/stats/:code", enhancedHandler.GetStats)
			protected.GET("/analytics/:code", enhancedHandler.GetAdvancedAnalytics)
			protected.GET("/visits/:code", enhancedHandler.GetRecentVisits)
			protected.DELETE("/urls/:code", enhancedHandler.DeleteURL)
			protected.GET("/urls", enhancedHandler.ListURLs)
			protected.GET("/urls/pagination", enhancedHandler.GetURLsWithPagination)
			protected.GET("/urls/search", enhancedHandler.SearchURLs)
			protected.POST("/cleanup", enhancedHandler.CleanupExpiredURLs)
			protected.GET("/keys", apiKeyHandler.ListKeys)
			protected.DELETE("/keys/:key", apiKeyHandler.RevokeKey)
		}
	}

	// 首页
	router.GET("/", func(c *gin.Context) {
		dbType := "sqlite"
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to URL Shortener Service",
			"version": "1.0.0",
			"database": dbType,
		})
	})

	// 创建服务器
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("Starting server on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

// GetDBType 获取数据库类型
