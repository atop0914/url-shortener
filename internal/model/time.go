package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// JSONTime 自定义时间类型，自动处理 SQLite TEXT 格式
type JSONTime time.Time

// Scan 实现 sql.Scanner 接口
func (jt *JSONTime) Scan(value interface{}) error {
	if value == nil {
		*jt = JSONTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*jt = JSONTime(v)
		return nil
	case []byte:
		// 尝试多种格式解析
		formats := []string{
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, string(v)); err == nil {
				*jt = JSONTime(t)
				return nil
			}
		}
		return fmt.Errorf("unable to parse time: %s", string(v))
	case string:
		formats := []string{
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, v); err == nil {
				*jt = JSONTime(t)
				return nil
			}
		}
		return fmt.Errorf("unable to parse time: %s", v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

// Value 实现 driver.Valuer 接口
func (jt JSONTime) Value() (driver.Value, error) {
	if time.Time(jt).IsZero() {
		return nil, nil
	}
	return time.Time(jt).Format("2006-01-02 15:04:05"), nil
}

// String 转换为字符串
func (jt JSONTime) String() string {
	return time.Time(jt).Format("2006-01-02 15:04:05")
}

// MarshalJSON 实现 JSON 序列化
func (jt JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(jt).Format("2006-01-02 15:04:05") + `"`), nil
}
