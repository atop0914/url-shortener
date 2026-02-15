package repository

import (
	"database/sql"
	"fmt"
	"time"

	"url-shortener/internal/database"
	"url-shortener/internal/model"
	"url-shortener/internal/utils"
)

// APIKeyRepository handles API key database operations
type APIKeyRepository struct {
	db      *sql.DB
	dialect database.Dialect
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	dbType := database.ParseDBType("")
	dialect := database.GetDialect(dbType)

	return &APIKeyRepository{
		db:      db,
		dialect: dialect,
	}
}

// NewAPIKeyRepositoryWithDialect 创建一个带有特定方言的仓库
func NewAPIKeyRepositoryWithDialect(db *sql.DB, dbType database.DBType) *APIKeyRepository {
	dialect := database.GetDialect(dbType)

	return &APIKeyRepository{
		db:      db,
		dialect: dialect,
	}
}

// InitSchema creates the API keys table if it doesn't exist
func (r *APIKeyRepository) InitSchema() error {
	boolType := r.dialect.GetBooleanType()
	dateTimeType := r.dialect.GetDateTimeType()
	ifNotExists := r.dialect.GetIfNotExists()
	defaultNow := r.dialect.GetDefaultNow()
	autoInc := r.dialect.GetAutoIncrement("id")

	query := fmt.Sprintf(`
		CREATE TABLE %s api_keys (
			id %s,
			key TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			created_at %s %s,
			expires_at %s,
			last_used %s,
			is_active %s NOT NULL DEFAULT 1
		);

		CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
		CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);
	`, ifNotExists, autoInc, dateTimeType, defaultNow, dateTimeType, dateTimeType, boolType)

	_, err := r.db.Exec(query)
	return err
}

// Create creates a new API key
func (r *APIKeyRepository) Create(apikey *model.APIKey) error {
	p1, p2, p3, p4 := r.dialect.GetPlaceholder(0), r.dialect.GetPlaceholder(1), r.dialect.GetPlaceholder(2), r.dialect.GetPlaceholder(3)
	query := fmt.Sprintf(`
		INSERT INTO api_keys (key, name, expires_at, is_active)
		VALUES (%s, %s, %s, %s)
	`, p1, p2, p3, p4)

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
	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`
		SELECT id, key, name, created_at, expires_at, last_used, is_active
		FROM api_keys
		WHERE key = %s
	`, p1)

	var createdAtStr, expiresAtStr, lastUsedStr sql.NullString
	apikey := &model.APIKey{}
	err := r.db.QueryRow(query, key).Scan(
		&apikey.ID, &apikey.Key, &apikey.Name,
		&createdAtStr, &expiresAtStr, &lastUsedStr, &apikey.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	// 使用工具函数解析时间
	apikey.CreatedAt = utils.ParseTime(createdAtStr.String)
	apikey.ExpiresAt = utils.ParseTimePtr(expiresAtStr.String)
	apikey.LastUsed = utils.ParseTimePtr(lastUsedStr.String)
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
		var createdAtStr, expiresAtStr, lastUsedStr sql.NullString
		err := rows.Scan(
			&apikey.ID, &apikey.Key, &apikey.Name,
			&createdAtStr, &expiresAtStr, &lastUsedStr, &apikey.IsActive,
		)
		if err != nil {
			return nil, err
		}
		// 使用工具函数解析时间
		apikey.CreatedAt = utils.ParseTime(createdAtStr.String)
		apikey.ExpiresAt = utils.ParseTimePtr(expiresAtStr.String)
		apikey.LastUsed = utils.ParseTimePtr(lastUsedStr.String)
		keys = append(keys, apikey)
	}
	return keys, nil
}

// UpdateLastUsed updates the last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(key string) error {
	p1, p2 := r.dialect.GetPlaceholder(0), r.dialect.GetPlaceholder(1)
	query := fmt.Sprintf(`UPDATE api_keys SET last_used = %s WHERE key = %s`, p1, p2)
	_, err := r.db.Exec(query, time.Now(), key)
	return err
}

// Deactivate disables an API key
func (r *APIKeyRepository) Deactivate(key string) error {
	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`UPDATE api_keys SET is_active = 0 WHERE key = %s`, p1)
	_, err := r.db.Exec(query, key)
	return err
}

// Delete removes an API key
func (r *APIKeyRepository) Delete(key string) error {
	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`DELETE FROM api_keys WHERE key = %s`, p1)
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
	if apikey.ExpiresAt != nil && !apikey.ExpiresAt.IsZero() && apikey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}
	// Update last used timestamp
	r.UpdateLastUsed(key)
	return apikey, nil
}

// DB returns the database connection
func (r *APIKeyRepository) DB() *sql.DB {
	return r.db
}
