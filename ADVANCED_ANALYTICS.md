# 高级统计分析功能

## 新增API接口

### 获取高级分析数据
```
GET /api/analytics/{code}
```

获取指定短链接的详细分析数据，包括：
- 总访问量和独立访客数
- 地理分布统计（国家/城市）
- 设备类型分布（桌面/移动/平板）
- 浏览器分布统计
- 操作系统分布统计
- 每日访问趋势
- 每小时访问分布
- 引荐来源统计

**查询参数:**
- `since`: 开始日期 (格式: YYYY-MM-DD)
- `until`: 结束日期 (格式: YYYY-MM-DD)

**示例请求:**
```
GET /api/analytics/abc123?since=2026-01-01&until=2026-02-02
```

**示例响应:**
```json
{
  "total_visits": 1250,
  "unique_visitors": 890,
  "top_countries": {
    "China": 450,
    "United States": 230,
    "Japan": 120
  },
  "top_devices": {
    "Mobile": 750,
    "Desktop": 400,
    "Tablet": 100
  },
  "top_browsers": {
    "Chrome": 800,
    "Safari": 300,
    "Firefox": 100
  },
  "top_os": {
    "Windows 10": 400,
    "Android": 500,
    "iOS": 300
  },
  "daily_visits": {
    "2026-01-15": 45,
    "2026-01-16": 67,
    "2026-01-17": 89
  },
  "hourly_visits": {
    "9": 45,
    "10": 67,
    "11": 89,
    "14": 120
  },
  "top_referrers": {
    "google.com": 200,
    "facebook.com": 150,
    "twitter.com": 80
  },
  "visit_timeline": [
    {
      "date": "2026-01-15",
      "count": 45
    },
    {
      "date": "2026-01-16",
      "count": 67
    }
  ]
}
```

### 获取最近访问记录
```
GET /api/visits/{code}
```

获取指定短链接的最近访问记录详情。

**查询参数:**
- `since`: 开始日期 (格式: YYYY-MM-DD)
- `limit`: 返回记录数量上限 (默认: 100)

**示例请求:**
```
GET /api/visits/abc123?since=2026-01-01&limit=50
```

## 新增功能特性

### 1. 地理位置追踪
- 自动识别访问者的地理位置（国家/城市）
- 基于IP地址的地理定位（需要集成第三方服务）

### 2. 设备类型检测
- 智能识别访问设备类型（桌面/移动/平板）
- 支持主流设备类型的精确分类

### 3. 浏览器和操作系统统计
- 识别访问者使用的浏览器和操作系统
- 支持主流浏览器和操作系统的识别

### 4. 访问来源追踪
- 记录访问来源（Referer）
- 统计各渠道的流量来源

### 5. 时间维度分析
- 按日期统计访问量（日活趋势）
- 按小时统计访问量（时段分析）

## 数据库变更

新增 `visit_records` 表用于存储详细的访问记录：

```sql
CREATE TABLE visit_records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  short_code TEXT NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  referer TEXT,
  country TEXT,
  city TEXT,
  user_os TEXT,
  browser TEXT,
  device_type TEXT,
  visited_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 隐私说明

- 所有访问数据均用于统计分析目的
- 不存储个人身份信息
- IP地址仅用于地理位置和访问统计，不做长期保存
- 遵循隐私保护最佳实践

## 集成IP地理位置服务

当前实现包含IP地理位置解析的框架，但需要集成第三方服务以获得准确的位置信息。推荐的服务包括：
- MaxMind GeoIP2
- IPinfo
- ip-api.com
- 其他商业或开源IP地理位置数据库