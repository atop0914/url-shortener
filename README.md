# URL Shortener Service

一个简单高效的 URL 短链接服务，可以将长 URL 转换为短链接，并提供访问统计功能。

## 功能特性

- 🚀 快速 URL 短链接生成
- 📊 访问统计跟踪
- 🔗 一键重定向
- 🛡️ SQLite 数据库存储
- 🌐 RESTful API 接口

## 技术栈

- **语言**: Go
- **Web 框架**: Gin
- **数据库**: SQLite
- **编码**: Base62

## API 接口

### 创建短链接
```
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url"
}
```

**响应**:
```json
{
  "short_url": "http://localhost:8080/abc123",
  "code": "abc123",
  "original": "https://example.com/very/long/url",
  "created_at": "2026-02-01T11:30:00Z"
}
```

### 重定向
```
GET /{short_code}
```
重定向到原始 URL。

### 查看统计
```
GET /api/stats/{code}
```

**响应**:
```json
{
  "original_url": "https://example.com/...",
  "short_code": "abc123",
  "clicks": 42,
  "created_at": "2026-02-01T11:30:00Z"
}
```

## 部署

### 环境变量

- `PORT`: 服务端口 (默认: 8080)
- `DB_PATH`: 数据库路径 (默认: ./urls.db)
- `BASE_URL`: 基础 URL (默认: http://localhost:8080)

### 本地运行

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

## 示例

```bash
# 创建短链接
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url"}'

# 访问短链接
curl http://localhost:8080/abc123

# 查看统计
curl http://localhost:8080/api/stats/abc123
```

## 项目结构

```
url-shortener/
├── cmd/server/
│   └── main.go          # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go    # 配置管理
│   ├── handler/
│   │   └── handler.go   # HTTP 处理器
│   ├── model/
│   │   └── url.go       # 数据模型
│   ├── repository/
│   │   └── url_repo.go  # 数据库操作
│   └── service/
│       └── shortener.go # 业务逻辑
├── go.mod
└── README.md
```

## 许可证

MIT