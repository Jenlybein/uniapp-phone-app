package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// 定义与settings.yaml匹配的结构体
// 注意：键名要与YAML文件中的键名完全一致（包括大小写）
type RawConfig struct {
	Port        int    `yaml:"port"`
	AiApiKey    string `yaml:"ai_api_key"`
	AiBaseUrl   string `yaml:"ai_base_url"`
	AiModel     string `yaml:"ai_model"`
}

func main() {
	// 读取YAML文件
	content, err := os.ReadFile("settings.yaml")
	if err != nil {
		fmt.Printf("读取YAML文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("YAML文件内容: %s\n", string(content))
	
	// 解析YAML
	var config RawConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		fmt.Printf("解析YAML失败: %v\n", err)
		return
	}
	
	// 打印解析结果
	fmt.Println("=== 解析结果 ===")
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("AiApiKey: %s\n", config.AiApiKey)
	fmt.Printf("AiBaseUrl: %s\n", config.AiBaseUrl)
	fmt.Printf("AiModel: %s\n", config.AiModel)
	
	// 验证AI配置是否正确加载
	if config.AiBaseUrl == "" {
		fmt.Println("❌ AiBaseUrl 加载失败")
	} else {
		fmt.Println("✅ AiBaseUrl 加载成功")
	}
}
