package repository

import (
	"database/sql"
	"fmt"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/database"
	"url-shortener/internal/model"
	"url-shortener/internal/utils"

	_ "github.com/mattn/go-sqlite3"
)

type URLRepository struct {
	db           *sql.DB
	dialect      database.Dialect
	urlCache     *cache.URLCache
	cacheEnabled bool
}

func NewURLRepository(db *sql.DB) *URLRepository {
	// 根据数据库类型获取对应的方言
	dbType := database.ParseDBType("")
	dialect := database.GetDialect(dbType)

	repo := &URLRepository{
		db:           db,
		dialect:      dialect,
		urlCache:     cache.NewURLCache(1000, 5*time.Minute),
		cacheEnabled: true,
	}
	repo.initDB()
	return repo
}

// NewURLRepositoryWithDialect 创建一个带有特定方言的仓库（用于已知数据库类型时）
func NewURLRepositoryWithDialect(db *sql.DB, dbType database.DBType) *URLRepository {
	dialect := database.GetDialect(dbType)

	repo := &URLRepository{
		db:      db,
		dialect: dialect,
	}
	repo.initDB()
	return repo
}

func (r *URLRepository) initDB() error {
	// 使用方言构建正确的 SQL
	autoInc := r.dialect.GetAutoIncrement("id")
	boolType := r.dialect.GetBooleanType()
	dateTimeType := r.dialect.GetDateTimeType()
	ifNotExists := r.dialect.GetIfNotExists()
	defaultNow := r.dialect.GetDefaultNow()

	query := fmt.Sprintf(`
		CREATE TABLE %s urls (
			id %s,
			original_url TEXT NOT NULL,
			short_code TEXT UNIQUE NOT NULL,
			created_at %s %s,
			expires_at %s,
			clicks INTEGER DEFAULT 0,
			is_active %s DEFAULT 1
		);
		CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_expires_at ON urls(expires_at);
	`, ifNotExists, autoInc, dateTimeType, defaultNow, dateTimeType, boolType)

	_, err := r.db.Exec(query)
	return err
}

func (r *URLRepository) CreateWithExpiry(originalURL, shortCode string, expiresAt *time.Time) error {
	placeholder := r.dialect.GetPlaceholder(0)

	query := fmt.Sprintf(`INSERT INTO urls (original_url, short_code, expires_at, is_active) VALUES (%s, %s, %s, 1)`, placeholder, placeholder, placeholder)
	_, err := r.db.Exec(query, originalURL, shortCode, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create URL record: %w", err)
	}
	return nil
}

func (r *URLRepository) GetByShortCode(shortCode string) (*model.URL, error) {
	// 先检查缓存
	if r.cacheEnabled {
		if cached, found := r.urlCache.Get(shortCode); found {
			// 从缓存数据构建 URL 对象
			url := &model.URL{
				ShortCode:   cached.ShortCode,
				OriginalURL: cached.OriginalURL,
			}
			return url, nil
		}
	}

	// 缓存未命中，从数据库查询
	var url model.URL
	var expiresAt interface{}

	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`SELECT id, original_url, short_code, created_at, expires_at, clicks, is_active FROM urls WHERE short_code = %s AND is_active = 1`, p1)
	err := r.db.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&expiresAt,
		&url.Clicks,
		&url.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found: %w", utils.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get URL by short code: %w", err)
	}

	// 处理过期时间
	url.ExpiresAt = r.parseExpiryTime(expiresAt)

	// 存入缓存（未过期或永不过期的URL）
	if r.cacheEnabled && (url.ExpiresAt == nil || time.Now().Before(*url.ExpiresAt)) {
		r.urlCache.Set(shortCode, url.OriginalURL, url.ExpiresAt)
	}

	return &url, nil
}

func (r *URLRepository) parseExpiryTime(value interface{}) *time.Time {
	if value == nil {
		return nil
	}

	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		t = v
	case string:
		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05Z07:00",
		}
		for _, format := range formats {
			if t, err = time.Parse(format, v); err == nil {
				return &t
			}
		}
		return nil
	default:
		return nil
	}

	if err == nil {
		return &t
	}
	return nil
}

