# 更新日志

## v1.1.0 - 高级统计分析功能

### 新增功能
- 高级统计分析：地理位置追踪（国家/城市）
- 设备类型检测：自动识别桌面/移动/平板设备
- 浏览器和操作系统统计
- 访问来源（Referer）追踪
- 时间维度分析：每日和每小时访问趋势
- 最近访问记录查看
- 详细的访问分析API

### API变更
- 新增 `GET /api/analytics/{code}` - 获取高级分析数据
- 新增 `GET /api/visits/{code}` - 获取最近访问记录
- 扩展了原有的统计功能

### 数据库变更
- 新增 `visit_records` 表用于存储详细访问信息
- 包含IP地址、用户代理、地理位置、设备信息等字段

### 架构改进
- 模块化解析用户代理字符串
- 分离基础服务和增强服务
- 共享数据库连接以支持多仓库操作
- 改进的错误处理和资源管理

### 文件变更
- 新增 `internal/model/analytics.go` - 分析数据模型
- 新增 `internal/repository/analytics_repo.go` - 分析数据仓库
- 新增 `internal/service/analytics_service.go` - 分析服务
- 新增 `internal/service/enhanced_shortener.go` - 增强短链接服务
- 新增 `internal/handler/enhanced_handler.go` - 增强处理器
- 新增 `internal/utils/user_agent_parser.go` - 用户代理解析工具
- 更新 `cmd/server/main.go` - 集成新功能
- 新增 `ADVANCED_ANALYTICS.md` - 详细功能说明
- 更新 `README.md` - 反映新功能