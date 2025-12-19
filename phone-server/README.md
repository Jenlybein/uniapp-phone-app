# WebSocket AI服务

一个基于Go语言和Gin框架实现的WebSocket服务器，支持接收文本和图片消息，并通过AI服务生成响应，以流式方式返回给客户端。

## 功能特性

- ✅ WebSocket服务，支持多客户端连接
- ✅ HTTP API接口，支持接收文本和图片消息
- ✅ AI集成，支持文本和图片的智能响应
- ✅ 流式响应，实时返回AI生成结果
- ✅ 健康检查和状态监控
- ✅ 可配置，支持环境变量和命令行参数
- ✅ 完善的错误处理和日志记录
- ✅ 前端测试页面，方便调试

## 技术栈

- Go 1.25.1
- Gin Web框架
- Gorilla WebSocket库
- OpenAI Go客户端库
- YAML配置管理

## 目录结构

```
.
├── configs/          # 配置管理
│   └── config.go     # 配置加载逻辑
├── handlers/         # 请求处理器
│   ├── http.go       # HTTP接口处理
│   └── websocket.go  # WebSocket处理
├── models/           # 数据模型
│   └── message.go    # 消息结构定义
├── services/         # 业务服务
│   ├── ai.go         # AI服务集成
│   ├── broker.go     # 消息广播服务
│   └── ...
├── utils/            # 工具函数
│   └── logger.go     # 日志工具
├── router/           # 路由配置
│   └── router.go     # 路由定义
├── test-websocket.html  # 前端测试页面
├── settings.yaml     # 配置文件
├── go.mod            # 依赖管理
└── main.go           # 程序入口
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 编译程序

```bash
go build -o phone-server.exe
```

### 3. 配置

#### 配置文件 (settings.yaml)

```yaml
# 服务器端口
port: 8080

# AI服务配置
# API密钥，用于访问AI服务
ai_api_key: "your-api-key-here"

# AI服务基础URL，用于指定非OpenAI的AI服务（如豆包）
ai_base_url: ""
```

#### 环境变量

```bash
# 服务器端口
set PORT=8080

# AI API密钥
set AI_API_KEY=your-api-key-here

# AI基础URL
set AI_BASE_URL=""
```

#### 命令行参数

```bash
./phone-server.exe --port 8080 --ai-api-key your-api-key-here --ai-base-url ""
```

### 4. 启动服务

```bash
./phone-server.exe
```

服务启动后，将监听在 `http://localhost:8080` 和 `ws://localhost:8080/ws`。

## API 接口

### 健康检查

```
GET /api/health
```

**响应示例：**
```json
{
  "status": "ok",
  "message": "服务运行正常"
}
```

### 服务状态

```
GET /api/status
```

**响应示例：**
```json
{
  "status": "ok",
  "connectionCount": 2,
  "message": "服务运行正常"
}
```

### 发送文本消息

```
POST /api/message
Content-Type: application/json
```

**请求示例：**
```json
{
  "content": "你好，AI！"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "消息发送成功"
}
```

### 发送图片消息

```
POST /api/image
Content-Type: multipart/form-data
```

**请求示例：**
```
curl -X POST -F "image=@image.jpg" http://localhost:8080/api/image
```

**响应示例：**
```json
{
  "code": 200,
  "message": "图片发送成功"
}
```

## WebSocket 接口

### 连接地址

```
ws://localhost:8080/ws
```

或使用根路径（自动升级为WebSocket）：

```
ws://localhost:8080
```

### 消息格式

#### 客户端发送消息

##### 文本消息

```json
{
  "type": "text",
  "content": "你的问题"
}
```

##### 图片消息

```json
{
  "type": "image",
  "content": "base64图片数据"
}
```

#### 服务器响应消息

```json
{
  "type": "text",
  "content": "AI生成的响应内容"
}
```

## 前端测试页面

服务器提供了一个前端测试页面，方便调试和验证WebSocket功能：

```
http://localhost:8080/test-websocket.html
```

### 功能说明

- 自动连接到WebSocket服务
- 显示连接状态
- 支持发送文本消息
- 支持上传和发送图片
- 实时显示消息记录
- 自动重连机制

## 验证接口可用性

### 1. 健康检查

使用curl或浏览器访问：

```bash
curl http://localhost:8080/api/health
```

预期响应：
```json
{"status":"ok","message":"服务运行正常"}
```

### 2. 服务状态

```bash
curl http://localhost:8080/api/status
```

预期响应：
```json
{"status":"ok","connectionCount":0,"message":"服务运行正常"}
```

### 3. WebSocket连接测试

使用浏览器控制台或WebSocket客户端工具：

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function() {
  console.log('连接成功');
  ws.send(JSON.stringify({type: 'text', content: '你好'}));
};

ws.onmessage = function(event) {
  console.log('收到消息:', event.data);
};

ws.onerror = function(error) {
  console.error('错误:', error);
};

ws.onclose = function() {
  console.log('连接关闭');
};
```

### 4. 使用前端测试页面

1. 在浏览器中打开 `http://localhost:8080/test-websocket.html`
2. 观察连接状态变为"已连接"
3. 输入文本消息并发送，查看AI响应
4. 上传图片并发送，查看AI响应

## 日志

日志文件位于 `logs/` 目录下，按日期命名（如 `2025-12-19.log`）。

## 错误处理

- 所有错误都会记录到日志文件
- WebSocket连接失败会自动重连
- AI服务异常会返回友好的错误信息
- 请求参数错误会返回400状态码

## 性能优化

- 使用协程处理每个WebSocket连接
- 采用通道和互斥锁保证线程安全
- 添加上下文超时，防止长时间阻塞
- 高效的消息序列化和反序列化

## 生产环境建议

1. 设置 `gin.SetMode(gin.ReleaseMode)` 开启生产模式
2. 配置具体的可信代理IP，而不是使用 `router.SetTrustedProxies(nil)`
3. 使用环境变量或配置文件管理敏感信息
4. 配置适当的日志级别
5. 考虑添加身份认证和授权机制
6. 部署多个实例，使用负载均衡

## 许可证

MIT
