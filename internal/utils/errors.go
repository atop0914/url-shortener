package utils

import (
	"errors"
	"fmt"
)

// 定义自定义错误类型
var (
	ErrURLNotFound         = errors.New("URL not found")
	ErrURLExpired          = errors.New("URL has expired")
	ErrCustomCodeExists    = errors.New("custom short code already exists")
	ErrInvalidCustomCode   = errors.New("invalid custom short code format")
	ErrGenerateShortCode   = errors.New("failed to generate unique short code after maximum retries")
	ErrInvalidIPFormat     = errors.New("invalid IP address format")
	ErrInvalidTimeFormat   = errors.New("invalid time format")
)

// ValidationError 表示验证错误
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error 实现 error 接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError 创建一个新的验证错误
func NewValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ValidateAndFormatError 验证并格式化错误信息
func ValidateAndFormatError(err error) string {
	if err == nil {
		return ""
	}
	
	// 检查是否是预定义错误
	switch {
	case errors.Is(err, ErrURLNotFound):
		return ErrURLNotFound.Error()
	case errors.Is(err, ErrURLExpired):
		return ErrURLExpired.Error()
	case errors.Is(err, ErrCustomCodeExists):
		return ErrCustomCodeExists.Error()
	case errors.Is(err, ErrInvalidCustomCode):
		return ErrInvalidCustomCode.Error()
	case errors.Is(err, ErrGenerateShortCode):
		return ErrGenerateShortCode.Error()
	default:
		// 对于其他错误，返回原始错误信息
		return err.Error()
	}
}