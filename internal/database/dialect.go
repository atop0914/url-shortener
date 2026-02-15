package database

import (
	"fmt"
	"strings"
)

// Dialect 提供特定数据库的 SQL 语法支持
type Dialect interface {
	// GetName 返回数据库类型名称
	GetName() DBType

	// GetPlaceholder 返回参数占位符
	GetPlaceholder(index int) string

	// GetAutoIncrement 返回自增字段定义
	GetAutoIncrement(columnName string) string

	// GetBooleanType 返回布尔类型定义
	GetBooleanType() string

	// GetDateTimeType 返回日期时间类型定义
	GetDateTimeType() string

	// GetIfNotExists 返回 IF NOT EXISTS 短语
	GetIfNotExists() string

	// GetDefaultNow 返回 CURRENT_TIMESTAMP 或等效表达式
	GetDefaultNow() string

	// GetDateFunction 返回提取日期的函数
	GetDateFunction(column string) string

	// GetDateHourFunction 返回提取小时的函数
	GetDateHourFunction(column string) string

	// GetCurrentTimestamp 返回当前时间戳函数
	GetCurrentTimestamp() string
}

// BaseDialect 基础方言
type BaseDialect struct {
	dbType DBType
}

func (d *BaseDialect) GetName() DBType {
	return d.dbType
}

func (d *BaseDialect) GetCurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// SQLiteDialect SQLite 方言
type SQLiteDialect struct {
	BaseDialect
}

func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{
		BaseDialect{dbType: DBTypeSQLite},
	}
}

func (d *SQLiteDialect) GetPlaceholder(index int) string {
	return "?"
}

func (d *SQLiteDialect) GetAutoIncrement(columnName string) string {
	return fmt.Sprintf("INTEGER PRIMARY KEY AUTOINCREMENT")
}

func (d *SQLiteDialect) GetBooleanType() string {
	return "INTEGER"
}

func (d *SQLiteDialect) GetDateTimeType() string {
	return "TEXT"
}

func (d *SQLiteDialect) GetIfNotExists() string {
	return "IF NOT EXISTS"
}

func (d *SQLiteDialect) GetDefaultNow() string {
	// SQLite 兼容 - 使用空字符串，让应用层插入时间
	return ""
}

func (d *SQLiteDialect) GetDateFunction(column string) string {
	return fmt.Sprintf("DATE(%s)", column)
}

func (d *SQLiteDialect) GetDateHourFunction(column string) string {
	return fmt.Sprintf("strftime('%%H', %s)", column)
}

// MySQLDialect MySQL 方言
type MySQLDialect struct {
	BaseDialect
}

func NewMySQLDialect() *MySQLDialect {
	return &MySQLDialect{
		BaseDialect{dbType: DBTypeMySQL},
	}
}

func (d *MySQLDialect) GetPlaceholder(index int) string {
	return "?"
}

func (d *MySQLDialect) GetAutoIncrement(columnName string) string {
	return fmt.Sprintf("INT AUTO_INCREMENT PRIMARY KEY")
}

func (d *MySQLDialect) GetBooleanType() string {
	return "TINYINT(1)"
}

func (d *MySQLDialect) GetDateTimeType() string {
	return "DATETIME"
}

func (d *MySQLDialect) GetIfNotExists() string {
	return "IF NOT EXISTS"
}

func (d *MySQLDialect) GetDefaultNow() string {
	return "CURRENT_TIMESTAMP"
}

func (d *MySQLDialect) GetDateFunction(column string) string {
	return fmt.Sprintf("DATE(%s)", column)
}

func (d *MySQLDialect) GetDateHourFunction(column string) string {
	return fmt.Sprintf("HOUR(%s)", column)
}

func (d *MySQLDialect) GetCurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// PostgreSQLDialect PostgreSQL 方言
type PostgreSQLDialect struct {
	BaseDialect
}

func NewPostgreSQLDialect() *PostgreSQLDialect {
	return &PostgreSQLDialect{
		BaseDialect{dbType: DBTypePostgreSQL},
	}
}

func (d *PostgreSQLDialect) GetPlaceholder(index int) string {
	return fmt.Sprintf("$%d", index+1)
}

func (d *PostgreSQLDialect) GetAutoIncrement(columnName string) string {
	// PostgreSQL 使用 IDENTITY 列
	return fmt.Sprintf("SERIAL PRIMARY KEY")
}

func (d *PostgreSQLDialect) GetBooleanType() string {
	return "BOOLEAN"
}

func (d *PostgreSQLDialect) GetDateTimeType() string {
	return "TIMESTAMP"
}

func (d *PostgreSQLDialect) GetIfNotExists() string {
	return "IF NOT EXISTS"
}

func (d *PostgreSQLDialect) GetDefaultNow() string {
	return "CURRENT_TIMESTAMP"
}

func (d *PostgreSQLDialect) GetDateFunction(column string) string {
	return fmt.Sprintf("DATE(%s)", column)
}

func (d *PostgreSQLDialect) GetDateHourFunction(column string) string {
	return fmt.Sprintf("EXTRACT(HOUR FROM %s)", column)
}

// GetDialect 根据数据库类型获取对应的方言
func GetDialect(dbType DBType) Dialect {
	switch dbType {
	case DBTypeMySQL:
		return NewMySQLDialect()
	case DBTypePostgreSQL:
		return NewPostgreSQLDialect()
	default:
		return NewSQLiteDialect()
	}
}

// BuildPlaceholders 构建多个占位符
func BuildPlaceholders(dialect Dialect, count int) string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = dialect.GetPlaceholder(i)
	}
	return strings.Join(placeholders, ", ")
}
