package utils

import (
	"time"
)

// ParseTime 解析时间字符串，支持多种格式
// 返回 time.Time，未解析成功返回零值
func ParseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	
	return time.Time{}
}

// ParseTimePtr 解析时间字符串，支持多种格式
// 返回 *time.Time，成功解析返回指针，未成功返回 nil
func ParseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	
	t := ParseTime(s)
	if t.IsZero() {
		return nil
	}
	
	return &t
}

// ParseTimeNullable 解析可能为 null 的时间字符串
// 返回 *time.Time，支持多种输入类型
func ParseTimeNullable(s interface{}) *time.Time {
	if s == nil {
		return nil
	}
	
	switch v := s.(type) {
	case time.Time:
		return &v
	case string:
		return ParseTimePtr(v)
	case []byte:
		return ParseTimePtr(string(v))
	default:
		return nil
	}
}
