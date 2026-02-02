package utils

import (
	"regexp"
	"strings"
)

// ParseUserAgent 解析用户代理字符串以获取设备类型、浏览器和操作系统信息
func ParseUserAgent(userAgent string) (os, browser, deviceType string) {
	userAgent = strings.ToLower(userAgent)

	// 操作系统检测
	os = detectOS(userAgent)
	
	// 浏览器检测
	browser = detectBrowser(userAgent)
	
	// 设备类型检测
	deviceType = detectDeviceType(userAgent, userAgent)

	return os, browser, deviceType
}

// detectOS 检测操作系统
func detectOS(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "windows nt 10.0"):
		return "Windows 10"
	case strings.Contains(userAgent, "windows nt 6.3"):
		return "Windows 8.1"
	case strings.Contains(userAgent, "windows nt 6.2"):
		return "Windows 8"
	case strings.Contains(userAgent, "windows nt 6.1"):
		return "Windows 7"
	case strings.Contains(userAgent, "windows nt 6.0"):
		return "Windows Vista"
	case strings.Contains(userAgent, "windows nt 5.1"):
		return "Windows XP"
	case strings.Contains(userAgent, "windows nt 5.0"):
		return "Windows 2000"
	case strings.Contains(userAgent, "mac os x"):
		return "Mac OS X"
	case strings.Contains(userAgent, "android"):
		return "Android"
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		return "iOS"
	case strings.Contains(userAgent, "linux"):
		return "Linux"
	case strings.Contains(userAgent, "ubuntu"):
		return "Ubuntu"
	case strings.Contains(userAgent, "freebsd"):
		return "FreeBSD"
	default:
		return "Unknown"
	}
}

// detectBrowser 检测浏览器
func detectBrowser(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "edge/") && strings.Contains(userAgent, "edg/"):
		return "Microsoft Edge"
	case strings.Contains(userAgent, "edge/") && !strings.Contains(userAgent, "edg/"):
		return "Edge Legacy"
	case strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "edg/"):
		return "Chrome"
	case strings.Contains(userAgent, "firefox/"):
		return "Firefox"
	case strings.Contains(userAgent, "safari/") && !strings.Contains(userAgent, "chrome"):
		return "Safari"
	case strings.Contains(userAgent, "opera/") || strings.Contains(userAgent, "opr/"):
		return "Opera"
	case strings.Contains(userAgent, "msie") || strings.Contains(userAgent, "trident"):
		return "Internet Explorer"
	default:
		return "Unknown"
	}
}

// detectDeviceType 检测设备类型
func detectDeviceType(userAgent, originalUA string) string {
	// 移动设备检测
	if isMobileDevice(userAgent) {
		return "Mobile"
	}
	
	// 平板设备检测
	if isTabletDevice(userAgent) {
		return "Tablet"
	}
	
	// 桌面设备检测
	if isDesktopDevice(userAgent) {
		return "Desktop"
	}
	
	// 默认分类
	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "phone") {
		return "Mobile"
	}
	
	if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "pad") {
		return "Tablet"
	}
	
	return "Desktop"
}

// isMobileDevice 检查是否为移动设备
func isMobileDevice(userAgent string) bool {
	mobileRegex := regexp.MustCompile(`(?i)(mobile|android|iphone|ipod|blackberry|iemobile|opera mini)`)
	return mobileRegex.MatchString(userAgent)
}

// isTabletDevice 检查是否为平板设备
func isTabletDevice(userAgent string) bool {
	tabletRegex := regexp.MustCompile(`(?i)(tablet|ipad|playbook|silk)`)
	return tabletRegex.MatchString(userAgent) && !strings.Contains(userAgent, "mobile")
}

// isDesktopDevice 检查是否为桌面设备
func isDesktopDevice(userAgent string) bool {
	// 如果不是移动或平板设备，则假定为桌面设备
	return !isMobileDevice(userAgent) && !isTabletDevice(userAgent)
}