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
	
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/handler"
	"url-shortener/internal/middleware"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	// 导入所有支持的数据库驱动
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// initApp 初始化应用程序组件
func initApp() (*http.Server, error) {
	// 加载配置
	cfg := config.LoadConfig()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// 初始化数据库连接
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 根据数据库类型获取方言并初始化存储库
	dbType := database.ParseDBType(cfg.DatabaseURL)
	urlRepo := repository.NewURLRepositoryWithDialect(db, dbType)
	analyticsRepo := repository.NewAnalyticsRepositoryWithDialect(db, dbType)
	apiKeyRepo := repository.NewAPIKeyRepositoryWithDialect(db, dbType)

	// 初始化 API Key 表
	if err := apiKeyRepo.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize API key schema: %w", err)
	}

	// 初始化服务
	shortenerService := service.NewEnhancedShortenerService(urlRepo, analyticsRepo, cfg.BaseURL)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo)

	// 初始化处理器
	enhancedHandler := handler.NewEnhancedHandler(shortenerService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)

	// 初始化中间件
	apiKeyMiddleware := middleware.NewAPIKeyAuthMiddleware(apiKeyService)

	// 初始化限流器
	var rateLimiter middleware.RateLimiter
	if cfg.RateLimitConfig.Enabled {
		rateLimiter = middleware.NewMemoryRateLimiter(cfg.RateLimitConfig)
		log.Printf("Rate limiting enabled: %d requests per minute", cfg.RateLimitConfig.RequestsPerMinute)
	} else {
		log.Println("Rate limiting disabled")
	}

	// 设置Gin模式
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由引擎
	router := gin.New()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS 中间件
	router.Use(cors.Default())

	// Rate Limiting 中间件（应用到所有 API 路由）
	if rateLimiter != nil && cfg.RateLimitConfig.Enabled {
		router.Use(middleware.RateLimitMiddleware(rateLimiter, cfg.RateLimitConfig.ExcludePaths))
	}

	// 健康检查路由
	router.GET("/health", enhancedHandler.HealthCheck)

	// 公开路由 - 不需要 API Key
	apiPublic := router.Group("/api")
	{
		// API Key 管理（公开，因为还没有 key）
		apiPublic.POST("/keys", apiKeyHandler.CreateKey)
		apiPublic.GET("/keys/validate", apiKeyHandler.ValidateKey)

		// 短链接相关路由（创建需要 API Key）
		apiPublic.POST("/shorten", apiKeyMiddleware.RequireAPIKey(), enhancedHandler.CreateShortURL)

		// 重定向路由（公开访问）
		router.GET("/:code", enhancedHandler.Redirect)
	}

	// 受保护路由 - 需要 API Key
	apiProtected := router.Group("/api")
	apiProtected.Use(apiKeyMiddleware.RequireAPIKey())
	{
		// 短链接相关路由
		apiProtected.GET("/stats/:code", enhancedHandler.GetStats)
		apiProtected.GET("/analytics/:code", enhancedHandler.GetAdvancedAnalytics)
		apiProtected.GET("/visits/:code", enhancedHandler.GetRecentVisits)
		apiProtected.DELETE("/urls/:code", enhancedHandler.DeleteURL)

		// 管理员功能
		apiProtected.GET("/urls", enhancedHandler.ListURLs)                         // 获取所有URL
		apiProtected.GET("/urls/pagination", enhancedHandler.GetURLsWithPagination) // 分页获取URL
		apiProtected.GET("/urls/search", enhancedHandler.SearchURLs)               // 搜索URL
		apiProtected.POST("/cleanup", enhancedHandler.CleanupExpiredURLs)

		// API Key 管理
		apiProtected.GET("/keys", apiKeyHandler.ListKeys)
		apiProtected.DELETE("/keys/:key", apiKeyHandler.RevokeKey)
	}

	// 主页路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Welcome to URL Shortener Service",
			"version":  "1.0.0",
			"docs":     "/docs (if available)",
			"database": string(dbType),
		})
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	return server, nil
}

// gracefulShutdown 等待中断信号并优雅关闭服务器
func gracefulShutdown(server *http.Server) {
	// 创建中断信号通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 创建5秒超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func main() {
	// 初始化应用
	server, err := initApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("Starting server on port %d", getPortFromAddr(server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号并优雅关闭
	gracefulShutdown(server)
}

// getPortFromAddr 从地址字符串中提取端口号
func getPortFromAddr(addr string) int {
	if len(addr) > 1 && addr[0] == ':' {
		var port int
		fmt.Sscanf(addr, ":%d", &port)
		return port
	}
	return 8080
}