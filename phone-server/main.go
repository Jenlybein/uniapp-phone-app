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

// @title 电话助手后端API
// @version 1.0
// @description 电话助手后端服务，提供WebSocket实时通信和AI对话功能
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// 加载配置
	cfg := configs.LoadConfig()

	// 使用配置初始化服务器日志系统
	utils.InitLoggerWithConfig(
		cfg.LogConfig.Level,
		cfg.LogConfig.FilePath,
		cfg.LogConfig.MaxSize,
		cfg.LogConfig.MaxDays,
		cfg.LogConfig.Compress,
		cfg.LogConfig.ConsoleServer,
	)
	defer utils.CloseLogger()

	// 初始化数据库日志系统（使用相同的配置，但日志类型为database）
	if err := utils.InitDatabaseLogger(utils.LoggerConfig{
		Level:    utils.GetLevelFromString(cfg.LogConfig.Level),
		FilePath: cfg.LogConfig.FilePath,
		MaxSize:  cfg.LogConfig.MaxSize,
		MaxDays:  cfg.LogConfig.MaxDays,
		Compress: cfg.LogConfig.Compress,
		Console:  cfg.LogConfig.ConsoleDatabase,
		Type:     utils.LoggerTypeDatabase,
	}); err != nil {
		utils.Errorf("初始化数据库日志系统失败: %v", err)
	}

	// 记录配置加载结果
	utils.Infof("配置加载成功，端口: %d, AI模型: %s, 日志级别: %s",
		cfg.Port, cfg.AIConfig.Model, cfg.LogConfig.Level)

	// 初始化数据库连接
	utils.Infof("正在初始化数据库连接...")
	db, err := configs.InitDatabase(cfg)
	if err != nil {
		utils.Fatalf("数据库初始化失败: %v", err)
	}
	utils.Infof("数据库连接成功")

	// 创建消息广播服务并启动
	broker := services.NewBroker()
	go broker.Start()
	utils.Infof("消息广播服务已启动")

	// 创建AI服务实例
	aiService := services.NewAIService(cfg.AIConfig.ApiKey, cfg.AIConfig.BaseURL, cfg.AIConfig.Model, cfg.AIConfig.Thinking)
	utils.Infof("AI服务实例创建成功，模型: %s, 思考模式: %s",
		cfg.AIConfig.Model, cfg.AIConfig.Thinking)

	// 创建认证处理器
	authHandler := handlers.NewAuthHandler(db, cfg.JWTConfig.SecretKey, cfg.JWTConfig.ExpireHour)
	utils.Infof("认证处理器创建成功")

	// 创建HTTP处理器
	httpHandler := handlers.NewHTTPHandler(broker, db, aiService)
	utils.Infof("HTTP处理器创建成功")

	// 创建WebSocket处理器
	wsHandler := handlers.NewWebSocketHandler(broker, db, aiService, cfg.JWTConfig.SecretKey)
	utils.Infof("WebSocket处理器创建成功")

	// 初始化路由
	router := router.SetupRouter(httpHandler, wsHandler, authHandler, cfg.JWTConfig.SecretKey)
	utils.Infof("路由初始化成功")

	// 显示启动提示信息
	showStartupInfo(cfg.Port)

	// 启动服务器（非阻塞）
	utils.Infof("正在启动服务器，监听端口: %d...", cfg.Port)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := router.Run(addr); err != nil {
			utils.Fatalf("服务器启动失败: %v", err)
		}
	}()
	utils.Infof("服务器启动成功，HTTP地址: http://localhost:%d, WebSocket地址: ws://localhost:%d/ws",
		cfg.Port, cfg.Port)

	// 等待中断信号，优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Infof("正在关闭服务器...")
	utils.Infof("服务器已关闭")
}

// showStartupInfo 显示启动提示信息
func showStartupInfo(port int) {
	fmt.Printf("HTTP服务地址: http://localhost:%d\n", port)
	fmt.Printf("WebSocket地址: ws://localhost:%d/ws\n", port)
	fmt.Println("按 Ctrl+C 停止服务器")
}
