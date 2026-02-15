package model

import (
	"time"
)

// URL 表示一个短链接实体
type URL struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OriginalURL string    `gorm:"type:varchar(2048);not null" json:"original_url"`
	ShortCode   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"short_code"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`
	Clicks      int64     `gorm:"default:0" json:"clicks"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}

func (URL) TableName() string {
	return "urls"
}

// CreateURLRequest 创建短链接的请求参数
type CreateURLRequest struct {
	URL        string `json:"url" binding:"required,url"`
	CustomCode string `json:"custom_code,omitempty"`
	ExpireIn   int    `json:"expire_in,omitempty"`
}

// CreateURLResponse 创建短链接的响应结果
type CreateURLResponse struct {
	ShortURL  string `json:"short_url"`
	Code      string `json:"code"`
	Original  string `json:"original"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// StatsResponse 短链接统计信息响应
type StatsResponse struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	Clicks      int64  `json:"clicks"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// ErrorResponse 用于统一错误响应格式
type ErrorResponse struct {
	Error string `json:"error"`
}
