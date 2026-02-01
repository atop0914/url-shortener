package repository

import (
	"database/sql"
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
		clicks INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *URLRepository) Create(originalURL, shortCode string) error {
	query := `INSERT INTO urls (original_url, short_code) VALUES (?, ?)`
	_, err := r.db.Exec(query, originalURL, shortCode)
	return err
}

func (r *URLRepository) GetByShortCode(shortCode string) (*model.URL, error) {
	var url model.URL
	query := `SELECT id, original_url, short_code, created_at, clicks FROM urls WHERE short_code = ?`
	err := r.db.QueryRow(query, shortCode).Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.Clicks)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *URLRepository) IncrementClicks(shortCode string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = ?`
	_, err := r.db.Exec(query, shortCode)
	return err
}

func (r *URLRepository) GetAll() ([]*model.URL, error) {
	query := `SELECT id, original_url, short_code, created_at, clicks FROM urls ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		var url model.URL
		err := rows.Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.Clicks)
		if err != nil {
			return nil, err
		}
		urls = append(urls, &url)
	}
	return urls, nil
}

func (r *URLRepository) DeleteByShortCode(shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = ?`
	_, err := r.db.Exec(query, shortCode)
	return err
}