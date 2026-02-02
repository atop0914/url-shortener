package repository

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"
	"url-shortener/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	repo := &AnalyticsRepository{db: db}
	repo.initDB()
	return repo
}

func (r *AnalyticsRepository) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS visit_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		referer TEXT,
		country TEXT,
		city TEXT,
		user_os TEXT,
		browser TEXT,
		device_type TEXT,
		visited_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_visit_short_code ON visit_records(short_code);
	CREATE INDEX IF NOT EXISTS idx_visit_visited_at ON visit_records(visited_at);
	CREATE INDEX IF NOT EXISTS idx_visit_ip_address ON visit_records(ip_address);
	`
	_, err := r.db.Exec(query)
	return err
}

// RecordVisit 记录一次访问
func (r *AnalyticsRepository) RecordVisit(record *model.VisitRecord) error {
	query := `INSERT INTO visit_records (short_code, ip_address, user_agent, referer, country, city, user_os, browser, device_type, visited_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, 
		record.ShortCode, 
		record.IPAddress, 
		record.UserAgent, 
		record.Referer, 
		record.Country, 
		record.City, 
		record.UserOS, 
		record.Browser, 
		record.DeviceType, 
		record.VisitedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record visit: %w", err)
	}
	return nil
}

// GetAnalyticsSummary 获取统计摘要
func (r *AnalyticsRepository) GetAnalyticsSummary(shortCode string, since *time.Time, until *time.Time) (*model.AnalyticsSummary, error) {
	summary := &model.AnalyticsSummary{
		TopCountries:  make(map[string]int),
		TopDevices:    make(map[string]int),
		TopBrowsers:   make(map[string]int),
		TopOS:         make(map[string]int),
		DailyVisits:   make(map[string]int),
		HourlyVisits:  make(map[int]int),
		TopReferrers:  make(map[string]int),
		VisitTimeline: []model.TimelinePoint{},
	}

	// 构建基础查询
	query := `SELECT COUNT(*) as total_visits FROM visit_records WHERE short_code = ?`
	args := []interface{}{shortCode}

	// 添加时间范围条件
	if since != nil {
		query += ` AND visited_at >= ?`
		args = append(args, since)
	}
	if until != nil {
		query += ` AND visited_at <= ?`
		args = append(args, until)
	}

	// 获取总访问量
	var totalVisits int64
	err := r.db.QueryRow(query, args...).Scan(&totalVisits)
	if err != nil {
		return nil, fmt.Errorf("failed to get total visits: %w", err)
	}
	summary.TotalVisits = totalVisits

	// 获取唯一访客数量
	uniqueQuery := `SELECT COUNT(DISTINCT ip_address) FROM visit_records WHERE short_code = ?`
	uniqueArgs := []interface{}{shortCode}
	
	if since != nil {
		uniqueQuery += ` AND visited_at >= ?`
		uniqueArgs = append(uniqueArgs, since)
	}
	if until != nil {
		uniqueQuery += ` AND visited_at <= ?`
		uniqueArgs = append(uniqueArgs, until)
	}

	var uniqueVisitors int64
	err = r.db.QueryRow(uniqueQuery, uniqueArgs...).Scan(&uniqueVisitors)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique visitors: %w", err)
	}
	summary.UniqueVisitors = uniqueVisitors

	// 获取按国家统计
	countryQuery := `SELECT country, COUNT(*) as count FROM visit_records WHERE short_code = ? AND country IS NOT NULL`
	countryArgs := []interface{}{shortCode}
	
	if since != nil {
		countryQuery += ` AND visited_at >= ?`
		countryArgs = append(countryArgs, since)
	}
	if until != nil {
		countryQuery += ` AND visited_at <= ?`
		countryArgs = append(countryArgs, until)
	}
	countryQuery += ` GROUP BY country ORDER BY count DESC LIMIT 10`

	rows, err := r.db.Query(countryQuery, countryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get country stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var country string
		var count int
		err := rows.Scan(&country, &count)
		if err != nil {
			continue
		}
		if country != "" {
			summary.TopCountries[country] = count
		}
	}

	// 获取按设备类型统计
	deviceQuery := `SELECT device_type, COUNT(*) as count FROM visit_records WHERE short_code = ? AND device_type IS NOT NULL`
	deviceArgs := []interface{}{shortCode}
	
	if since != nil {
		deviceQuery += ` AND visited_at >= ?`
		deviceArgs = append(deviceArgs, since)
	}
	if until != nil {
		deviceQuery += ` AND visited_at <= ?`
		deviceArgs = append(deviceArgs, until)
	}
	deviceQuery += ` GROUP BY device_type ORDER BY count DESC LIMIT 10`

	rows, err = r.db.Query(deviceQuery, deviceArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get device stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var deviceType string
		var count int
		err := rows.Scan(&deviceType, &count)
		if err != nil {
			continue
		}
		if deviceType != "" {
			summary.TopDevices[deviceType] = count
		}
	}

	// 获取按浏览器统计
	browserQuery := `SELECT browser, COUNT(*) as count FROM visit_records WHERE short_code = ? AND browser IS NOT NULL`
	browserArgs := []interface{}{shortCode}
	
	if since != nil {
		browserQuery += ` AND visited_at >= ?`
		browserArgs = append(browserArgs, since)
	}
	if until != nil {
		browserQuery += ` AND visited_at <= ?`
		browserArgs = append(browserArgs, until)
	}
	browserQuery += ` GROUP BY browser ORDER BY count DESC LIMIT 10`

	rows, err = r.db.Query(browserQuery, browserArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get browser stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var browser string
		var count int
		err := rows.Scan(&browser, &count)
		if err != nil {
			continue
		}
		if browser != "" {
			summary.TopBrowsers[browser] = count
		}
	}

	// 获取按操作系统统计
	osQuery := `SELECT user_os, COUNT(*) as count FROM visit_records WHERE short_code = ? AND user_os IS NOT NULL`
	osArgs := []interface{}{shortCode}
	
	if since != nil {
		osQuery += ` AND visited_at >= ?`
		osArgs = append(osArgs, since)
	}
	if until != nil {
		osQuery += ` AND visited_at <= ?`
		osArgs = append(osArgs, until)
	}
	osQuery += ` GROUP BY user_os ORDER BY count DESC LIMIT 10`

	rows, err = r.db.Query(osQuery, osArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get OS stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userOS string
		var count int
		err := rows.Scan(&userOS, &count)
		if err != nil {
			continue
		}
		if userOS != "" {
			summary.TopOS[userOS] = count
		}
	}

	// 获取每日访问统计
	dailyQuery := `SELECT DATE(visited_at) as date, COUNT(*) as count FROM visit_records WHERE short_code = ?`
	dailyArgs := []interface{}{shortCode}
	
	if since != nil {
		dailyQuery += ` AND visited_at >= ?`
		dailyArgs = append(dailyArgs, since)
	}
	if until != nil {
		dailyQuery += ` AND visited_at <= ?`
		dailyArgs = append(dailyArgs, until)
	}
	dailyQuery += ` GROUP BY DATE(visited_at) ORDER BY date`

	rows, err = r.db.Query(dailyQuery, dailyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var count int
		err := rows.Scan(&date, &count)
		if err != nil {
			continue
		}
		summary.DailyVisits[date] = count
		summary.VisitTimeline = append(summary.VisitTimeline, model.TimelinePoint{
			Date:  date,
			Count: count,
		})
	}

	// 获取每小时访问统计
	hourlyQuery := `SELECT strftime('%H', visited_at) as hour, COUNT(*) as count FROM visit_records WHERE short_code = ?`
	hourlyArgs := []interface{}{shortCode}
	
	if since != nil {
		hourlyQuery += ` AND visited_at >= ?`
		hourlyArgs = append(hourlyArgs, since)
	}
	if until != nil {
		hourlyQuery += ` AND visited_at <= ?`
		hourlyArgs = append(hourlyArgs, until)
	}
	hourlyQuery += ` GROUP BY strftime('%H', visited_at) ORDER BY hour`

	rows, err = r.db.Query(hourlyQuery, hourlyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var hourStr string
		var count int
		err := rows.Scan(&hourStr, &count)
		if err != nil {
			continue
		}
		hour := 0
		fmt.Sscanf(hourStr, "%d", &hour) // 将字符串转换为整数
		summary.HourlyVisits[hour] = count
	}

	// 获取顶级引荐来源
	referrerQuery := `SELECT referer, COUNT(*) as count FROM visit_records WHERE short_code = ? AND referer IS NOT NULL AND referer != ''`
	referrerArgs := []interface{}{shortCode}
	
	if since != nil {
		referrerQuery += ` AND visited_at >= ?`
		referrerArgs = append(referrerArgs, since)
	}
	if until != nil {
		referrerQuery += ` AND visited_at <= ?`
		referrerArgs = append(referrerArgs, until)
	}
	referrerQuery += ` GROUP BY referer ORDER BY count DESC LIMIT 10`

	rows, err = r.db.Query(referrerQuery, referrerArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get referrer stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var referer string
		var count int
		err := rows.Scan(&referer, &count)
		if err != nil {
			continue
		}
		if referer != "" {
			// 简化URL，只保留域名部分
			parsedURL, err := url.Parse(referer)
			if err == nil && parsedURL.Host != "" {
				summary.TopReferrers[parsedURL.Host] = count
			} else {
				// 如果解析失败，使用原始值的前缀
				host := strings.Split(referer, "/")[2]
				if len(host) > 0 {
					summary.TopReferrers[host] = count
				} else {
					summary.TopReferrers[referer] = count
				}
			}
		}
	}

	return summary, nil
}

