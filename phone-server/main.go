package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"phone-server/configs"
	"phone-server/handlers"
	"phone-server/router"
	"phone-server/services"
	"phone-server/utils"
)

func main() {
	// 初始化日志系统
	utils.InitLogger()
	defer utils.CloseLogger()

	// 加载配置
	cfg := configs.LoadConfig()

	// 初始化数据库连接
	db, err := configs.InitDatabase(cfg)
	if err != nil {
		utils.Fatalf("数据库初始化失败: %v", err)
	}

	// 创建消息广播服务并启动
	broker := services.NewBroker()
	go broker.Start()

	// 创建AI服务实例
	aiService := services.NewAIService(cfg.AIConfig.ApiKey, cfg.AIConfig.BaseURL, cfg.AIConfig.Model, cfg.AIConfig.Thinking)

	// 创建认证处理器
	authHandler := handlers.NewAuthHandler(db, cfg.JWTConfig.SecretKey, cfg.JWTConfig.ExpireHour)

	// 创建HTTP处理器
	httpHandler := handlers.NewHTTPHandler(broker, db, aiService)

	// 创建WebSocket处理器
	wsHandler := handlers.NewWebSocketHandler(broker, db, aiService, cfg.JWTConfig.SecretKey)

	// 初始化路由
	router := router.SetupRouter(httpHandler, wsHandler, authHandler, cfg.JWTConfig.SecretKey)

	// 显示启动提示信息
	showStartupInfo(cfg.Port)

	// 启动服务器（非阻塞）
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := router.Run(addr); err != nil {
			utils.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号，优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Infof("正在关闭服务器...")
	utils.Infof("服务器已关闭")
}

// showStartupInfo 显示启动提示信息
func showStartupInfo(port int) {
	fmt.Println("\n=== WebSocket 服务器已启动 ===")
	fmt.Printf("HTTP服务地址: http://localhost:%d\n", port)
	fmt.Printf("WebSocket地址: ws://localhost:%d/ws\n", port)
	fmt.Println()
	fmt.Println("使用说明:")
	fmt.Println("1. HTTP POST接口用于接收文本或图片数据")
	fmt.Printf("   - 发送文本: POST http://localhost:%d/api/message\n", port)
	fmt.Println("     请求体: {\"content\": \"具体文本内容\"}")
	fmt.Printf("   - 发送图片: POST http://localhost:%d/api/image\n", port)
	fmt.Println("     表单字段: image (文件)")
	fmt.Printf("2. WebSocket客户端连接到 ws://localhost:%d/ws 接收实时消息\n", port)
	fmt.Println("3. 支持多客户端同时连接")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务器")
	fmt.Println()
}
