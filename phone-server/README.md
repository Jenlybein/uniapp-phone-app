# Phone Answer Server

Phone Answer Server 是一个基于 Go 语言开发的后端服务，提供了 WebSocket 通信、AI 聊天、数据库操作等功能，专为电话助手应用设计。

## 功能特性

- **WebSocket 实时通信**：支持多客户端连接，实现实时消息广播
- **AI 聊天功能**：
  - 文本聊天：通过 AI 模型进行文本对话
  - 图片聊天：支持图片识别和描述
  - 流式响应：AI 回复实时流式显示
  - 思考模式控制：可配置 AI 思考模式
- **数据库操作**：自动迁移表结构，支持用户、设备、消息等数据管理
- **JWT 认证**：安全的用户认证机制
- **完善的日志系统**：
  - 分级日志：DEBUG/INFO/WARN/ERROR/FATAL
  - 分开存储：服务器日志和数据库日志分离
  - 控制台输出控制：可独立配置服务器和数据库日志是否输出到控制台
  - 文件滚动：按日期生成日志文件
- **Swagger 文档**：自动生成的 API 文档，便于测试和使用

## 技术栈

- **语言框架**：Go 1.20+
- **Web 框架**：Gin Web Framework
- **数据库**：MySQL
- **ORM**：GORM
- **认证**：JWT
- **配置管理**：YAML
- **API 文档**：Swagger

## 项目结构

```
phone-server/
├── configs/          # 配置相关代码
├── docs/             # Swagger 文档
├── handlers/         # 请求处理器
├── models/           # 数据模型
├── router/           # 路由配置
├── services/         # 业务逻辑
├── utils/            # 工具函数
├── main.go           # 项目入口
├── settings.yaml     # 配置文件
└── README.md         # 项目说明
```

## 安装部署

### 环境要求

- Go 1.20+
- MySQL 5.7+

### 安装步骤

1. 克隆项目：
   ```bash
git clone <repository-url>
cd phone-server
```

2. 安装依赖：
   ```bash
go mod tidy
```

3. 配置数据库：
   修改 `settings.yaml` 中的数据库配置

4. 运行项目：
   ```bash
go run main.go
```

   或构建可执行文件：
   ```bash
go build
echo "构建完成，生成可执行文件：phone-server.exe"
```

## 配置说明

配置文件：`settings.yaml`

主要配置项：

```yaml
# 服务器端口
port: 8080

# 数据库配置
db_host: "127.0.0.1"
db_port: 3308
db_username: "root"
db_password: "root"
db_name: "phone_db"

# AI 服务配置
ai_api_key: "your-api-key"
ai_base_url: "https://api.example.com"
ai_model: "model-name"
thinking: "disabled"  # enabled/disabled

# 日志配置
log_level: "INFO"       # DEBUG/INFO/WARN/ERROR/FATAL
log_file_path: "logs/"
log_max_size: 100      # MB
log_max_days: 7
log_compress: false
log_console_server: true   # 服务器日志是否输出到控制台
log_console_database: false  # 数据库日志是否输出到控制台
```

## API 文档

### Swagger 文档

启动服务后，访问：
```
http://localhost:8080/swagger/index.html
```

### 主要 API

#### 认证相关

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新令牌

#### 消息相关

- `POST /api/message` - 发送文本消息
- `POST /api/image` - 发送图片消息

#### AI 聊天

- `POST /api/ai/chat` - AI 聊天

#### WebSocket

- `GET /ws` - WebSocket 连接

## 使用示例

### WebSocket 连接

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('WebSocket 连接成功');
};

ws.onmessage = (event) => {
  console.log('收到消息:', event.data);
};

ws.send(JSON.stringify({
  type: 'text',
  content: '你好'
}));
```

### AI 聊天 API

```bash
curl -X POST http://localhost:8080/api/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"type":"text","content":"你好"}'
```

## 开发指南

### 目录结构

- `handlers/`：HTTP 请求处理器
- `models/`：数据模型定义
- `services/`：业务逻辑
- `router/`：路由配置
- `utils/`：工具函数

### 日志系统

日志文件位于 `logs/` 目录：
- `YYYY-MM-DD-server.log`：服务器日志
- `YYYY-MM-DD-database.log`：数据库日志

### 开发流程

1. 创建新分支
2. 编写代码
3. 运行测试：`go test ./...`
4. 提交代码

## 许可证

MIT License

## 联系方式

如有问题或建议，欢迎提出 Issue 或 Pull Request。