// GetRecentVisits 获取最近访问记录
func (r *AnalyticsRepository) GetRecentVisits(shortCode string, limit int, since *time.Time) ([]*model.VisitRecord, error) {
	query := `SELECT id, short_code, ip_address, user_agent, referer, country, city, user_os, browser, device_type, visited_at FROM visit_records WHERE short_code = ?`
	args := []interface{}{shortCode}

	if since != nil {
		query += ` AND visited_at >= ?`
		args = append(args, since)
	}

	query += ` ORDER BY visited_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent visits: %w", err)
	}
	defer rows.Close()

	var visits []*model.VisitRecord
	for rows.Next() {
		var visit model.VisitRecord
		var visitedAtStr string
		
		err := rows.Scan(
			&visit.ID,
			&visit.ShortCode,
			&visit.IPAddress,
			&visit.UserAgent,
			&visit.Referer,
			&visit.Country,
			&visit.City,
			&visit.UserOS,
			&visit.Browser,
			&visit.DeviceType,
			&visitedAtStr,
		)
		if err != nil {
			continue
		}

		// 解析时间
		visitedAt, err := time.Parse("2006-01-02 15:04:05", visitedAtStr)
		if err == nil {
			visit.VisitedAt = visitedAt
		}

		visits = append(visits, &visit)
	}

	return visits, nil
}