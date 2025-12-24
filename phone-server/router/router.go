package router

import (
	"phone-server/handlers"
	"phone-server/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "phone-server/docs" // 导入Swagger生成的docs包
)

// SetupRouter 初始化并配置Gin路由
func SetupRouter(httpHandler *handlers.HTTPHandler, wsHandler *handlers.WebSocketHandler, authHandler *handlers.AuthHandler, jwtSecret string) *gin.Engine {
	// 创建Gin引擎
	// 生产环境中使用gin.ReleaseMode
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 添加默认中间件
	router.Use(gin.Recovery())
	// 添加请求日志中间件
	router.Use(middleware.RequestLogger())

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
		messageGroup.Use(middleware.AuthMiddleware(jwtSecret))
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
