package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
	"url-shortener/internal/utils"
)

type ShortenerService struct {
	repo    *repository.URLRepository
	baseURL string
	mutex   sync.Mutex // 用于保护生成唯一短码的过程
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

func (s *ShortenerService) GetByShortCode(shortCode string) (*model.URL, error) {
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
		// 为了防止并发问题，这里可以考虑使用更精细的锁
		// 或者使用数据库的原子更新操作
		_ = s.repo.IncrementClicks(shortCode)
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
	// 使用互斥锁确保并发安全
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := 0; i < MaxRetries; i++ {
		shortCode, err := s.generateRandomString(DefaultShortCodeLength)
		if err != nil {
			// 如果随机数生成失败，记录错误并继续尝试
			continue
		}

		// 检查短码是否已存在
		_, err = s.repo.GetByShortCode(shortCode)
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

func (s *ShortenerService) generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(Base62Chars))))
		if err != nil {
			// 随机数生成失败，返回错误而不是使用可预测的回退
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}
		result[i] = Base62Chars[num.Int64()]
	}
	return string(result), nil
}

func (s *ShortenerService) GetAllURLs() ([]*model.URL, error) {
	return s.repo.GetAll()
}

func (s *ShortenerService) DeleteShortCode(shortCode string) error {
	return s.repo.DeleteByShortCode(shortCode)
}

// CleanupExpiredURLs 清理过期链接的方法
func (s *ShortenerService) CleanupExpiredURLs() error {
	return s.repo.DeleteExpiredURLs()
}