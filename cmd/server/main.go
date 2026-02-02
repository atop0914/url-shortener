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
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// initApp 初始化应用程序组件
func initApp() (*http.Server, error) {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	db, err := config.InitDB(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化存储库
	urlRepo := repository.NewURLRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// 初始化服务
	shortenerService := service.NewEnhancedShortenerService(urlRepo, analyticsRepo, cfg.BaseURL)

	// 初始化处理器
	enhancedHandler := handler.NewEnhancedHandler(shortenerService)

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

	// 健康检查路由
	router.GET("/health", enhancedHandler.HealthCheck)

	// API路由组
	api := router.Group("/api")
	{
		// 短链接相关路由
		api.POST("/shorten", enhancedHandler.CreateShortURL)
		api.GET("/stats/:code", enhancedHandler.GetStats)
		api.GET("/analytics/:code", enhancedHandler.GetAdvancedAnalytics)
		api.GET("/visits/:code", enhancedHandler.GetRecentVisits)
		api.DELETE("/urls/:code", enhancedHandler.DeleteURL)
		
		// 管理员功能
		api.GET("/urls", enhancedHandler.ListURLs)
		api.POST("/cleanup", enhancedHandler.CleanupExpiredURLs)
	}

	// 主页路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to URL Shortener Service",
			"version": "1.0.0",
			"docs":    "/docs (if available)",
		})
	})

	// 重定向路由 - 放在最后，避免与其他路由冲突
	router.GET("/:code", enhancedHandler.Redirect)

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
	// 简单的端口提取（在完整应用中可能需要更复杂的逻辑）
	if len(addr) > 1 && addr[0] == ':' {
		var port int
		fmt.Sscanf(addr, ":%d", &port)
		return port
	}
	return 8080 // 默认端口
}