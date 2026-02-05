# URL Shortener Service

一个简单高效的 URL 短链接服务，可以将长 URL 转换为短链接，并提供访问统计功能。

## 功能特性

- 🚀 快速 URL 短链接生成
- 🔐 **API Key 认证** - 保护 API 访问安全
- 📊 访问统计跟踪
- 📈 高级统计分析（地理分布、设备类型、浏览器统计、访问来源等）
- 🔗 一键重定向
- 🛡️ **多数据库支持** - SQLite、MySQL、PostgreSQL
- 🌐 RESTful API 接口
- ⏰ 链接有效期控制（24小时、7天、30天等）
- 🎯 自定义短码功能（支持字母、数字、下划线、连字符）
- 🗑️ 自动清理过期链接
- 🔍 访问者分析（IP地址、地理位置、设备信息、浏览器类型等）
- 📅 时间维度分析（每日、每小时访问趋势）
- 🔒 统一错误处理和输入验证
- 🛡️ 并发安全保证
- 📋 健康检查端点

## 技术栈

- **语言**: Go
- **Web 框架**: Gin
- **数据库**: SQLite / MySQL / PostgreSQL
- **编码**: Base62

## 快速开始

### 环境要求

- Go 1.23+
- SQLite (默认), MySQL 5.7+ 或 PostgreSQL 12+

### 安装依赖

```bash
go mod tidy
```

### 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## 数据库配置

### 默认配置 (SQLite)

默认使用 SQLite 数据库，无需额外配置：

```bash
export DATABASE_URL="./urls.db"
# 或使用默认路径
```

### MySQL 配置

使用 MySQL 数据库，设置 `DATABASE_URL` 环境变量：

```bash
export DATABASE_URL="mysql://username:password@host:port/database?parseTime=true"
# 简化格式也支持：
export DATABASE_URL="username:password@tcp(host:port)/database?parseTime=true"
```

示例：

```bash
export DATABASE_URL="root:secret@tcp(localhost:3306)/urlshortener?parseTime=true"
go run cmd/server/main.go
```

### PostgreSQL 配置

使用 PostgreSQL 数据库：

```bash
export DATABASE_URL="postgres://username:password@host:port/database?sslmode=disable"
# 简化格式也支持：
export DATABASE_URL="user=username password=password host=host port=5432 dbname=database sslmode=disable"
```

示例：

```bash
export DATABASE_URL="postgres://postgres:secret@localhost:5432/urlshortener?sslmode=disable"
go run cmd/server/main.go
```

### Docker 部署

```bash
# MySQL
docker run -d -p 8080:8080 \
  -e DATABASE_URL="root:secret@tcp(mysql:3306)/urlshortener?parseTime=true" \
  -e BASE_URL=http://your-domain.com \
  url-shortener

# PostgreSQL
docker run -d -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:secret@postgresql:5432/urlshortener?sslmode=disable" \
  -e BASE_URL=http://your-domain.com \
  url-shortener
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | 服务端口 | 8080 |
| `DATABASE_URL` | 数据库连接字符串 | `./urls.db` (SQLite) |
| `BASE_URL` | 基础URL，用于生成短链接 | `http://localhost:8080` |
| `DEBUG` | 调试模式 | false |

## API 接口

### 创建短链接（需要 API Key）
```
POST /api/shorten
Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxx
Content-Type: application/json

{
  "url": "https://www.example.com/very/long/url",
  "custom_code": "mycode",      // 可选：自定义短码
  "expire_in": 24               // 可选：过期时间（小时），0表示永不过期
}
```

响应：
```json
{
  "short_url": "http://localhost:8080/a1b2c3",
  "code": "a1b2c3",
  "original": "https://www.example.com/very/long/url",
  "created_at": "2026-01-27T10:00:00Z",
  "expires_at": "2026-01-28T10:00:00Z"
}
```

### 重定向到原始链接（公开访问）
```
GET /{short_code}
```

### 获取短链接统计信息（需要 API Key）
```
GET /api/stats/{short_code}
Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxx
```

响应：
```json
{
  "original_url": "https://www.example.com/very/long/url",
  "short_code": "a1b2c3",
  "clicks": 150,
  "created_at": "2026-01-27T10:00:00Z",
  "expires_at": "2026-01-28T10:00:00Z",
  "is_active": true
}
```

### 获取高级分析数据（需要 API Key）
```
GET /api/analytics/{short_code}[?since=2026-01-01&until=2026-01-31]
Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxx
```

响应：
```json
{
  "total_visits": 150,
  "unique_visitors": 120,
  "geographic_distribution": {
    "China": 80,
    "United States": 40,
    "Other": 30
  },
  "device_types": {
    "mobile": 90,
    "desktop": 50,
    "tablet": 10
  },
  "daily_trend": [
    {
      "date": "2026-01-27",
      "visits": 25
    }
  ]
}
```

### 获取最近访问记录（需要 API Key）
```
GET /api/visits/{short_code}?limit=50&since=2026-01-01
Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxx
```

响应：
```json
{
  "visits": [
    {
      "timestamp": "2026-01-27T10:00:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "referer": "https://www.google.com"
    }
  ]
}
```

### 健康检查（公开访问）
```
GET /health
```

响应：
```json
{
  "status": "healthy",
  "message": "URL shortener service is running",
  "timestamp": 1706323200,
  "database": "mysql"
}
```

### 🔑 API Key 管理

#### 创建 API Key（无需认证）
```
POST /api/keys
Content-Type: application/json

{
  "name": "my-key-name",        // 必填：密钥名称
  "expires_in": 30              // 可选：过期天数，0表示永不过期
}
```

