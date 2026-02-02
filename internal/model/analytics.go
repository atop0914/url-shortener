package model

import "time"

// VisitRecord 存储每次访问的详细信息
type VisitRecord struct {
	ID           int64     `json:"id" db:"id"`
	ShortCode    string    `json:"short_code" db:"short_code"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	Referer      string    `json:"referer" db:"referer"`
	Country      string    `json:"country,omitempty" db:"country"`
	City         string    `json:"city,omitempty" db:"city"`
	UserOS       string    `json:"user_os,omitempty" db:"user_os"`
	Browser      string    `json:"browser,omitempty" db:"browser"`
	DeviceType   string    `json:"device_type,omitempty" db:"device_type"`
	VisitedAt    time.Time `json:"visited_at" db:"visited_at"`
}

// AnalyticsSummary 统计摘要
type AnalyticsSummary struct {
	TotalVisits     int64                    `json:"total_visits"`
	UniqueVisitors  int64                    `json:"unique_visitors"`
	TopCountries    map[string]int           `json:"top_countries"`
	TopDevices      map[string]int           `json:"top_devices"`
	TopBrowsers     map[string]int           `json:"top_browsers"`
	TopOS           map[string]int           `json:"top_os"`
	DailyVisits     map[string]int           `json:"daily_visits"` // YYYY-MM-DD as key
	HourlyVisits    map[int]int              `json:"hourly_visits"` // 0-23 as key
	TopReferrers    map[string]int           `json:"top_referrers"`
	VisitTimeline   []TimelinePoint          `json:"visit_timeline"`
}

// TimelinePoint 时间线数据点
type TimelinePoint struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"`
}

// VisitAnalyticsRequest 请求体结构
type VisitAnalyticsRequest struct {
	ShortCode string `uri:"code" binding:"required"`
	Since     string `form:"since"`     // 格式: YYYY-MM-DD
	Until     string `form:"until"`     // 格式: YYYY-MM-DD
	Limit     int    `form:"limit,default=100"`
}