func (r *URLRepository) IncrementClicks(shortCode string) error {
	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`UPDATE urls SET clicks = clicks + 1 WHERE short_code = %s AND is_active = 1`, p1)
	result, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL not found: %w", utils.ErrNotFound)
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
		var expiresAt interface{}
		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.ShortCode,
			&url.CreatedAt,
			&expiresAt,
			&url.Clicks,
			&url.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}

		url.ExpiresAt = r.parseExpiryTime(expiresAt)
		urls = append(urls, &url)
	}
	return urls, nil
}

// PaginatedQuery 分页查询参数
type PaginatedQuery struct {
	Page     int    // 页码 (从1开始)
	PageSize int    // 每页数量
	Keyword  string // 搜索关键词 (模糊匹配 original_url 和 short_code)
}

// PaginatedResult 分页查询结果
type PaginatedResult struct {
	Items      []*model.URL `json:"items"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
	HasNext    bool         `json:"has_next"`
	HasPrev    bool         `json:"has_prev"`
}

// GetWithPagination 分页获取URL列表
func (r *URLRepository) GetWithPagination(query *PaginatedQuery) (*PaginatedResult, error) {
	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	offset := (query.Page - 1) * query.PageSize

	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM urls WHERE is_active = 1`
	if query.Keyword != "" {
		p1 := r.dialect.GetPlaceholder(0)
		p2 := r.dialect.GetPlaceholder(1)
		countQuery = fmt.Sprintf(`SELECT COUNT(*) FROM urls WHERE is_active = 1 AND (original_url LIKE %s OR short_code LIKE %s)`, p1, p2)
	}
	err := r.db.QueryRow(countQuery, "%"+query.Keyword+"%", "%"+query.Keyword+"%").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count URLs: %w", err)
	}

	// 获取分页数据
	var rows *sql.Rows
	baseQuery := `SELECT id, original_url, short_code, created_at, expires_at, clicks, is_active FROM urls WHERE is_active = 1`
	if query.Keyword != "" {
		p1 := r.dialect.GetPlaceholder(0)
		p2 := r.dialect.GetPlaceholder(1)
		p3 := r.dialect.GetPlaceholder(2)
		p4 := r.dialect.GetPlaceholder(3)
		sql := fmt.Sprintf(`%s AND (original_url LIKE %s OR short_code LIKE %s) ORDER BY created_at DESC LIMIT %s OFFSET %s`, baseQuery, p1, p2, p3, p4)
		rows, err = r.db.Query(sql, "%"+query.Keyword+"%", "%"+query.Keyword+"%", query.PageSize, offset)
	} else {
		p1 := r.dialect.GetPlaceholder(0)
		p2 := r.dialect.GetPlaceholder(1)
		sql := fmt.Sprintf(`%s ORDER BY created_at DESC LIMIT %s OFFSET %s`, baseQuery, p1, p2)
		rows, err = r.db.Query(sql, query.PageSize, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		var url model.URL
		var expiresAt interface{}
		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.ShortCode,
			&url.CreatedAt,
			&expiresAt,
			&url.Clicks,
			&url.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}

		url.ExpiresAt = r.parseExpiryTime(expiresAt)
		urls = append(urls, &url)
	}

	// 计算分页信息
	totalPages := (total + query.PageSize - 1) / query.PageSize

	return &PaginatedResult{
		Items:      urls,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}, nil
}

// SearchURLs 搜索URL (快捷方法)
func (r *URLRepository) SearchURLs(keyword string, page, pageSize int) (*PaginatedResult, error) {
	return r.GetWithPagination(&PaginatedQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	})
}

func (r *URLRepository) DeleteByShortCode(shortCode string) error {
	// 先使缓存失效
	if r.cacheEnabled {
		r.urlCache.Invalidate(shortCode)
	}

	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`UPDATE urls SET is_active = 0 WHERE short_code = %s`, p1)
	result, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL by short code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL not found: %w", utils.ErrNotFound)
	}

	return nil
}

// DeleteExpiredURLs 删除过期的URL
func (r *URLRepository) DeleteExpiredURLs() error {
	now := time.Now()
	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`UPDATE urls SET is_active = 0 WHERE expires_at IS NOT NULL AND expires_at < %s AND is_active = 1`, p1)
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
