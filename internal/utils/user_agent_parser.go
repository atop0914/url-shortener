package utils

import (
	"regexp"
	"strings"
)

// UserAgentInfo 用户代理信息结构
type UserAgentInfo struct {
	Browser   string `json:"browser"`
	Version   string `json:"version"`
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	Device    string `json:"device"`
	DeviceType string `json:"device_type"`
}

// ParseUserAgent 解析用户代理字符串
func ParseUserAgent(userAgent string) *UserAgentInfo {
	if userAgent == "" {
		return &UserAgentInfo{
			Browser:    "Unknown",
			OS:         "Unknown",
			DeviceType: "Desktop",
		}
	}

	info := &UserAgentInfo{
		Browser:  "Unknown",
		OS:       "Unknown",
		Device:   "Unknown",
		DeviceType: "Desktop", // 默认为桌面端
	}

	// 解析浏览器和版本
	info.Browser, info.Version = parseBrowser(userAgent)

	// 解析操作系统和版本
	info.OS, info.OSVersion = parseOS(userAgent)

	// 解析设备类型
	info.DeviceType = parseDeviceType(userAgent)

	// 设备品牌/型号（简化版）
	info.Device = parseDeviceBrand(userAgent)

	return info
}

// parseBrowser 解析浏览器信息
func parseBrowser(userAgent string) (string, string) {
	// Chrome
	if strings.Contains(userAgent, "Chrome/") {
		parts := strings.Split(userAgent, "Chrome/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			return "Chrome", strings.Split(version, ".")[0] // 只返回主版本号
		}
		return "Chrome", ""
	}

	// Firefox
	if strings.Contains(userAgent, "Firefox/") {
		parts := strings.Split(userAgent, "Firefox/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			return "Firefox", strings.Split(version, ".")[0]
		}
		return "Firefox", ""
	}

	// Safari
	if strings.Contains(userAgent, "Safari/") && !strings.Contains(userAgent, "Chrome/") {
		parts := strings.Split(userAgent, "Safari/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			return "Safari", strings.Split(version, ".")[0]
		}
		return "Safari", ""
	}

	// Edge
	if strings.Contains(userAgent, "Edg/") || strings.Contains(userAgent, "Edge/") {
		var version string
		if strings.Contains(userAgent, "Edg/") {
			parts := strings.Split(userAgent, "Edg/")
			if len(parts) > 1 {
				version = strings.Split(parts[1], " ")[0]
			}
		} else {
			parts := strings.Split(userAgent, "Edge/")
			if len(parts) > 1 {
				version = strings.Split(parts[1], " ")[0]
			}
		}
		return "Edge", strings.Split(version, ".")[0]
	}

	// Internet Explorer
	if strings.Contains(userAgent, "MSIE") {
		parts := strings.Split(userAgent, "MSIE ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], ";")[0]
			return "Internet Explorer", strings.Split(version, ".")[0]
		}
		return "Internet Explorer", ""
	}

	return "Unknown", ""
}

// parseOS 解析操作系统信息
func parseOS(userAgent string) (string, string) {
	// Windows
	if strings.Contains(userAgent, "Windows NT") {
		parts := strings.Split(userAgent, "Windows NT ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], ";")[0]
			version = strings.Split(version, " ")[0]
			
			versions := map[string]string{
				"5.1": "XP",
				"6.0": "Vista",
				"6.1": "7",
				"6.2": "8",
				"6.3": "8.1",
				"10.0": "10",
				"11.0": "11",
			}
			
			osName := "Windows " + versions[version]
			if osName == "Windows " {
				osName = "Windows NT " + version
			}
			return osName, version
		}
		return "Windows", ""
	}

	// macOS
	if strings.Contains(userAgent, "Mac OS X") {
		parts := strings.Split(userAgent, "Mac OS X ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			version = strings.ReplaceAll(version, "_", ".")
			return "macOS", version
		}
		return "macOS", ""
	}

	// iOS
	if strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") {
		// 查找 OS 标识
		re := regexp.MustCompile(`OS (\d+)_(\d+)_?(\d+)?`)
		matches := re.FindStringSubmatch(userAgent)
		if len(matches) > 1 {
			version := matches[1] + "." + matches[2]
			if len(matches) > 3 && matches[3] != "" {
				version += "." + matches[3]
			}
			return "iOS", version
		}
		return "iOS", ""
	}

	// Android
	if strings.Contains(userAgent, "Android") {
		parts := strings.Split(userAgent, "Android ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			return "Android", version
		}
		return "Android", ""
	}

	// Linux
	if strings.Contains(userAgent, "Linux") {
		return "Linux", ""
	}

	return "Unknown", ""
}

// parseDeviceType 解析设备类型
func parseDeviceType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// 检查移动设备特征
	mobileIndicators := []string{
		"iphone", "ipad", "android", "mobile", "blackberry", "windows phone",
		"opera mini", "iemobile", "mobile safari", "phone",
	}

	for _, indicator := range mobileIndicators {
		if strings.Contains(userAgent, indicator) {
			// 检查是否是平板
			tabletIndicators := []string{
				"ipad", "tablet", "playbook", "silk-accelerated",
			}
			
			for _, tabletIndicator := range tabletIndicators {
				if strings.Contains(userAgent, tabletIndicator) {
					return "Tablet"
				}
			}
			
			return "Mobile"
		}
	}

	// 检查是否是机器人
	botIndicators := []string{
		"bot", "crawl", "spider", "slurp", "facebookexternalhit",
		"facebookplatform", "twitterbot", "linkedinbot", "embedly",
		"pinterest", "slackbot", "vkshare", "bingpreview", "tumblr",
	}

	for _, indicator := range botIndicators {
		if strings.Contains(userAgent, indicator) {
			return "Bot"
		}
	}

	return "Desktop"
}

// parseDeviceBrand 解析设备品牌
func parseDeviceBrand(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// 常见设备品牌
	brands := []string{
		"iphone", "ipad", "samsung", "galaxy", "nexus", "pixel",
		"htc", "lg", "motorola", "sony", "xiaomi", "huawei",
		"oppo", "vivo", "oneplus", "lenovo", "dell", "hp",
		"asus", "acer", "toshiba", "surface", "mac", "thinkpad",
	}

	for _, brand := range brands {
		if strings.Contains(userAgent, brand) {
			return strings.Title(brand)
		}
	}

	return "Unknown"
}

// GetUserDeviceType 便捷函数，只获取设备类型
func GetUserDeviceType(userAgent string) string {
	return parseDeviceType(userAgent)
}

// GetBrowserName 便捷函数，只获取浏览器名称
func GetBrowserName(userAgent string) string {
	browser, _ := parseBrowser(userAgent)
	return browser
}

// GetOSName 便捷函数，只获取操作系统名称
func GetOSName(userAgent string) string {
	os, _ := parseOS(userAgent)
	return os
}