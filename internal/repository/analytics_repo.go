package repository

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"url-shortener/internal/database"
	"url-shortener/internal/model"
)

type AnalyticsRepository struct {
	db      *sql.DB
	dialect database.Dialect
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	dbType := database.ParseDBType("")
	dialect := database.GetDialect(dbType)

	repo := &AnalyticsRepository{
		db:      db,
		dialect: dialect,
	}
	repo.initDB()
	return repo
}

func NewAnalyticsRepositoryWithDialect(db *sql.DB, dbType database.DBType) *AnalyticsRepository {
	dialect := database.GetDialect(dbType)

	repo := &AnalyticsRepository{
		db:      db,
		dialect: dialect,
	}
	repo.initDB()
	return repo
}

func (r *AnalyticsRepository) initDB() error {
	ifNotExists := r.dialect.GetIfNotExists()
	dateTimeType := r.dialect.GetDateTimeType()

	query := fmt.Sprintf(`
		CREATE TABLE %s visit_records (
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
			visited_at %s DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_visit_short_code ON visit_records(short_code);
		CREATE INDEX IF NOT EXISTS idx_visit_visited_at ON visit_records(visited_at);
		CREATE INDEX IF NOT EXISTS idx_visit_ip_address ON visit_records(ip_address);
	`, ifNotExists, dateTimeType)

	_, err := r.db.Exec(query)
	return err
}

