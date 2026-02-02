package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
	"url-shortener/internal/utils"
)

const (
	Base62Chars             = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	DefaultShortCodeLength  = 6
	MaxRetries              = 10
)

type EnhancedShortenerService struct {
	repo           *repository.URLRepository
	analyticsRepo  *repository.AnalyticsRepository
	analyticsSvc   *AnalyticsService
	baseURL        string
	mutex          sync.Mutex // 用于保护生成唯一短码的过程
}

func NewEnhancedShortenerService(
	repo *repository.URLRepository, 
	analyticsRepo *repository.AnalyticsRepository, 
	baseURL string) *EnhancedShortenerService {
	
	analyticsSvc := NewAnalyticsService(repo, analyticsRepo)
	
	return &EnhancedShortenerService{
		repo:          repo,
		analyticsRepo: analyticsRepo,
		analyticsSvc:  analyticsSvc,
		baseURL:       baseURL,
	}
}

func (s *EnhancedShortenerService) CreateShortURL(originalURL string, customCode string, expireInHours int) (*model.CreateURLResponse, error) {
	var shortCode string

	// 如果提供了自定义短码，验证并使用它
	if customCode != "" {
		// 验证自定义短码
		if !utils.IsValidCustomCode(customCode) {
			return nil, utils.ErrInvalidCustomCode
		}

		// 检查自定义短码是否已被使用
		_, err := s.repo.GetByShortCode(customCode)
		if err == nil {
			// 如果没有报错，说明短码已存在
			return nil, utils.ErrCustomCodeExists
		} else if err != utils.ErrURLNotFound {
			// 如果是其他错误，返回错误
			return nil, fmt.Errorf("failed to check if custom code exists: %w", err)
		}

		shortCode = customCode
	} else {
		// 生成随机短码
		var err error
		shortCode, err = s.generateUniqueShortCode()
		if err != nil {
			return nil, err
		}
	}

	// 计算过期时间
	var expiresAt *time.Time
	if expireInHours > 0 {
		expireTime := time.Now().Add(time.Duration(expireInHours) * time.Hour)
		expiresAt = &expireTime
	}

	// 保存到数据库
	err := s.repo.CreateWithExpiry(originalURL, shortCode, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL with expiry: %w", err)
	}

	// 构建响应
	response := &model.CreateURLResponse{
		ShortURL:  s.baseURL + "/" + shortCode,
		Code:      shortCode,
		Original:  originalURL,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if expiresAt != nil {
		response.ExpiresAt = expiresAt.Format(time.RFC3339)
	}

	return response, nil
}

func (s *EnhancedShortenerService) GetByShortCode(shortCode string) (*model.URL, error) {
	url, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// 检查链接是否已过期
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, utils.ErrURLExpired
	}

	// 增加点击次数（异步执行以提高性能）
	go func() {
		_ = s.repo.IncrementClicks(shortCode)
	}()

	return url, nil
}

// GetByShortCodeWithContext 通过上下文获取短链接（用于记录分析数据）
func (s *EnhancedShortenerService) GetByShortCodeWithContext(ctx context.Context, shortCode, ipAddress, userAgent, referer string) (*model.URL, error) {
	url, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// 检查链接是否已过期
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, utils.ErrURLExpired
	}

	// 记录访问分析数据
	go func() {
		_ = s.analyticsSvc.RecordVisit(ctx, shortCode, ipAddress, userAgent, referer)
	}()

	// 增加点击次数（异步执行以提高性能）
	go func() {
		_ = s.repo.IncrementClicks(shortCode)
	}()

	return url, nil
}

func (s *EnhancedShortenerService) GetStats(shortCode string) (*model.StatsResponse, error) {
	url, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// 检查链接是否已过期
	isActive := url.IsActive
	if url.ExpiresAt != nil {
		isActive = time.Now().Before(*url.ExpiresAt) && url.IsActive
	}

	stats := &model.StatsResponse{
		OriginalURL: url.OriginalURL,
		ShortCode:   url.ShortCode,
		Clicks:      url.Clicks,
		CreatedAt:   url.CreatedAt.Format(time.RFC3339),
		IsActive:    isActive,
	}

	if url.ExpiresAt != nil {
		stats.ExpiresAt = url.ExpiresAt.Format(time.RFC3339)
	}

	return stats, nil
}

// GetAdvancedAnalytics 获取高级分析数据
func (s *EnhancedShortenerService) GetAdvancedAnalytics(shortCode string, since *time.Time, until *time.Time) (*model.AnalyticsSummary, error) {
	return s.analyticsSvc.GetAnalyticsSummary(shortCode, since, until)
}

// GetRecentVisits 获取最近访问记录
func (s *EnhancedShortenerService) GetRecentVisits(shortCode string, limit int, since *time.Time) ([]*model.VisitRecord, error) {
	return s.analyticsSvc.GetRecentVisits(shortCode, limit, since)
}

func (s *EnhancedShortenerService) generateUniqueShortCode() (string, error) {
	// 使用互斥锁确保并发安全
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := 0; i < MaxRetries; i++ {
		shortCode := s.generateRandomString(DefaultShortCodeLength)

		// 检查短码是否已存在
		_, err := s.repo.GetByShortCode(shortCode)
		if err != nil {
			if err == utils.ErrURLNotFound {
				// 如果是 ErrURLNotFound 错误，说明短码不存在，可以使用
				return shortCode, nil
			}
			// 如果是其他错误，返回错误
			return "", fmt.Errorf("failed to check if short code exists: %w", err)
		}
		// 如果没有错误，说明短码已存在，继续循环
	}

	return "", utils.ErrGenerateShortCode
}

func (s *EnhancedShortenerService) generateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(Base62Chars))))
		if err != nil {
			// 如果随机数生成失败，回退到伪随机
			// 这里使用 math/rand 是为了确保始终有返回值
			result[i] = Base62Chars[i%len(Base62Chars)]
		} else {
			result[i] = Base62Chars[num.Int64()]
		}
	}
	return string(result)
}

func (s *EnhancedShortenerService) GetAllURLs() ([]*model.URL, error) {
	return s.repo.GetAll()
}

func (s *EnhancedShortenerService) DeleteShortCode(shortCode string) error {
	return s.repo.DeleteByShortCode(shortCode)
}

// CleanupExpiredURLs 清理过期链接的方法
func (s *EnhancedShortenerService) CleanupExpiredURLs() error {
	return s.repo.DeleteExpiredURLs()
}