package model

import "time"

type URL struct {
	ID          int64     `json:"id" db:"id"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Clicks      int64     `json:"clicks" db:"clicks"`
}

type CreateURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type CreateURLResponse struct {
	ShortURL  string `json:"short_url"`
	Code      string `json:"code"`
	Original  string `json:"original"`
	CreatedAt string `json:"created_at"`
}

type StatsResponse struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	Clicks      int64  `json:"clicks"`
	CreatedAt   string `json:"created_at"`
}