// RecordVisit 记录一次访问
func (r *AnalyticsRepository) RecordVisit(record *model.VisitRecord) error {
	placeholders := database.BuildPlaceholders(r.dialect, 10)
	query := fmt.Sprintf(`INSERT INTO visit_records (short_code, ip_address, user_agent, referer, country, city, user_os, browser, device_type, visited_at) VALUES (%s)`, placeholders)
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
	p1 := r.dialect.GetPlaceholder(0)
	query := fmt.Sprintf(`SELECT COUNT(*) as total_visits FROM visit_records WHERE short_code = %s`, p1)
	args := []interface{}{shortCode}

	// 添加时间范围条件
	if since != nil {
		pN := r.dialect.GetPlaceholder(len(args))
		query += fmt.Sprintf(` AND visited_at >= %s`, pN)
		args = append(args, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(args))
		query += fmt.Sprintf(` AND visited_at <= %s`, pN)
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
	uniqueQuery := fmt.Sprintf(`SELECT COUNT(DISTINCT ip_address) FROM visit_records WHERE short_code = %s`, p1)
	uniqueArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(uniqueArgs))
		uniqueQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		uniqueArgs = append(uniqueArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(uniqueArgs))
		uniqueQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
		uniqueArgs = append(uniqueArgs, until)
	}

	var uniqueVisitors int64
	err = r.db.QueryRow(uniqueQuery, uniqueArgs...).Scan(&uniqueVisitors)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique visitors: %w", err)
	}
	summary.UniqueVisitors = uniqueVisitors

	// 获取按国家统计
	countryQuery := fmt.Sprintf(`SELECT country, COUNT(*) as count FROM visit_records WHERE short_code = %s AND country IS NOT NULL`, p1)
	countryArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(countryArgs))
		countryQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		countryArgs = append(countryArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(countryArgs))
		countryQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
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
	deviceQuery := fmt.Sprintf(`SELECT device_type, COUNT(*) as count FROM visit_records WHERE short_code = %s AND device_type IS NOT NULL`, p1)
	deviceArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(deviceArgs))
		deviceQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		deviceArgs = append(deviceArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(deviceArgs))
		deviceQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
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
	browserQuery := fmt.Sprintf(`SELECT browser, COUNT(*) as count FROM visit_records WHERE short_code = %s AND browser IS NOT NULL`, p1)
	browserArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(browserArgs))
		browserQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		browserArgs = append(browserArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(browserArgs))
		browserQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
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
	osQuery := fmt.Sprintf(`SELECT user_os, COUNT(*) as count FROM visit_records WHERE short_code = %s AND user_os IS NOT NULL`, p1)
	osArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(osArgs))
		osQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		osArgs = append(osArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(osArgs))
		osQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
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
	dateFunc := r.dialect.GetDateFunction("visited_at")
	dailyQuery := fmt.Sprintf(`SELECT %s as date, COUNT(*) as count FROM visit_records WHERE short_code = %s`, dateFunc, p1)
	dailyArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(dailyArgs))
		dailyQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		dailyArgs = append(dailyArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(dailyArgs))
		dailyQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
		dailyArgs = append(dailyArgs, until)
	}
	dailyQuery += ` GROUP BY date ORDER BY date`

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
	hourFunc := r.dialect.GetDateHourFunction("visited_at")
	hourlyQuery := fmt.Sprintf(`SELECT %s as hour, COUNT(*) as count FROM visit_records WHERE short_code = %s`, hourFunc, p1)
	hourlyArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(hourlyArgs))
		hourlyQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		hourlyArgs = append(hourlyArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(hourlyArgs))
		hourlyQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
		hourlyArgs = append(hourlyArgs, until)
	}
	hourlyQuery += ` GROUP BY hour ORDER BY hour`

	rows, err = r.db.Query(hourlyQuery, hourlyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var hour interface{}
		var count int
		err := rows.Scan(&hour, &count)
		if err != nil {
			continue
		}

		var h int
		switch v := hour.(type) {
		case int64:
			h = int(v)
		case int32:
			h = int(v)
		case int:
			h = v
		case string:
			fmt.Sscanf(v, "%d", &h)
		default:
			h = 0
		}
		summary.HourlyVisits[h] = count
	}

	// 获取顶级引荐来源
	referrerQuery := fmt.Sprintf(`SELECT referer, COUNT(*) as count FROM visit_records WHERE short_code = %s AND referer IS NOT NULL AND referer != ''`, p1)
	referrerArgs := []interface{}{shortCode}

	if since != nil {
		pN := r.dialect.GetPlaceholder(len(referrerArgs))
		referrerQuery += fmt.Sprintf(` AND visited_at >= %s`, pN)
		referrerArgs = append(referrerArgs, since)
	}
	if until != nil {
		pN := r.dialect.GetPlaceholder(len(referrerArgs))
		referrerQuery += fmt.Sprintf(` AND visited_at <= %s`, pN)
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
				parts := strings.SplitN(referer, "/", 3)
				if len(parts) >= 2 {
					host := parts[2]
					if idx := strings.Index(host, "?"); idx > 0 {
						host = host[:idx]
					}
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
	p1 := r.dialect.GetPlaceholder(0)
	p2 := r.dialect.GetPlaceholder(1)

	query := fmt.Sprintf(`SELECT id, short_code, ip_address, user_agent, referer, country, city, user_os, browser, device_type, visited_at FROM visit_records WHERE short_code = %s`, p1)
	args := []interface{}{shortCode}

	if since != nil {
		query += fmt.Sprintf(` AND visited_at >= %s`, p2)
		args = append(args, since)
	}

	query += fmt.Sprintf(` ORDER BY visited_at DESC LIMIT %s`, r.dialect.GetPlaceholder(len(args)))
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent visits: %w", err)
	}
	defer rows.Close()

	var visits []*model.VisitRecord
	for rows.Next() {
		var visit model.VisitRecord
		var visitedAt interface{}

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
			&visitedAt,
		)
		if err != nil {
			continue
		}

		// 解析时间
		visit.VisitedAt = r.parseVisitedAt(visitedAt)

		visits = append(visits, &visit)
	}

	return visits, nil
}

func (r *AnalyticsRepository) parseVisitedAt(value interface{}) time.Time {
	if value == nil {
		return time.Time{}
	}

	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
		return time.Time{}
	default:
		return time.Time{}
	}
}
