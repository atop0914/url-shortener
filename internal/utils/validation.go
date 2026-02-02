package utils

import (
	"regexp"
	"strings"
)

// IsValidCustomCode 验证自定义短码是否有效
// 短码只能包含字母、数字、下划线和连字符
func IsValidCustomCode(code string) bool {
	if code == "" {
		return false
	}
	
	// 检查长度
	if len(code) < 3 || len(code) > 32 {
		return false
	}
	
	// 使用正则表达式检查格式
	// 只允许字母、数字、下划线、连字符
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, code)
	if err != nil {
		return false
	}
	
	return matched
}

// SanitizeCustomCode 清理自定义短码
// 移除潜在的危险字符
func SanitizeCustomCode(code string) string {
	// 移除前后空格
	code = strings.TrimSpace(code)
	
	// 只保留允许的字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	code = reg.ReplaceAllString(code, "")
	
	return code
}

// IsValidURL 验证URL格式
// 这是一个简单的URL格式验证
func IsValidURL(rawURL string) bool {
	// 基本检查：非空、长度合理、包含协议和域名
	if rawURL == "" {
		return false
	}
	
	// 检查最小长度
	if len(rawURL) < 10 {
		return false
	}
	
	// 检查是否包含协议
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return false
	}
	
	// 简单的域名格式检查
	// 在协议后应该有域名部分
	parts := strings.SplitN(rawURL, "://", 2)
	if len(parts) < 2 {
		return false
	}
	
	domainPart := parts[1]
	
	// 检查是否包含至少一个点（典型的域名特征）
	if !strings.Contains(domainPart, ".") {
		return false
	}
	
	// 检查是否包含有效的字符
	matched, err := regexp.MatchString(`^[a-zA-Z0-9._~:/?#[\]@!$&'()*+,;=%-]+$`, domainPart)
	if err != nil {
		return false
	}
	
	return matched
}

// NormalizeURL 标准化URL
// 移除尾随斜杠等
func NormalizeURL(rawURL string) string {
	// 移除前后空格
	url := strings.TrimSpace(rawURL)
	
	// 确保有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	return url
}