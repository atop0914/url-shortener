package handler

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/service"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
)

// EnhancedHandler 封装了所有处理函数
type EnhancedHandler struct {
	service *service.EnhancedShortenerService
}

// NewEnhancedHandler 创建一个新的 EnhancedHandler 实例
func NewEnhancedHandler(service *service.EnhancedShortenerService) *EnhancedHandler {
	return &EnhancedHandler{
		service: service,
	}
}

// CreateShortURL 处理创建短链接请求
func (h *EnhancedHandler) CreateShortURL(c *gin.Context) {
	var req model.CreateURLRequest
	
	// 绑定请求体到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: utils.ValidateAndFormatError(err),
		})
		return
	}

	// 验证URL格式
	if !utils.IsValidURL(req.URL) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid URL format",
		})
		return
	}

	// 调用服务层创建短链接
	resp, err := h.service.CreateShortURL(req.URL, req.CustomCode, req.ExpireIn)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Redirect 处理短链接重定向请求
func (h *EnhancedHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	// 验证短码格式
	if !utils.IsValidShortCode(shortCode) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid short code format",
		})
		return
	}

	// 获取客户端IP地址
	clientIP := h.getClientIP(c.Request)

	// 获取User-Agent和Referer
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// 获取短链接并记录分析数据
	url, err := h.service.GetByShortCodeWithContext(c.Request.Context(), shortCode, clientIP, userAgent, referer)
	if err != nil {
		h.handleURLError(c, err)
		return
	}

	// 301永久重定向到原链接
	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

// GetStats 获取短链接统计信息
func (h *EnhancedHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")

	// 验证短码格式
	if !utils.IsValidShortCode(shortCode) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid short code format",
		})
		return
	}

	stats, err := h.service.GetStats(shortCode)
	if err != nil {
		h.handleURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAdvancedAnalytics 获取高级分析数据
func (h *EnhancedHandler) GetAdvancedAnalytics(c *gin.Context) {
	var req model.VisitAnalyticsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: utils.ValidateAndFormatError(err),
		})
		return
	}

	// 验证短码格式
	if !utils.IsValidShortCode(req.ShortCode) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid short code format",
		})
		return
	}

	// 解析时间范围参数
	since, until, parseErr := h.parseTimeRangeParams(c)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: parseErr.Error()})
		return
	}

	analytics, err := h.service.GetAdvancedAnalytics(req.ShortCode, since, until)
	if err != nil {
		h.handleURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetRecentVisits 获取最近访问记录
func (h *EnhancedHandler) GetRecentVisits(c *gin.Context) {
	var req model.VisitAnalyticsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: utils.ValidateAndFormatError(err),
		})
		return
	}

	// 验证短码格式
	if !utils.IsValidShortCode(req.ShortCode) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid short code format",
		})
		return
	}

	// 解析时间范围和分页参数
	since, _, parseErr := h.parseTimeRangeParams(c)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: parseErr.Error()})
		return
	}

	// 解析分页参数
	limit := h.parseLimitParam(c.DefaultQuery("limit", "100"))

	visits, err := h.service.GetRecentVisits(req.ShortCode, limit, since)
	if err != nil {
		h.handleURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, visits)
}

// ListURLs 列出所有URL（管理员功能）
func (h *EnhancedHandler) ListURLs(c *gin.Context) {
	urls, err := h.service.GetAllURLs()
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, urls)
}

// GetURLsWithPagination 分页获取URL列表
// GET /api/urls?page=1&page_size=10&keyword=example
func (h *EnhancedHandler) GetURLsWithPagination(c *gin.Context) {
	// 解析分页参数
	page := h.parsePageParam(c.DefaultQuery("page", "1"))
	pageSize := h.parsePageParam(c.DefaultQuery("page_size", "10"))
	if pageSize > 100 { // 限制最大页面大小
		pageSize = 100
	}
	keyword := c.Query("keyword")

	result, err := h.service.GetURLsWithPagination(page, pageSize, keyword)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchURLs 搜索URL
// GET /api/urls/search?keyword=example&page=1&page_size=10
func (h *EnhancedHandler) SearchURLs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Keyword is required"})
		return
	}

	page := h.parsePageParam(c.DefaultQuery("page", "1"))
	pageSize := h.parsePageParam(c.DefaultQuery("page_size", "10"))
	if pageSize > 100 { // 限制最大页面大小
		pageSize = 100
	}

	result, err := h.service.SearchURLs(keyword, page, pageSize)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteURL 删除指定的短链接
