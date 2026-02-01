# URL Shortener Service

一个简单高效的 URL 短链接服务，可以将长 URL 转换为短链接，并提供访问统计功能。

## 功能特性

- 🚀 快速 URL 短链接生成
- 📊 访问统计跟踪
- 🔗 一键重定向
- 🛡️ SQLite 数据库存储
- 🌐 RESTful API 接口
- ⏰ **新增**: 链接有效期控制（24小时、7天、30天等）
- 🎯 **新增**: 自定义短码功能（支持字母、数字、下划线、连字符）
- 🗑️ **新增**: 自动清理过期链接
- 🔒 **新增**: 统一错误处理和输入验证
- 🛡️ **新增**: 并发安全保证

## 技术栈

- **语言**: Go
- **Web 框架**: Gin
- **数据库**: SQLite
- **编码**: Base62

## API 接口

### 创建短链接（支持自定义短码和过期时间）
```
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_code": "mylink",        # 可选：自定义短码（3-20字符，支持字母、数字、下划线、连字符）
  "expire_in": 24                 # 可选：过期时间（小时），0表示不过期
}
```

**响应**:
```json
{
  "short_url": "http://localhost:8080/mylink",
  "code": "mylink",
  "original": "https://example.com/very/long/url",
  "created_at": "2026-02-01T11:30:00Z",
  "expires_at": "2026-02-02T11:30:00Z"  # 如果设置了过期时间
}
```

### 重定向
```
GET /{short_code}
```
重定向到原始 URL。如果链接已过期，返回 410 状态码。

### 查看统计
```
GET /api/stats/{code}
```

**响应**:
```json
{
  "original_url": "https://example.com/...",
  "short_code": "mylink",
  "clicks": 42,
  "created_at": "2026-02-01T11:30:00Z",
  "expires_at": "2026-02-02T11:30:00Z",  # 如果设置了过期时间
  "is_active": true
}
```

### 清理过期链接
```
POST /api/cleanup
```
手动清理所有已过期的链接。

## 部署

### 环境变量

- `PORT`: 服务端口 (默认: 8080)
- `DB_PATH`: 数据库路径 (默认: ./urls.db)
- `BASE_URL`: 基础 URL (默认: http://localhost:8080)
- `LOG_LEVEL`: 日志级别 (默认: info)

### 本地运行

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

## 示例

```bash
# 创建带自定义短码的链接
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url", "custom_code": "mylink"}'

# 创建带24小时过期时间的链接
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url", "expire_in": 24}'

# 创建带自定义短码和过期时间的链接
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url", "custom_code": "special", "expire_in": 168}'  # 168小时 = 7天

# 访问短链接
curl http://localhost:8080/mylink

# 查看统计
curl http://localhost:8080/api/stats/mylink

# 清理过期链接
curl -X POST http://localhost:8080/api/cleanup
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
│   ├── service/
│   │   └── shortener.go # 业务逻辑
│   └── utils/
│       ├── errors.go    # 统一错误定义
│       └── validation.go # 输入验证
├── go.mod
└── README.md
```

## 重构改进点

1. **错误处理**: 统一错误定义和处理
2. **输入验证**: 更严格的输入验证，防止无效的自定义短码
3. **并发安全**: 使用互斥锁保护关键操作
4. **代码结构**: 清晰的模块划分
5. **资源管理**: 确保数据库连接正确关闭
6. **代码质量**: 更好的注释和文档

## 许可证

MIT