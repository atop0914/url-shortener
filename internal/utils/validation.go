package utils

import (
	"regexp"
)

const (
	MinCustomCodeLength = 3
	MaxCustomCodeLength = 20
)

// IsValidCustomCode 验证自定义短码是否符合要求
func IsValidCustomCode(code string) bool {
	if len(code) < MinCustomCodeLength || len(code) > MaxCustomCodeLength {
		return false
	}

	// 检查是否只包含字母、数字和连字符
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validPattern.MatchString(code)
}