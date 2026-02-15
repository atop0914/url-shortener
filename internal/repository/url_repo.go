package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"url-shortener/internal/model"
	"url-shortener/internal/utils"
)

// PaginatedQuery 分页查询
type PaginatedQuery struct {
	Page     int
	PageSize int
	Keyword  string
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	Items      []*model.URL `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// URLRepository URL 仓储
type URLRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) CreateWithExpiry(originalURL, shortCode string, expiresAt *time.Time) error {
	url := &model.URL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}
	return r.db.Create(url).Error
}

func (r *URLRepository) GetByShortCode(code string) (*model.URL, error) {
	var url model.URL
	err := r.db.Where("short_code = ? AND is_active = ?", code, true).First(&url).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("URL not found: %w", utils.ErrURLNotFound)
		}
		return nil, err
	}
	return &url, nil
}

func (r *URLRepository) IncrementClicks(shortCode string) error {
	return r.db.Model(&model.URL{}).Where("short_code = ?", shortCode).UpdateColumn("clicks", gorm.Expr("clicks + ?", 1)).Error
}

func (r *URLRepository) GetAll() ([]*model.URL, error) {
	var urls []*model.URL
	err := r.db.Where("is_active = ?", true).Order("created_at DESC").Find(&urls).Error
	return urls, err
}

func (r *URLRepository) GetWithPagination(query *PaginatedQuery) (*PaginatedResult, error) {
	var urls []*model.URL
	var total int64

	db := r.db.Model(&model.URL{}).Where("is_active = ?", true)
	
	if query.Keyword != "" {
		db = db.Where("original_url LIKE ? OR short_code LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	db.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&urls).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

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

func (r *URLRepository) SearchURLs(keyword string, page, pageSize int) (*PaginatedResult, error) {
	return r.GetWithPagination(&PaginatedQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	})
}

func (r *URLRepository) DeleteByShortCode(code string) error {
	return r.db.Model(&model.URL{}).Where("short_code = ?", code).Update("is_active", false).Error
}

func (r *URLRepository) DeleteExpiredURLs() error {
	return r.db.Model(&model.URL{}).Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", time.Now(), true).Update("is_active", false).Error
}
