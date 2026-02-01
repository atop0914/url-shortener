package model

import (
	"time"
)

type URL struct {
	ID          int64      `json:"id" db:"id"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"` // 可选的过期时间
	Clicks      int64      `json:"clicks" db:"clicks"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

type CreateURLRequest struct {
	URL         string `json:"url" binding:"required,url"`
	CustomCode  string `json:"custom_code,omitempty"`      // 自定义短码
	ExpireIn    int    `json:"expire_in,omitempty"`        // 过期时间（小时），0表示不过期
}

type CreateURLResponse struct {
	ShortURL  string `json:"short_url"`
	Code      string `json:"code"`
	Original  string `json:"original"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at,omitempty"`         // 可选的过期时间
}

type StatsResponse struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	Clicks      int64  `json:"clicks"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`       // 可选的过期时间
	IsActive    bool   `json:"is_active"`
}