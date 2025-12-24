package router

import (
	"bytes"
	"io"
	"time"

	"phone-server/handlers"
	"phone-server/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "phone-server/docs" // 导入Swagger生成的docs包
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 请求ID（如果没有则生成一个简单的）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = time.Now().Format("20060102150405") + "-" + c.ClientIP()
		}
		c.Set("request_id", requestID)

		// 请求方法、路径、查询参数
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 记录请求体（如果是POST/PUT等方法）
		var requestBody string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// 保存原始请求体
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// 重置请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 记录请求信息
		utils.Infofc(c.Request.Context(), "[REQUEST] request_id=%s, method=%s, path=%s, query=%s, client_ip=%s, body=%s",
			requestID, method, path, query, c.ClientIP(), requestBody)

		// 响应写入器包装
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// 记录响应信息
		statusCode := c.Writer.Status()
		responseBody := blw.body.String()
		bodySize := c.Writer.Size()

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			utils.Errorf("[RESPONSE] request_id=%s, method=%s, path=%s, status=%d, duration=%v, body_size=%d, client_ip=%s, response=%s",
				requestID, method, path, statusCode, duration, bodySize, c.ClientIP(), responseBody)
		} else if statusCode >= 400 {
			utils.Warnf("[RESPONSE] request_id=%s, method=%s, path=%s, status=%d, duration=%v, body_size=%d, client_ip=%s, response=%s",
				requestID, method, path, statusCode, duration, bodySize, c.ClientIP(), responseBody)
		} else {
			utils.Infof("[RESPONSE] request_id=%s, method=%s, path=%s, status=%d, duration=%v, body_size=%d, client_ip=%s",
				requestID, method, path, statusCode, duration, bodySize, c.ClientIP())
		}
	}
}

// bodyLogWriter 用于记录响应体的包装器
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// SetupRouter 初始化并配置Gin路由
func SetupRouter(httpHandler *handlers.HTTPHandler, wsHandler *handlers.WebSocketHandler, authHandler *handlers.AuthHandler, jwtSecret string) *gin.Engine {
	// 创建Gin引擎
	// 生产环境中使用gin.ReleaseMode
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 添加默认中间件
	router.Use(gin.Recovery())
	// 添加请求日志中间件
	router.Use(RequestLogger())

	// 设置可信代理，解决GIN警告
	// 生产环境中应该设置具体的可信代理IP
	router.SetTrustedProxies(nil)

	// 配置CORS中间件
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// 设置路由
	// API路由组
	apiGroup := router.Group("/api")
	{
		// 认证路由组
		authGroup := apiGroup.Group("/auth")
		{
			// 添加OPTIONS路由处理
			authGroup.OPTIONS("/*path", func(c *gin.Context) {
				c.Header("Access-Control-Allow-Origin", "*")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
				c.Header("Access-Control-Expose-Headers", "Content-Length")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Max-Age", "43200")
				c.AbortWithStatus(200)
			})
			// 用户注册
			authGroup.POST("/register", authHandler.Register)
			// 用户登录
			authGroup.POST("/login", authHandler.Login)
			// 刷新Token
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		// 消息路由组（需要认证中间件）
		messageGroup := apiGroup.Group("/")
		messageGroup.Use(handlers.AuthMiddleware(jwtSecret))
		{
			// 发送文本消息
			messageGroup.POST("/message", httpHandler.SendTextMessage)
			// 发送图片消息
			messageGroup.POST("/image", httpHandler.SendImageMessage)
			// AI聊天
			messageGroup.POST("/ai/chat", httpHandler.ChatWithAI)
		}
	}

	// WebSocket路由
	router.GET("/ws", wsHandler.HandleWebSocket)

	// 根路径处理：根据请求头判断是HTTP请求还是WebSocket连接请求
	router.GET("/", func(c *gin.Context) {
		// 检查请求头是否包含Upgrade: websocket
		if c.GetHeader("Upgrade") == "websocket" {
			// 升级为WebSocket连接
			wsHandler.HandleWebSocket(c)
			return
		}
		// 返回HTTP响应
		c.JSON(200, gin.H{
			"message": "WebSocket服务器已启动",
			"api": gin.H{
				"register":  "/api/auth/register (POST)",
				"login":     "/api/auth/login (POST)",
				"sendText":  "/api/message (POST)",
				"sendImage": "/api/image (POST)",
				"websocket": "/ws (GET) 或 / (GET with Upgrade: websocket)",
				"swagger":   "/swagger/index.html",
			},
		})
	})

	// 添加Swagger路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
