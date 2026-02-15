package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"url-shortener/internal/model"
	"url-shortener/internal/utils"
)

// APIKeyRepository API Key 仓储
type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) InitSchema() error {
	// GORM 会自动迁移，不需要手动创建表
	return nil
}

func (r *APIKeyRepository) Create(key *model.APIKey) error {
	return r.db.Create(key).Error
}

func (r *APIKeyRepository) GetByKey(key string) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := r.db.Where("`key` = ?", key).First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) ValidateKey(key string) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := r.db.Where("`key` = ? AND is_active = ?", key, true).First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid API key: %w", utils.ErrUnauthorized)
		}
		return nil, err
	}
	// 检查过期
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired: %w", utils.ErrUnauthorized)
	}
	// 更新最后使用时间
	r.db.Model(&apiKey).Update("last_used", time.Now())
	return &apiKey, nil
}

func (r *APIKeyRepository) GetAll() ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepository) Deactivate(key string) error {
	return r.db.Model(&model.APIKey{}).Where("`key` = ?", key).Update("is_active", false).Error
}

func (r *APIKeyRepository) Delete(key string) error {
	return r.db.Where("`key` = ?", key).Delete(&model.APIKey{}).Error
}
