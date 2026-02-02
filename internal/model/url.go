package model

import "time"

// URL 表示一个短链接实体
type URL struct {
	ID          int64      `json:"id" db:"id"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"` // 可选的过期时间
	Clicks      int64      `json:"clicks" db:"clicks"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

// CreateURLRequest 创建短链接的请求参数
type CreateURLRequest struct {
	URL        string `json:"url" binding:"required,url"`           // 原始URL，必须是有效的URL格式
	CustomCode string `json:"custom_code,omitempty"`               // 自定义短码（可选）
	ExpireIn   int    `json:"expire_in,omitempty"`                 // 过期时间（小时），0表示不过期
}

// CreateURLResponse 创建短链接的响应结果
type CreateURLResponse struct {
	ShortURL  string `json:"short_url"`                            // 生成的短链接
	Code      string `json:"code"`                                // 短码
	Original  string `json:"original"`                            // 原始URL
	CreatedAt string `json:"created_at"`                          // 创建时间
	ExpiresAt string `json:"expires_at,omitempty"`                // 可选的过期时间
}

// StatsResponse 短链接统计信息响应
type StatsResponse struct {
	OriginalURL string `json:"original_url"`                       // 原始URL
	ShortCode   string `json:"short_code"`                         // 短码
	Clicks      int64  `json:"clicks"`                           // 点击次数
	CreatedAt   string `json:"created_at"`                         // 创建时间
	ExpiresAt   string `json:"expires_at,omitempty"`               // 可选的过期时间
	IsActive    bool   `json:"is_active"`                         // 是否活跃
}

// ErrorResponse 用于统一错误响应格式
type ErrorResponse struct {
	Error string `json:"error"`                                   // 错误信息
}

// VisitAnalyticsRequest 访问分析请求参数
type VisitAnalyticsRequest struct {
	ShortCode string `uri:"code" binding:"required"`              // 短码，从URI路径获取
}