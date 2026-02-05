package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Validator 输入验证器
type Validator struct {
	errors map[string]string
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

// HasErrors 检查是否有验证错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors 获取所有验证错误
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// AddError 添加验证错误
func (v *Validator) AddError(field, message string) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = message
	}
}

// Check 检查条件并添加错误
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

// In 检查值是否在允许的值列表中
func In(value string, list []string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}

// MinLength 检查字符串最小长度
func MinLength(value string, n int) bool {
	return len(value) >= n
}

// MaxLength 检查字符串最大长度
func MaxLength(value string, n int) bool {
	return len(value) <= n
}

// IsValidEmail 检查是否为有效邮箱
func IsValidEmail(email string) bool {
	const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailPattern).MatchString(email)
}

// IsValidURL 检查是否为有效URL
func IsValidURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme != "" && parsed.Host != ""
}

// IsValidCustomCode 检查自定义短码格式
func IsValidCustomCode(code string) bool {
	// 自定义短码只能包含字母、数字、下划线和连字符，长度3-32
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]{3,32}$`, code)
	return err == nil && matched
}

// IsValidShortCode 检查短码格式
func IsValidShortCode(code string) bool {
	// 短码只能包含字母、数字，长度至少2个字符
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]{2,}$`, code)
	return err == nil && matched
}

// SanitizeCustomCode 清理自定义短码
func SanitizeCustomCode(code string) string {
	// 移除不安全的字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := reg.ReplaceAllString(code, "")
	
	// 限制长度
	if len(sanitized) > 32 {
		sanitized = sanitized[:32]
	} else if len(sanitized) < 3 {
		sanitized = ""
	}
	
	return sanitized
}

// IsValidInteger 检查是否为有效的整数
func IsValidInteger(s string) bool {
	if s == "" {
		return false
	}
	
	// 检查是否只包含数字和可选的负号
	re := regexp.MustCompile(`^-?\d+$`)
	return re.MatchString(s)
}

// IsValidPositiveInteger 检查是否为正整数
func IsValidPositiveInteger(s string) bool {
	if !IsValidInteger(s) {
		return false
	}
	
	num := strings.TrimLeft(s, "-")
	if num == "" {
		return false
	}
	
	// 检查是否为正数（不能以0开头，除非就是0）
	if len(num) > 1 && num[0] == '0' {
		return false
	}
	
	return true
}

// ValidateURLShortenRequest 验证URL缩短请求
func ValidateURLShortenRequest(longURL, customCode string, expireInHours int) *Validator {
	v := NewValidator()
	
	// 验证长URL
	v.Check(IsValidURL(longURL), "url", "invalid URL format")
	
	// 验证自定义短码
	if customCode != "" {
		v.Check(IsValidCustomCode(customCode), "custom_code", "invalid custom code format. Must contain only letters, numbers, underscores, and hyphens, 3-32 characters")
	}
	
	// 验证过期时间
	v.Check(expireInHours >= 0 && expireInHours <= 8760, "expire_in", "expire_in must be between 0 and 8760 hours (1 year)")
	
	return v
}

// ValidateAPIKeyRequest 验证API密钥创建请求
func ValidateAPIKeyRequest(name string, expiresAt *int) *Validator {
	v := NewValidator()
	
	// 验证名称
	v.Check(name != "", "name", "name is required")
	v.Check(MinLength(name, 1), "name", "name must be at least 1 character")
	v.Check(MaxLength(name, 100), "name", "name must be no more than 100 characters")
	
	// 验证过期时间
	if expiresAt != nil {
		v.Check(*expiresAt >= 0 && *expiresAt <= 365, "expires_in", "expires_in must be between 0 and 365 days")
	}
	
	return v
}

// ValidatePaginationParams 验证分页参数
func ValidatePaginationParams(page, pageSize int) *Validator {
	v := NewValidator()
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	
	return v
}

// SanitizeAndValidateURL 清理和验证URL
func SanitizeAndValidateURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}
	
	// 去除首尾空白
	rawURL = strings.TrimSpace(rawURL)
	
	// 如果没有协议，添加默认协议
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	
	// 验证URL格式
	if !IsValidURL(rawURL) {
		return "", fmt.Errorf("invalid URL format: %s", rawURL)
	}
	
	return rawURL, nil
}

// ValidateAndFormatError 格式化验证错误
func ValidateAndFormatError(err error) string {
	if err == nil {
		return ""
	}
	
	// 这里可以根据具体的验证错误格式化输出
	// Gin框架通常会提供 validator.ValidationErrors 类型的错误
	// 为了通用性，我们直接返回错误信息
	return err.Error()
}