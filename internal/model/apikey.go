package model

import "time"

// APIKey represents an API key for authentication
type APIKey struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
	IsActive  bool      `json:"is_active"`
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	ExpiresIn int    `json:"expires_in"` // in days, 0 means never expires
}

// APIKeyResponse represents the API key response (key is only shown once)
type APIKeyResponse struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}
