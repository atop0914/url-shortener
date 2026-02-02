package middleware

import (
	"net/http"
	"strings"

	"url-shortener/internal/service"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware validates API keys for protected routes
type APIKeyAuthMiddleware struct {
	service *service.APIKeyService
}

// NewAPIKeyAuthMiddleware creates a new API key auth middleware
func NewAPIKeyAuthMiddleware(service *service.APIKeyService) *APIKeyAuthMiddleware {
	return &APIKeyAuthMiddleware{service: service}
}

// RequireAPIKey returns a Gin middleware function that requires a valid API key
func (m *APIKeyAuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		// Check for Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Use: Bearer <api_key>")
			c.Abort()
			return
		}

		apiKey := parts[1]

		// Validate the key
		apikey, err := m.service.ValidateKey(apiKey)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid API key: "+err.Error())
			c.Abort()
			return
		}

		// Store the key info in context
		c.Set("api_key", apiKey)
		c.Set("api_key_name", apikey.Name)

		c.Next()
	}
}