响应：
```json
{
  "message": "API key created successfully",
  "data": {
    "id": 1,
    "key": "sk_c6e7248365ff24d6f323b296a0ee60c931f7a47905fdb05d62fd564c9a621c5b",
    "name": "my-key-name",
    "created_at": "2026-02-02T22:00:00Z",
    "expires_at": "2026-03-04T22:00:00Z",
    "is_active": true
  }
}
```

⚠️ **注意：** `key` 字段只在创建时返回一次，请妥善保管！

#### 验证 API Key（无需认证）
```
GET /api/keys/validate?key=sk_xxxxxxxxxxxxxxxxxxxx
```

响应：
```json
{
  "valid": true,
  "data": {
    "name": "my-key-name",
    "created_at": "2026-02-02T22:00:00Z",
    "last_used": "2026-02-02T22:30:00Z"
  }
}
```

#### 列出所有 API Keys（需要 API Key）
```
GET /api/keys
Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxx
```

#### 撤销 API Key（需要 API Key）
```
DELETE /api/keys/{key}
Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxx
```

## 项目结构

```
url-shortener/
├── cmd/
│   └── server/
│       └── main.go             # 应用入口点
├── internal/
│   ├── database/               # 数据库抽象层（新增）
│   │   ├── database.go         # 数据库连接管理
│   │   └── dialect.go          # SQL 方言适配器
│   ├── model/                  # 数据模型定义
│   │   ├── url.go              # URL实体定义
│   │   ├── analytics.go        # 分析数据模型
│   │   └── apikey.go           # API Key 模型
│   ├── service/                # 业务逻辑层
│   │   ├── shortener.go        # 基础短链接服务
│   │   ├── enhanced_shortener.go # 增强短链接服务
│   │   ├── analytics_service.go # 分析服务
│   │   └── apikey_service.go   # API Key 服务
│   ├── handler/                # HTTP处理器
│   │   ├── handler.go          # 基础处理器
│   │   ├── enhanced_handler.go # 增强处理器
│   │   └── apikey_handler.go   # API Key 处理器
│   ├── repository/             # 数据访问层
│   │   ├── url_repo.go         # URL数据访问
│   │   ├── analytics_repo.go   # 分析数据访问
│   │   └── apikey_repo.go      # API Key 访问
│   ├── middleware/             # 中间件
│   │   └── auth.go             # API Key 认证中间件
│   ├── utils/                  # 工具函数
│   │   ├── errors.go           # 错误定义和处理
│   │   ├── validation.go       # 输入验证
│   │   ├── user_agent_parser.go # 用户代理解析
│   │   └── response.go         # 统一响应格式
│   └── config/                 # 配置管理
│       └── config.go           # 应用配置
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 多数据库支持实现

本项目使用数据库抽象层来支持多种数据库：

### 数据库类型检测

系统会自动根据 `DATABASE_URL` 的格式检测数据库类型：

- 以 `mysql:` 或 `tcp(` 开头 → MySQL
- 以 `postgres:` 或 `postgresql:` 开头 → PostgreSQL
- 其他情况（本地文件路径） → SQLite

### 方言适配器

不同数据库的 SQL 语法差异由 `Dialect` 接口处理：

| 特性 | SQLite | MySQL | PostgreSQL |
|------|--------|-------|------------|
| 占位符 | `?` | `?` | `$1`, `$2`... |
| 自增字段 | AUTOINCREMENT | AUTO_INCREMENT | SERIAL |
| 布尔类型 | INTEGER | TINYINT(1) | BOOLEAN |
| 日期函数 | DATE() | DATE() | DATE() |
| 时间提取 | strftime() | HOUR() | EXTRACT() |

### 添加新数据库支持

要支持新的数据库，只需：

1. 添加对应的驱动 import
2. 实现 `Dialect` 接口
3. 更新 `ParseDBType()` 函数

## Docker 完整部署示例

### 使用 MySQL

```bash
# 1. 创建网络
docker network create url-shortener-network

# 2. 启动 MySQL
docker run -d \
  --name mysql \
  --network url-shortener-network \
  -e MYSQL_ROOT_PASSWORD=secret \
  -e MYSQL_DATABASE=urlshortener \
  mysql:8

# 3. 启动应用
docker run -d \
  --name url-shortener \
  --network url-shortener-network \
  -p 8080:8080 \
  -e DATABASE_URL="root:secret@tcp(mysql:3306)/urlshortener?parseTime=true" \
  -e BASE_URL=http://localhost:8080 \
  url-shortener
```

### 使用 PostgreSQL

```bash
# 1. 创建网络
docker network create url-shortener-network

# 2. 启动 PostgreSQL
docker run -d \
  --name postgresql \
  --network url-shortener-network \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=urlshortener \
  postgres:15

# 3. 启动应用
docker run -d \
  --name url-shortener \
  --network url-shortener-network \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:secret@postgresql:5432/urlshortener?sslmode=disable" \
  -e BASE_URL=http://localhost:8080 \
  url-shortener
```

## 安全考虑

- 输入验证：所有输入都会经过严格验证
- 短码生成：使用加密安全的随机数生成器
- 速率限制：防止滥用
- SQL 注入防护：使用参数化查询
- XSS 防护：输出转义
- **API Key 认证**：保护敏感 API 端点

## 性能优化

- 数据库索引：为常用查询字段建立索引
- 连接池：使用数据库连接池
- 异步操作：点击计数等非关键操作异步执行
- 缓存：热点数据缓存

## 错误处理

- 统一错误响应格式
- 详细的错误日志
- 优雅的错误恢复

## 许可证

MIT License
