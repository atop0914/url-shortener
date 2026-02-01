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
	base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCodeLength = 6
	maxRetries = 10
)

type ShortenerService struct {
	repo *repository.URLRepository
	baseURL string
}

func NewShortenerService(repo *repository.URLRepository, baseURL string) *ShortenerService {
	return &ShortenerService{
		repo: repo,
		baseURL: baseURL,
	}
}

func (s *ShortenerService) CreateShortURL(originalURL string) (*model.CreateURLResponse, error) {
	// 检查原 URL 是否已存在
	urls, err := s.repo.GetAll()
	if err == nil {
		for _, url := range urls {
			if url.OriginalURL == originalURL {
				return &model.CreateURLResponse{
					ShortURL: s.baseURL + "/" + url.ShortCode,
					Code:     url.ShortCode,
					Original: originalURL,
					CreatedAt: url.CreatedAt.Format(time.RFC3339),
				}, nil
			}
		}
	}

	// 生成短码
	shortCode, err := s.generateUniqueShortCode()
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	err = s.repo.Create(originalURL, shortCode)
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

	return response, nil
}

func (s *ShortenerService) GetByShortCode(shortCode string) (*model.URL, error) {
	url, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
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

	stats := &model.StatsResponse{
		OriginalURL: url.OriginalURL,
		ShortCode:   url.ShortCode,
		Clicks:      url.Clicks,
		CreatedAt:   url.CreatedAt.Format(time.RFC3339),
	}

	return stats, nil
}

func (s *ShortenerService) generateUniqueShortCode() (string, error) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < maxRetries; i++ {
		shortCode := s.generateRandomString(shortCodeLength)
		
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