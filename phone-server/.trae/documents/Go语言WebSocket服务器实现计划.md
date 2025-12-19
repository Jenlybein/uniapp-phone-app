# Go语言WebSocket服务器工程化实现计划

## 项目结构
```
phone-server/
├── main.go              # 主程序入口，初始化服务器
├── config/              # 配置相关
│   └── config.go        # 配置定义和加载
├── handlers/            # 请求处理器
│   ├── http.go          # HTTP接口处理
│   └── websocket.go     # WebSocket连接管理
├── models/              # 数据模型
│   └── message.go       # 消息结构体定义
├── services/            # 业务逻辑
│   └── broker.go        # 消息广播服务
├── go.mod               # 依赖管理
└── go.sum               # 依赖校验
```

## 核心功能实现

### 1. 配置管理 (config/config.go)
- 定义服务器配置结构体
- 支持通过环境变量或命令行参数配置
- 默认端口8080

### 2. 消息模型 (models/message.go)
- 定义统一的消息格式
- 支持文本和图片类型
- 提供序列化/反序列化方法

### 3. 消息广播服务 (services/broker.go)
- 维护线程安全的WebSocket客户端连接池
- 提供消息广播功能
- 处理客户端连接的注册和注销

### 4. HTTP接口 (handlers/http.go)
- 使用GIN框架提供HTTP服务
- `POST /api/message` - 接收文本消息
- `POST /api/image` - 接收图片文件并转换为base64
- 调用broker服务广播消息

### 5. WebSocket处理 (handlers/websocket.go)
- 使用gorilla/websocket库处理WebSocket连接
- 实现连接升级、消息接收和发送
- 管理客户端连接的生命周期

### 6. 主程序入口 (main.go)
- 初始化配置
- 设置路由
- 启动HTTP和WebSocket服务
- 显示启动提示信息

## 依赖库
- `github.com/gin-gonic/gin` - HTTP框架
- `github.com/gorilla/websocket` - WebSocket库

## 工程化特性
1. **模块化设计** - 代码按功能模块划分，职责清晰
2. **配置灵活** - 支持环境变量和命令行参数配置
3. **线程安全** - 客户端连接池使用互斥锁保护
4. **错误处理** - 完善的错误处理和日志记录
5. **易于维护** - 代码结构清晰，便于后续扩展
6. **启动提示** - 提供清晰的服务地址和使用说明

## 实现步骤
1. 初始化Go模块
2. 创建项目目录结构
3. 实现配置管理
4. 定义消息模型
5. 实现消息广播服务
6. 实现HTTP接口处理
7. 实现WebSocket处理
8. 编写主程序入口
9. 测试和调试

## 预期结果
- 服务器能够接收HTTP请求，包括文本消息和图片文件
- 能够维护多个WebSocket连接
- 将接收到的HTTP数据实时转发给所有已连接的WebSocket客户端
- 支持配置服务器端口
- 提供清晰的启动提示信息
- 实现基本的错误处理和日志输出
- 代码结构清晰，便于后续维护