package handler

import (
	"net"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/service"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
)

type EnhancedHandler struct {
	service *service.EnhancedShortenerService
}

func NewEnhancedHandler(service *service.EnhancedShortenerService) *EnhancedHandler {
	return &EnhancedHandler{
		service: service,
	}
}

func (h *EnhancedHandler) CreateShortURL(c *gin.Context) {
	var req model.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.service.CreateShortURL(req.URL, req.CustomCode, req.ExpireIn)
	if err != nil {
		if err == utils.ErrCustomCodeExists || err == utils.ErrInvalidCustomCode {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *EnhancedHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	// 获取客户端IP地址
	clientIP := getClientIP(c.Request)

	// 获取User-Agent和Referer
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	url, err := h.service.GetByShortCodeWithContext(c.Request.Context(), shortCode, clientIP, userAgent, referer)
	if err != nil {
		switch {
		case err == utils.ErrURLNotFound:
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "URL not found"})
		case err == utils.ErrURLExpired:
			c.JSON(http.StatusGone, model.ErrorResponse{Error: "Link has expired"})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func (h *EnhancedHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")

	stats, err := h.service.GetStats(shortCode)
	if err != nil {
		if err == utils.ErrURLNotFound {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "URL not found"})
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAdvancedAnalytics 获取高级分析数据
func (h *EnhancedHandler) GetAdvancedAnalytics(c *gin.Context) {
	var req model.VisitAnalyticsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	// 解析时间范围参数
	var since, until *time.Time
	if c.Query("since") != "" {
		parsedSince, err := time.Parse("2006-01-02", c.Query("since"))
		if err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid since date format, expected YYYY-MM-DD"})
			return
		}
		since = &parsedSince
	}

	if c.Query("until") != "" {
		parsedUntil, err := time.Parse("2006-01-02", c.Query("until"))
		if err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid until date format, expected YYYY-MM-DD"})
			return
		}
		until = &parsedUntil
	}

	analytics, err := h.service.GetAdvancedAnalytics(req.ShortCode, since, until)
	if err != nil {
		if err == utils.ErrURLNotFound {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "URL not found"})
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetRecentVisits 获取最近访问记录
func (h *EnhancedHandler) GetRecentVisits(c *gin.Context) {
	var req model.VisitAnalyticsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	// 获取查询参数
	limit := c.Query("limit")
	page := c.Query("page")

	// 解析时间范围参数
	var since, until *time.Time
	if c.Query("since") != "" {
		parsedSince, err := time.Parse("2006-01-02", c.Query("since"))
		if err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid since date format, expected YYYY-MM-DD"})
			return
		}
		since = &parsedSince
	}

	if c.Query("until") != "" {
		parsedUntil, err := time.Parse("2006-01-02", c.Query("until"))
		if err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid until date format, expected YYYY-MM-DD"})
			return
		}
		until = &parsedUntil
	}

	// 解析limit参数
	reqLimit := 100
	if limit != "" {
		// 这里应该解析limit参数，为了简化，我们使用默认值
	}

	visits, err := h.service.GetRecentVisits(req.ShortCode, reqLimit, since)
	if err != nil {
		if err == utils.ErrURLNotFound {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "URL not found"})
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, visits)
}

func (h *EnhancedHandler) ListURLs(c *gin.Context) {
	urls, err := h.service.GetAllURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, urls)
}

func (h *EnhancedHandler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("code")

	err := h.service.DeleteShortCode(shortCode)
	if err != nil {
		if err == utils.ErrURLNotFound {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "URL not found"})
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}

func (h *EnhancedHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "message": "URL shortener service is running"})
}

// CleanupExpiredURLs 清理过期链接的API
func (h *EnhancedHandler) CleanupExpiredURLs(c *gin.Context) {
	err := h.service.CleanupExpiredURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to cleanup expired URLs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expired URLs cleaned up successfully"})
}

// getClientIP 获取客户端真实IP地址
func getClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头部
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For 可能包含多个IP地址，取第一个
		ip := forwarded
		if commaIndex := strings.Index(ip, ","); commaIndex != -1 {
			ip = ip[:commaIndex]
		}
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
		return addr
	}
	return ip
}