package gorm

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"url-shortener/internal/model"
)

type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase() (*Database, error) {
	dsn := os.Getenv("DATABASE_URL")
	
	var db *gorm.DB
	var err error

	if dsn == "" {
		// 默认使用 SQLite
		log.Println("使用 SQLite 数据库")
		db, err = gorm.Open(sqlite.Open("urls.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else if contains(dsn, "mysql") || contains(dsn, "tcp(") {
		// MySQL
		log.Println("使用 MySQL 数据库")
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else if contains(dsn, "postgres") {
		// PostgreSQL
		log.Println("使用 PostgreSQL 数据库")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// 默认使用 SQLite
		log.Println("使用 SQLite 数据库")
		db, err = gorm.Open(sqlite.Open("urls.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&model.URL{}, &model.APIKey{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Database{DB: db}, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

// GetDB 获取数据库连接
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping 检查数据库连接
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// URLRepository URL 仓储
type URLRepository struct {
	DB *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{DB: db}
}

// Create 创建短链接
func (r *URLRepository) Create(url *model.URL) error {
	return r.DB.Create(url).Error
}

// GetByShortCode 根据短码获取
func (r *URLRepository) GetByShortCode(code string) (*model.URL, error) {
	var url model.URL
	err := r.DB.Where("short_code = ? AND is_active = ?", code, true).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// IncrementClicks 增加点击数
func (r *URLRepository) IncrementClicks(code string) error {
	return r.DB.Model(&model.URL{}).Where("short_code = ?", code).UpdateColumn("clicks", gorm.Expr("clicks + ?", 1)).Error
}

// GetAll 获取所有活跃的短链接
func (r *URLRepository) GetAll() ([]model.URL, error) {
	var urls []model.URL
	err := r.DB.Where("is_active = ?", true).Order("created_at DESC").Find(&urls).Error
	return urls, err
}

// DeleteExpiredURLs 删除过期链接
func (r *URLRepository) DeleteExpiredURLs() error {
	return r.DB.Model(&model.URL{}).Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", time.Now(), true).Update("is_active", false).Error
}

// APIKeyRepository API Key 仓储
type APIKeyRepository struct {
	DB *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{DB: db}
}

// Create 创建 API Key
func (r *APIKeyRepository) Create(key *model.APIKey) error {
	return r.DB.Create(key).Error
}

// GetByKey 根据 Key 获取
func (r *APIKeyRepository) GetByKey(key string) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := r.DB.Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// ValidateKey 验证 Key 是否有效
func (r *APIKeyRepository) ValidateKey(key string) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := r.DB.Where("key = ? AND is_active = ?", key, true).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	// 检查是否过期
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}
	// 更新最后使用时间
	r.DB.Model(&apiKey).Update("last_used", time.Now())
	return &apiKey, nil
}

// GetAll 获取所有 API Keys
func (r *APIKeyRepository) GetAll() ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.DB.Order("created_at DESC").Find(&keys).Error
	return keys, err
}

// Deactivate 禁用 API Key
func (r *APIKeyRepository) Deactivate(key string) error {
	return r.DB.Model(&model.APIKey{}).Where("key = ?", key).Update("is_active", false).Error
}
