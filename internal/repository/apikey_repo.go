package repository

import (
	"database/sql"
	"fmt"
	"time"

	"url-shortener/internal/model"
)

// APIKeyRepository handles API key database operations
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// InitSchema creates the API keys table if it doesn't exist
func (r *APIKeyRepository) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		last_used DATETIME,
		is_active INTEGER NOT NULL DEFAULT 1
	);
	
	CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
	CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);
	`
	_, err := r.db.Exec(query)
	return err
}

// Create creates a new API key
func (r *APIKeyRepository) Create(apikey *model.APIKey) error {
	query := `
	INSERT INTO api_keys (key, name, expires_at, is_active)
	VALUES (?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, apikey.Key, apikey.Name, apikey.ExpiresAt, apikey.IsActive)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	apikey.ID = id
	return nil
}

// GetByKey retrieves an API key by its key value
func (r *APIKeyRepository) GetByKey(key string) (*model.APIKey, error) {
	query := `
	SELECT id, key, name, created_at, expires_at, last_used, is_active
	FROM api_keys
	WHERE key = ?
	`
	apikey := &model.APIKey{}
	err := r.db.QueryRow(query, key).Scan(
		&apikey.ID, &apikey.Key, &apikey.Name,
		&apikey.CreatedAt, &apikey.ExpiresAt, &apikey.LastUsed, &apikey.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return apikey, nil
}

// GetAll retrieves all API keys
func (r *APIKeyRepository) GetAll() ([]model.APIKey, error) {
	query := `
	SELECT id, key, name, created_at, expires_at, last_used, is_active
	FROM api_keys
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		apikey := model.APIKey{}
		err := rows.Scan(
			&apikey.ID, &apikey.Key, &apikey.Name,
			&apikey.CreatedAt, &apikey.ExpiresAt, &apikey.LastUsed, &apikey.IsActive,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, apikey)
	}
	return keys, nil
}

// UpdateLastUsed updates the last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(key string) error {
	query := `UPDATE api_keys SET last_used = ? WHERE key = ?`
	_, err := r.db.Exec(query, time.Now(), key)
	return err
}

// Deactivate disables an API key
func (r *APIKeyRepository) Deactivate(key string) error {
	query := `UPDATE api_keys SET is_active = 0 WHERE key = ?`
	_, err := r.db.Exec(query, key)
	return err
}

// Delete removes an API key
func (r *APIKeyRepository) Delete(key string) error {
	query := `DELETE FROM api_keys WHERE key = ?`
	_, err := r.db.Exec(query, key)
	return err
}

// ValidateKey validates an API key and returns its info if valid
func (r *APIKeyRepository) ValidateKey(key string) (*model.APIKey, error) {
	apikey, err := r.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if apikey == nil {
		return nil, fmt.Errorf("invalid API key")
	}
	if !apikey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}
	if !apikey.ExpiresAt.IsZero() && apikey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}
	// Update last used timestamp
	r.UpdateLastUsed(key)
	return apikey, nil
}
