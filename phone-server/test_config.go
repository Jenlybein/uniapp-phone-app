package main

import (
	"fmt"
	"os"
	"path/filepath"
	"phone-server/configs"
)

func main() {
	// 打印当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前工作目录失败: %v\n", err)
		return
	}
	fmt.Printf("当前工作目录: %s\n", cwd)

	// 检查配置文件是否存在
	yamlPath := filepath.Join(cwd, "settings.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		fmt.Printf("配置文件不存在: %s\n", yamlPath)
		// 尝试使用绝对路径
		yamlPath = "e:/project/phone_answer/phone-server/settings.yaml"
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			fmt.Printf("绝对路径配置文件也不存在: %s\n", yamlPath)
			return
		}
		fmt.Printf("使用绝对路径配置文件: %s\n", yamlPath)
	} else {
		fmt.Printf("使用相对路径配置文件: %s\n", yamlPath)
	}

	// 加载配置
	config := configs.LoadConfig()

	// 打印配置信息
	fmt.Println("=== 配置加载测试 ===")
	fmt.Printf("端口: %d\n", config.Port)
	fmt.Printf("AI API Key: %s\n", config.AIConfig.ApiKey)
	fmt.Printf("AI Base URL: %s\n", config.AIConfig.BaseURL)
	fmt.Printf("AI Model: %s\n", config.AIConfig.Model)

	// 验证AI配置是否正确加载
	if config.AIConfig.BaseURL == "" {
		fmt.Println("❌ AI Base URL 加载失败")
	} else {
		fmt.Println("✅ AI Base URL 加载成功")
	}
}
