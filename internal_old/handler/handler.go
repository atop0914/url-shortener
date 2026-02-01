package handler

import (
	"net/http"
	"url-shortener/internal/model"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.ShortenerService
}

func NewHandler(service *service.ShortenerService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateShortURL(c *gin.Context) {
	var req model.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateShortURL(req.URL, req.CustomCode, req.ExpireIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	url, err := h.service.GetByShortCode(shortCode)
	if err != nil {
		// 检查是否是因为链接过期
		if err.Error() == "link has expired" {
			c.JSON(http.StatusGone, gin.H{"error": "Link has expired"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func (h *Handler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")

	stats, err := h.service.GetStats(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) ListURLs(c *gin.Context) {
	urls, err := h.service.GetAllURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, urls)
}

func (h *Handler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("code")

	err := h.service.DeleteShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "message": "URL shortener service is running"})
}

// 清理过期链接的API
func (h *Handler) CleanupExpiredURLs(c *gin.Context) {
	err := h.service.CleanupExpiredURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup expired URLs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expired URLs cleaned up successfully"})
}