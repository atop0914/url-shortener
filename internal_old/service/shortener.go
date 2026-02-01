package service

import (
	"errors"
	"math/rand"
	"strings"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

const (
	base62Chars       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	defaultShortCodeLength = 6
	minCustomCodeLength = 3
	maxCustomCodeLength = 20
	maxRetries        = 10
)

type ShortenerService struct {
	repo    *repository.URLRepository
	baseURL string
}

func NewShortenerService(repo *repository.URLRepository, baseURL string) *ShortenerService {
	return &ShortenerService{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *ShortenerService) CreateShortURL(originalURL string, customCode string, expireInHours int) (*model.CreateURLResponse, error) {
	var shortCode string
	
	// 如果提供了自定义短码，验证并使用它
	if customCode != "" {
		// 验证自定义短码
		if len(customCode) < minCustomCodeLength || len(customCode) > maxCustomCodeLength {
			return nil, errors.New("custom code length must be between 3 and 20 characters")
		}
		
		// 检查自定义短码是否已被使用
		_, err := s.repo.GetByShortCode(customCode)
		if err == nil {
			// 如果没有报错，说明短码已存在
			return nil, errors.New("custom code already exists, please choose another one")
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
		return nil, err
	}

	// 构建响应
	response := &model.CreateURLResponse{
		ShortURL: s.baseURL + "/" + shortCode,
		Code:     shortCode,
		Original: originalURL,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if expiresAt != nil {
		response.ExpiresAt = expiresAt.Format(time.RFC3339)
	}

	return response, nil
}

func (s *ShortenerService) GetByShortCode(shortCode string) (*model.URL, error) {
	url, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// 检查链接是否已过期
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		// 链接已过期，可以选择性地删除或标记为不活跃
		return nil, errors.New("link has expired")
	}

	// 增加点击次数
	go func() {
		s.repo.IncrementClicks(shortCode)
	}()

	return url, nil
}

func (s *ShortenerService) GetStats(shortCode string) (*model.StatsResponse, error) {
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

func (s *ShortenerService) generateUniqueShortCode() (string, error) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < maxRetries; i++ {
		shortCode := s.generateRandomString(defaultShortCodeLength)
		
		// 检查短码是否已存在
		_, err := s.repo.GetByShortCode(shortCode)
		if err != nil {
			// 如果是 ErrNoRows 错误，说明短码不存在，可以使用
			if strings.Contains(err.Error(), "no rows in result set") || 
				strings.Contains(err.Error(), "SQL logic error") {
				return shortCode, nil
			}
		}
		// 如果没有错误，说明短码已存在，继续循环
	}

	return "", errors.New("failed to generate unique short code after retries")
}

func (s *ShortenerService) generateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = base62Chars[rand.Intn(len(base62Chars))]
	}
	return string(result)
}

func (s *ShortenerService) GetAllURLs() ([]*model.URL, error) {
	return s.repo.GetAll()
}

func (s *ShortenerService) DeleteShortCode(shortCode string) error {
	return s.repo.DeleteByShortCode(shortCode)
}

// 清理过期链接的方法
func (s *ShortenerService) CleanupExpiredURLs() error {
	return s.repo.DeleteExpiredURLs()
}