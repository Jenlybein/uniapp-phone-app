package router

import (
	"phone-server/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化并配置Gin路由
func SetupRouter(httpHandler *handlers.HTTPHandler, wsHandler *handlers.WebSocketHandler) *gin.Engine {
	// 创建Gin引擎
	// 生产环境中使用gin.ReleaseMode
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 设置可信代理，解决GIN警告
	// 生产环境中应该设置具体的可信代理IP
	router.SetTrustedProxies(nil)

	// 设置路由
	// API路由组
	apiGroup := router.Group("/api")
	{
		// 健康检查
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "服务运行正常",
			})
		})
		// 获取服务状态
		apiGroup.GET("/status", httpHandler.GetStatus)
		// 发送文本消息
		apiGroup.POST("/message", httpHandler.SendTextMessage)
		// 发送图片消息
		apiGroup.POST("/image", httpHandler.SendImageMessage)
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
				"sendText":  "/api/message (POST)",
				"sendImage": "/api/image (POST)",
				"websocket": "/ws (GET) 或 / (GET with Upgrade: websocket)",
			},
		})
	})

	return router
}
