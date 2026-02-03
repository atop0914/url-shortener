package utils

import (
	"errors"
	"fmt"
)

// AppError 应用错误结构
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 实现错误包装接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建新的应用错误
func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WrapError 包装错误
func WrapError(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsAppError 检查是否为应用错误
func IsAppError(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// 预定义错误
var (
	ErrNotFound          = NewAppError("NOT_FOUND", "resource not found")
	ErrInvalidInput      = NewAppError("INVALID_INPUT", "invalid input provided")
	ErrUnauthorized      = NewAppError("UNAUTHORIZED", "unauthorized access")
	ErrForbidden         = NewAppError("FORBIDDEN", "forbidden access")
	ErrInternal          = NewAppError("INTERNAL_ERROR", "internal server error")
	ErrRateLimitExceeded = NewAppError("RATE_LIMIT_EXCEEDED", "rate limit exceeded")
	ErrDatabaseError     = NewAppError("DATABASE_ERROR", "database error occurred")
	ErrAlreadyExists     = NewAppError("ALREADY_EXISTS", "resource already exists")
	ErrExpired           = NewAppError("EXPIRED", "resource has expired")
	ErrURLNotFound       = NewAppError("URL_NOT_FOUND", "short URL not found")
	ErrURLExpired        = NewAppError("URL_EXPIRED", "short URL has expired")
	ErrInvalidCustomCode = NewAppError("INVALID_CUSTOM_CODE", "invalid custom code format")
	ErrCustomCodeExists  = NewAppError("CUSTOM_CODE_EXISTS", "custom code already exists")
	ErrGenerateShortCode = NewAppError("GENERATE_SHORT_CODE_FAILED", "failed to generate unique short code after multiple attempts")
)