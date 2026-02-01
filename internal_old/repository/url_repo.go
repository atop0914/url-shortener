package repository

import (
	"database/sql"
	"time"
	"url-shortener/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(dbPath string) (*URLRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &URLRepository{db: db}
	err = repo.initDB()
	if err != nil {
		return nil, err
	}

	return repo, nil
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
	return err
}

func (r *URLRepository) GetByShortCode(shortCode string) (*model.URL, error) {
	var url model.URL
	var expiresAtStr *string
	
	query := `SELECT id, original_url, short_code, created_at, expires_at, clicks, is_active FROM urls WHERE short_code = ? AND is_active = 1`
	err := r.db.QueryRow(query, shortCode).Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &expiresAtStr, &url.Clicks, &url.IsActive)
	if err != nil {
		return nil, err
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
	_, err := r.db.Exec(query, shortCode)
	return err
}

func (r *URLRepository) GetAll() ([]*model.URL, error) {
	query := `SELECT id, original_url, short_code, created_at, expires_at, clicks, is_active FROM urls ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		var url model.URL
		var expiresAtStr *string
		err := rows.Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &expiresAtStr, &url.Clicks, &url.IsActive)
		if err != nil {
			return nil, err
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
	_, err := r.db.Exec(query, shortCode)
	return err
}

// 删除过期的URL
func (r *URLRepository) DeleteExpiredURLs() error {
	now := time.Now().Format(time.RFC3339)
	query := `UPDATE urls SET is_active = 0 WHERE expires_at IS NOT NULL AND expires_at < ? AND is_active = 1`
	_, err := r.db.Exec(query, now)
	return err
}