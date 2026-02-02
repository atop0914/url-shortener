package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
	"url-shortener/internal/utils"
)

type AnalyticsService struct {
	urlRepo         *repository.URLRepository
	analyticsRepo   *repository.AnalyticsRepository
}

func NewAnalyticsService(urlRepo *repository.URLRepository, analyticsRepo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
	}
}

// RecordVisit 记录访问事件
func (s *AnalyticsService) RecordVisit(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error {
	// 获取真实IP地址（处理代理情况）
	realIP := s.getRealIP(ctx, ipAddress)

	// 解析用户代理信息
	userOS, browser, deviceType := utils.ParseUserAgent(userAgent)

	// TODO: 实现IP地理位置解析
	// 这里只是一个模拟实现，实际部署时可以集成真实的IP地理位置服务
	country, city := s.getLocationFromIP(realIP)

	// 创建访问记录
	visitRecord := &model.VisitRecord{
		ShortCode:  shortCode,
		IPAddress:  realIP,
		UserAgent:  userAgent,
		Referer:    referer,
		Country:    country,
		City:       city,
		UserOS:     userOS,
		Browser:    browser,
		DeviceType: deviceType,
		VisitedAt:  time.Now(),
	}

	// 保存访问记录
	err := s.analyticsRepo.RecordVisit(visitRecord)
	if err != nil {
		return fmt.Errorf("failed to record visit: %w", err)
	}

	return nil
}

// GetAnalyticsSummary 获取统计摘要
func (s *AnalyticsService) GetAnalyticsSummary(shortCode string, since *time.Time, until *time.Time) (*model.AnalyticsSummary, error) {
	return s.analyticsRepo.GetAnalyticsSummary(shortCode, since, until)
}

// GetRecentVisits 获取最近访问记录
func (s *AnalyticsService) GetRecentVisits(shortCode string, limit int, since *time.Time) ([]*model.VisitRecord, error) {
	return s.analyticsRepo.GetRecentVisits(shortCode, limit, since)
}

// getRealIP 获取真实IP地址（处理代理情况）
func (s *AnalyticsService) getRealIP(ctx context.Context, ipAddress string) string {
	// 在上下文中可能有更准确的IP信息
	// 这里简单处理，实际实现可能需要检查X-Forwarded-For、X-Real-IP等头部
	if forwardedFor := ctx.Value("X-Forwarded-For"); forwardedFor != nil {
		if forwardedStr, ok := forwardedFor.(string); ok {
			ips := strings.Split(forwardedStr, ",")
			if len(ips) > 0 {
				return strings.TrimSpace(ips[0])
			}
		}
	}

	if realIP := ctx.Value("X-Real-IP"); realIP != nil {
		if realIPStr, ok := realIP.(string); ok {
			return realIPStr
		}
	}

	return ipAddress
}

// getLocationFromIP 从IP获取地理位置信息（模拟实现）
func (s *AnalyticsService) getLocationFromIP(ip string) (country, city string) {
	// 这里只是一个模拟实现
	// 实际部署时应该使用真实的IP地理位置服务，如：
	// - MaxMind GeoIP2
	// - IPinfo
	// - ip-api.com
	// - 其他第三方服务
	
	// 对于私有IP地址，返回特殊标记
	if s.isPrivateIP(ip) {
		return "Local", "Private Network"
	}
	
	// 这里可以集成真实的地理位置服务
	// 目前返回未知
	return "Unknown", "Unknown"
}

// isPrivateIP 检查是否为私有IP地址
func (s *AnalyticsService) isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // 如果无法解析，假设是私有的
	}
	
	// 检查是否为私有IP段
	privateCIDRs := []string{
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918
		"192.168.0.0/16", // RFC 1918
		"127.0.0.0/8",    // localhost
		"::1/128",        // IPv6 localhost
	}
	
	for _, cidr := range privateCIDRs {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(parsedIP) {
			return true
		}
	}
	
	return false
}