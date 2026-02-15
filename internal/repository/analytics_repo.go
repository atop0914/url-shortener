package repository

import (
	"time"

	"gorm.io/gorm"

	"url-shortener/internal/model"
)

// AnalyticsRepository 分析数据仓储
type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) RecordVisit(record *model.VisitRecord) error {
	return r.db.Create(record).Error
}

func (r *AnalyticsRepository) GetAnalyticsSummary(shortCode string, since, until *time.Time) (*model.AnalyticsSummary, error) {
	var totalVisits int64
	r.db.Model(&model.VisitRecord{}).Where("short_code = ?", shortCode).Count(&totalVisits)

	var uniqueVisitors int64
	r.db.Model(&model.VisitRecord{}).Where("short_code = ?", shortCode).Distinct("ip_address").Count(&uniqueVisitors)

	summary := &model.AnalyticsSummary{
		TotalVisits:    totalVisits,
		UniqueVisitors: uniqueVisitors,
		TopCountries:  make(map[string]int),
		TopDevices:    make(map[string]int),
		TopBrowsers:   make(map[string]int),
		TopOS:         make(map[string]int),
		DailyVisits:   make(map[string]int),
		HourlyVisits:  make(map[int]int),
		TopReferrers:  make(map[string]int),
	}

	return summary, nil
}

func (r *AnalyticsRepository) GetRecentVisits(shortCode string, limit int, since *time.Time) ([]*model.VisitRecord, error) {
	var visits []*model.VisitRecord
	query := r.db.Where("short_code = ?", shortCode)
	if since != nil {
		query = query.Where("visited_at >= ?", *since)
	}
	err := query.Order("visited_at DESC").Limit(limit).Find(&visits).Error
	return visits, err
}
