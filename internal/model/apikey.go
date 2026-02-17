package model

import (
	"time"
)

// APIKey API Key 实体
type APIKey struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	APIKey    string     `gorm:"type:varchar(128);uniqueIndex;not null;column:key" json:"key"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at,omitempty"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

// CreateAPIKeyRequest 创建 API Key 的请求参数
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	ExpiresIn int    `json:"expires_in"`
}

// APIKeyResponse API Key 响应
type APIKeyResponse struct {
	ID        uint      `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}
