package cache

import (
	"sync"
	"time"
)

// Cache 接口定义
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	store   map[string]*cacheItem
	mu      sync.RWMutex
	maxSize int
}

// cacheItem 缓存项
type cacheItem struct {
	value      interface{}
	expireAt   time.Time
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		store:   make(map[string]*cacheItem),
		maxSize: maxSize,
	}
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.store[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.expireAt) {
		delete(c.store, key)
		return nil, false
	}

	return item.value, true
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果超过最大容量，清理旧项
	if len(c.store) >= c.maxSize {
		c.evictOldest()
	}

	c.store[key] = &cacheItem{
		value:    value,
		expireAt: time.Now().Add(ttl),
	}
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]*cacheItem)
}

// evictOldest 清理最旧的缓存项
func (c *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.store {
		if oldestKey == "" || item.expireAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expireAt
		}
	}

	if oldestKey != "" {
		delete(c.store, oldestKey)
	}
}

// URLCache URL 专用缓存
type URLCache struct {
	cache     *MemoryCache
	defaultTTL time.Duration
}

// NewURLCache 创建 URL 缓存
func NewURLCache(maxEntries int, defaultTTL time.Duration) *URLCache {
	return &URLCache{
		cache:     NewMemoryCache(maxEntries),
		defaultTTL: defaultTTL,
	}
}

// Get 获取缓存的 URL
func (u *URLCache) Get(shortCode string) (*CachedURL, bool) {
	data, exists := u.cache.Get(shortCode)
	if !exists {
		return nil, false
	}

	if cached, ok := data.(*CachedURL); ok {
		return cached, true
	}

	return nil, false
}

// Set 缓存 URL
func (u *URLCache) Set(shortCode string, originalURL string, expiresAt *time.Time) {
	var ttl time.Duration
	if expiresAt != nil {
		ttl = time.Until(*expiresAt)
		if ttl > u.defaultTTL {
			ttl = u.defaultTTL
		}
	} else {
		ttl = u.defaultTTL
	}

	cached := &CachedURL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CachedAt:    time.Now(),
	}

	u.cache.Set(shortCode, cached, ttl)
}

// Invalidate 使缓存失效
func (u *URLCache) Invalidate(shortCode string) {
	u.cache.Delete(shortCode)
}

// Clear 清空缓存
func (u *URLCache) Clear() {
	u.cache.Clear()
}

// CachedURL 缓存的 URL 信息
type CachedURL struct {
	ShortCode   string
	OriginalURL string
	CachedAt    time.Time
}
