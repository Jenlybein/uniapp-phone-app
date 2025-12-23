package main

import (
	"fmt"
	"phone-server/configs"
)

func main() {
	// 加载配置
	config := configs.LoadConfig()

	// 打印配置信息
	fmt.Println("=== 配置加载测试 ===")
	fmt.Printf("端口: %d\n", config.Port)
	fmt.Printf("AI API Key: %s\n", config.AIConfig.ApiKey)
	fmt.Printf("AI Base URL: %s\n", config.AIConfig.BaseURL)
	fmt.Printf("AI Model: %s\n", config.AIConfig.Model)
	fmt.Printf("AI Thinking: %s\n", config.AIConfig.Thinking)

	fmt.Println("\n=== AI服务配置测试 ===")
	fmt.Printf("AI思考模式: %s\n", config.AIConfig.Thinking)

	if config.AIConfig.Thinking != "" {
		fmt.Println("✅ thinking参数已正确加载和设置")
	} else {
		fmt.Println("❌ thinking参数未能正确加载")
	}
}
