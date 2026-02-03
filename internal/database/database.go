package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DBType 定义支持的数据库类型
type DBType string

const (
	DBTypeSQLite    DBType = "sqlite"
	DBTypeMySQL     DBType = "mysql"
	DBTypePostgreSQL DBType = "postgres"
)

// ParseDBType 从连接字符串解析数据库类型
func ParseDBType(databaseURL string) DBType {
	// 移除空格并转换为小写
	lowerURL := strings.ToLower(strings.TrimSpace(databaseURL))

	// 根据协议前缀判断
	if strings.HasPrefix(lowerURL, "mysql:") {
		return DBTypeMySQL
	}
	if strings.HasPrefix(lowerURL, "postgres:") || strings.HasPrefix(lowerURL, "postgresql:") {
		return DBTypePostgreSQL
	}
	// 默认使用 SQLite（用于本地文件路径）
	return DBTypeSQLite
}

// GetDriverName 获取数据库驱动名称
func GetDriverName(dbType DBType) string {
	switch dbType {
	case DBTypeMySQL:
		return "mysql"
	case DBTypePostgreSQL:
		return "postgres"
	default:
		return "sqlite3"
	}
}

// NewConnection 创建数据库连接
func NewConnection(databaseURL string) (*sql.DB, error) {
	dbType := ParseDBType(databaseURL)
	driverName := GetDriverName(dbType)

	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60 * 1000) // 5分钟

	return db, nil
}
