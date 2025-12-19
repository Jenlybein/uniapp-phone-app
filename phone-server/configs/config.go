package configs

import (
	"flag"
	"os"
	"phone-server/utils"
	"strconv"

	"github.com/goccy/go-yaml"
)

type AIConfig struct {
	ApiKey  string // AI服务API密钥
	BaseURL string // AI服务基础URL
	Model   string // AI模型名称
}

// Config 服务器配置结构体
type Config struct {
	Port     int      // 服务器端口
	AIConfig AIConfig // AI服务配置
}

// LoadConfig 加载配置
// 优先级：命令行参数 > 环境变量 > yaml配置文件 > 默认值
func LoadConfig() *Config {
	// 默认配置
	config := &Config{
		Port: 8080, // 默认端口8080
	}

	// 从yaml配置文件加载
	config.loadFromYaml()

	// 从环境变量加载（优先级高于yaml）
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}
	// 加载AI API密钥
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		config.AIConfig.ApiKey = apiKey
	}
	// 加载AI基础URL
	if baseURL := os.Getenv("AI_BASE_URL"); baseURL != "" {
		config.AIConfig.BaseURL = baseURL
	}
	// 加载AI模型名称
	if model := os.Getenv("AI_MODEL"); model != "" {
		config.AIConfig.Model = model
	}

	// 从命令行参数加载（优先级最高）
	portFlag := flag.Int("port", 0, "服务器端口")
	apiKeyFlag := flag.String("ai-api-key", "", "AI服务API密钥")
	baseURLFlag := flag.String("ai-base-url", "", "AI服务基础URL")
	modelFlag := flag.String("ai-model", "", "AI模型名称")
	flag.Parse()

	if *portFlag != 0 {
		config.Port = *portFlag
	}
	if *apiKeyFlag != "" {
		config.AIConfig.ApiKey = *apiKeyFlag
	}
	if *baseURLFlag != "" {
		config.AIConfig.BaseURL = *baseURLFlag
	}
	if *modelFlag != "" {
		config.AIConfig.Model = *modelFlag
	}

	return config
}

// loadFromYaml 从yaml配置文件加载配置
func (c *Config) loadFromYaml() {
	// 读取yaml配置文件
	yamlFile, err := os.ReadFile("settings.yaml")
	if err != nil {
		utils.Errorf("读取配置文件失败: %v", err)
		return
	}

	// 解析yaml文件
	var yamlConfig struct {
		Port      int    `yaml:"port"`
		AIAPIKey  string `yaml:"ai_api_key"`
		AIBaseURL string `yaml:"ai_base_url"`
		AIModel   string `yaml:"ai_model"`
	}

	if err := yaml.Unmarshal(yamlFile, &yamlConfig); err != nil {
		utils.Errorf("解析配置文件失败: %v", err)
		return
	}

	utils.Infof("从配置文件读取到的AI配置: APIKey=%s, BaseURL=%s, Model=%s",
		yamlConfig.AIAPIKey, yamlConfig.AIBaseURL, yamlConfig.AIModel)

	// 只使用yaml中明确指定的配置项（非零值）
	if yamlConfig.Port != 0 {
		c.Port = yamlConfig.Port
	}
	if yamlConfig.AIAPIKey != "" {
		c.AIConfig.ApiKey = yamlConfig.AIAPIKey
		utils.Infof("设置AI API Key: %s", yamlConfig.AIAPIKey)
	}
	if yamlConfig.AIBaseURL != "" {
		c.AIConfig.BaseURL = yamlConfig.AIBaseURL
		utils.Infof("设置AI BaseURL: %s", yamlConfig.AIBaseURL)
	}
	if yamlConfig.AIModel != "" {
		c.AIConfig.Model = yamlConfig.AIModel
		utils.Infof("设置AI Model: %s", yamlConfig.AIModel)
	}
}
