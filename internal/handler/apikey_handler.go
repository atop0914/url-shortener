package handler

import (
	"net/http"

	"url-shortener/internal/model"
	"url-shortener/internal/service"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler handles API key HTTP requests
type APIKeyHandler struct {
	service *service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(service *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

// CreateKey handles POST /api/keys
func (h *APIKeyHandler) CreateKey(c *gin.Context) {
	var req model.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	response, err := h.service.GenerateKey(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create API key: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "API key created successfully",
		"data":    response,
	})
}

// ListKeys handles GET /api/keys
func (h *APIKeyHandler) ListKeys(c *gin.Context) {
	keys, err := h.service.ListKeys()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list API keys: "+err.Error())
		return
	}

	// Mask the keys for security
	maskedKeys := make([]model.APIKey, len(keys))
	for i, key := range keys {
		key.Key = maskKey(key.Key)
		maskedKeys[i] = key
	}

	c.JSON(http.StatusOK, gin.H{
		"data": maskedKeys,
	})
}

// RevokeKey handles DELETE /api/keys/:key
func (h *APIKeyHandler) RevokeKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "API key is required")
		return
	}

	if err := h.service.RevokeKey(key); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to revoke API key: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
	})
}

// ValidateKey handles GET /api/keys/validate
func (h *APIKeyHandler) ValidateKey(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "API key is required")
		return
	}

	apikey, err := h.service.ValidateKey(key)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid API key: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"data": gin.H{
			"name":      apikey.Name,
			"created_at": apikey.CreatedAt,
			"last_used":  apikey.LastUsed,
		},
	})
}

// maskKey masks a portion of the API key for security
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}