func (h *EnhancedHandler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("code")

	// 验证短码格式
	if !utils.IsValidShortCode(shortCode) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid short code format",
		})
		return
	}

	err := h.service.DeleteShortCode(shortCode)
	if err != nil {
		h.handleURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}

// HealthCheck 健康检查端点
func (h *EnhancedHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "URL shortener service is running",
		"timestamp": time.Now().Unix(),
		"version": "1.1.0", // 添加版本信息
	})
}

// CleanupExpiredURLs 清理过期链接的API
func (h *EnhancedHandler) CleanupExpiredURLs(c *gin.Context) {
	err := h.service.CleanupExpiredURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "Failed to cleanup expired URLs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expired URLs cleaned up successfully"})
}

// 辅助方法：解析时间范围参数
func (h *EnhancedHandler) parseTimeRangeParams(c *gin.Context) (*time.Time, *time.Time, error) {
	var since, until *time.Time

	if c.Query("since") != "" {
		parsedSince, err := time.Parse("2006-01-02", c.Query("since"))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid since date format, expected YYYY-MM-DD: %w", err)
		}
		since = &parsedSince
	}

	if c.Query("until") != "" {
		parsedUntil, err := time.Parse("2006-01-02", c.Query("until"))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid until date format, expected YYYY-MM-DD: %w", err)
		}
		until = &parsedUntil
	}

	// 验证时间范围逻辑
	if since != nil && until != nil && since.After(*until) {
		return nil, nil, fmt.Errorf("since date cannot be after until date")
	}

	return since, until, nil
}

// 辅助方法：解析limit参数
func (h *EnhancedHandler) parseLimitParam(limitStr string) int {
	var limit int
	_, err := fmt.Sscanf(limitStr, "%d", &limit)
	if err != nil || limit <= 0 {
		return 100 // 默认值
	}
	if limit > 1000 { // 限制最大值
		return 1000
	}
	return limit
}

// 辅助方法：解析page参数
func (h *EnhancedHandler) parsePageParam(pageStr string) int {
	var page int
	_, err := fmt.Sscanf(pageStr, "%d", &page)
	if err != nil || page <= 0 {
		return 1 // 默认值
	}
	if page > 10000 { // 防止过大的页码
		return 10000
	}
	return page
}

// 辅助方法：处理通用服务错误
func (h *EnhancedHandler) handleServiceError(c *gin.Context, err error) {
	// 根据错误类型返回相应HTTP状态码
	switch {
	case errors.Is(err, utils.ErrCustomCodeExists):
		c.JSON(http.StatusConflict, model.ErrorResponse{Error: err.Error()})
	case errors.Is(err, utils.ErrInvalidCustomCode):
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	case strings.Contains(err.Error(), "database"):
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error occurred"})
	case strings.Contains(err.Error(), "generate"):
		c.JSON(http.StatusTooManyRequests, model.ErrorResponse{Error: "Rate limit exceeded, please try again later"})
	default:
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
}

// 辅助方法：处理URL相关错误
func (h *EnhancedHandler) handleURLError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, utils.ErrURLNotFound):
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "URL not found"})
	case errors.Is(err, utils.ErrURLExpired):
		c.JSON(http.StatusGone, model.ErrorResponse{Error: "Link has expired"})
	case strings.Contains(err.Error(), "database"):
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error occurred"})
	default:
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
}

// getClientIP 获取客户端真实IP地址
func (h *EnhancedHandler) getClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头部（可能包含多个IP，取第一个）
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ip := strings.Split(forwarded, ",")[0]
		return strings.TrimSpace(ip)
	}

	// 检查 X-Real-IP 头部
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 使用 RemoteAddr
	addr := r.RemoteAddr
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		// 如果无法分离主机和端口，返回原始地址
		return addr
	}
	return ip
}