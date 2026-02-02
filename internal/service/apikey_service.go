package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

// APIKeyService handles API key business logic
type APIKeyService struct {
	repo *repository.APIKeyRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(repo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// GenerateKey generates a new API key
func (s *APIKeyService) GenerateKey(req *model.CreateAPIKeyRequest) (*model.APIKeyResponse, error) {
	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	key := "sk_" + hex.EncodeToString(keyBytes)

	// Calculate expiry
	var expiresAt time.Time
	if req.ExpiresIn > 0 {
		expiresAt = time.Now().AddDate(0, 0, req.ExpiresIn)
	}

	apikey := &model.APIKey{
		Key:       key,
		Name:      req.Name,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := s.repo.Create(apikey); err != nil {
		return nil, err
	}

	return &model.APIKeyResponse{
		ID:        apikey.ID,
		Key:       apikey.Key,
		Name:      apikey.Name,
		CreatedAt: apikey.CreatedAt,
		ExpiresAt: apikey.ExpiresAt,
		IsActive:  apikey.IsActive,
	}, nil
}

// ListKeys lists all API keys (without showing the full key for security)
func (s *APIKeyService) ListKeys() ([]model.APIKey, error) {
	return s.repo.GetAll()
}

// ValidateKey validates an API key
func (s *APIKeyService) ValidateKey(key string) (*model.APIKey, error) {
	return s.repo.ValidateKey(key)
}

// RevokeKey deactivates an API key
func (s *APIKeyService) RevokeKey(key string) error {
	return s.repo.Deactivate(key)
}

// DeleteKey permanently deletes an API key
func (s *APIKeyService) DeleteKey(key string) error {
	return s.repo.Delete(key)
}
