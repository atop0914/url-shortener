package repository

import (
	"database/sql"
	"fmt"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/utils"

	_ "github.com/mattn/go-sqlite3"
)

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	repo := &URLRepository{db: db}
	repo.initDB()
	return repo
}

func (r *URLRepository) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_url TEXT NOT NULL,
		short_code TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		clicks INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	CREATE INDEX IF NOT EXISTS idx_expires_at ON urls(expires_at);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *URLRepository) CreateWithExpiry(originalURL, shortCode string, expiresAt *time.Time) error {
	var expiresAtStr *string
	if expiresAt != nil {
		temp := expiresAt.Format(time.RFC3339)
		expiresAtStr = &temp
	} else {
		expiresAtStr = nil
	}

	query := `INSERT INTO urls (original_url, short_code, expires_at, is_active) VALUES (?, ?, ?, 1)`
	_, err := r.db.Exec(query, originalURL, shortCode, expiresAtStr)
	if err != nil {
		return fmt.Errorf("failed to create URL record: %w", err)
	}
	return nil
}

func (r *URLRepository) GetByShortCode(shortCode string) (*model.URL, error) {
	var url model.URL
	var expiresAtStr *string

	query := `SELECT id, original_url, short_code, created_at, expires_at, clicks, is_active FROM urls WHERE short_code = ? AND is_active = 1`
	err := r.db.QueryRow(query, shortCode).Scan(
		&url.ID, 
		&url.OriginalURL, 
		&url.ShortCode, 
		&url.CreatedAt, 
		&expiresAtStr, 
		&url.Clicks, 
		&url.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL by short code: %w", err)
	}

	// 处理过期时间
	if expiresAtStr != nil && *expiresAtStr != "" {
		expiresAt, err := time.Parse(time.RFC3339, *expiresAtStr)
		if err == nil {
			url.ExpiresAt = &expiresAt
		}
	}

	return &url, nil
}

func (r *URLRepository) IncrementClicks(shortCode string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = ? AND is_active = 1`
	result, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		// 短码不存在或已被禁用
		return utils.ErrURLNotFound
	}

	return nil
}

func (r *URLRepository) GetAll() ([]*model.URL, error) {
	query := `SELECT id, original_url, short_code, created_at, expires_at, clicks, is_active FROM urls WHERE is_active = 1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all URLs: %w", err)
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		var url model.URL
		var expiresAtStr *string
		err := rows.Scan(
			&url.ID, 
			&url.OriginalURL, 
			&url.ShortCode, 
			&url.CreatedAt, 
			&expiresAtStr, 
			&url.Clicks, 
			&url.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}

		// 处理过期时间
		if expiresAtStr != nil && *expiresAtStr != "" {
			expiresAt, err := time.Parse(time.RFC3339, *expiresAtStr)
			if err == nil {
				url.ExpiresAt = &expiresAt
			}
		}
		urls = append(urls, &url)
	}
	return urls, nil
}

func (r *URLRepository) DeleteByShortCode(shortCode string) error {
	query := `UPDATE urls SET is_active = 0 WHERE short_code = ?`
	result, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL by short code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return utils.ErrURLNotFound
	}

	return nil
}

// DeleteExpiredURLs 删除过期的URL
func (r *URLRepository) DeleteExpiredURLs() error {
	now := time.Now().Format(time.RFC3339)
	query := `UPDATE urls SET is_active = 0 WHERE expires_at IS NOT NULL AND expires_at < ? AND is_active = 1`
	_, err := r.db.Exec(query, now)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired URLs: %w", err)
	}
	return nil
}

// DB 返回数据库连接
func (r *URLRepository) DB() *sql.DB {
	return r.db
}

// Close 关闭数据库连接
func (r *URLRepository) Close() error {
	return r.db.Close()
}