package configs

import (
	"flag"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

// AIConfig AI服务配置结构体
type AIConfig struct {
	ApiKey   string `yaml:"ai_api_key"`  // AI服务API密钥
	BaseURL  string `yaml:"ai_base_url"` // AI服务基础URL
	Model    string `yaml:"ai_model"`    // AI模型名称
	Thinking string `yaml:"thinking"`    // AI思考模式
}

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Host     string // 数据库主机
	Port     int    // 数据库端口
	Username string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名称
}

// JWTConfig JWT配置结构体
type JWTConfig struct {
	SecretKey  string // JWT密钥
	ExpireHour int    // JWT过期时间（小时）
}

// Config 服务器配置结构体
type Config struct {
	Port           int            `yaml:"port"` // 服务器端口
	AIConfig       AIConfig       // AI服务配置
	DatabaseConfig DatabaseConfig // 数据库配置
	JWTConfig      JWTConfig      // JWT配置
}

// LoadConfig 加载配置
// 优先级：命令行参数 > 环境变量 > yaml配置文件 > 默认值
func LoadConfig() *Config {
	// 默认配置
	config := &Config{
		Port: 8080, // 默认端口8080
		DatabaseConfig: DatabaseConfig{
			Host:     "127.0.0.1", // 默认数据库主机
			Port:     3308,        // 默认数据库端口
			Username: "root",      // 默认数据库用户名
			Password: "root",      // 默认数据库密码
			DBName:   "phone_db",  // 默认数据库名称
		},
		JWTConfig: JWTConfig{
			SecretKey:  "your-secret-key", // 默认JWT密钥
			ExpireHour: 24,                // 默认JWT过期时间（小时）
		},
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
	// 加载数据库配置
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.DatabaseConfig.Host = dbHost
	}
	if dbPortStr := os.Getenv("DB_PORT"); dbPortStr != "" {
		if dbPort, err := strconv.Atoi(dbPortStr); err == nil {
			config.DatabaseConfig.Port = dbPort
		}
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.DatabaseConfig.Username = dbUser
	}
	if dbPass := os.Getenv("DB_PASS"); dbPass != "" {
		config.DatabaseConfig.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.DatabaseConfig.DBName = dbName
	}
	// 加载JWT配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTConfig.SecretKey = jwtSecret
	}
	if jwtExpireStr := os.Getenv("JWT_EXPIRE"); jwtExpireStr != "" {
		if jwtExpire, err := strconv.Atoi(jwtExpireStr); err == nil {
			config.JWTConfig.ExpireHour = jwtExpire
		}
	}

	// 从命令行参数加载（优先级最高）
	portFlag := flag.Int("port", 0, "服务器端口")
	apiKeyFlag := flag.String("ai-api-key", "", "AI服务API密钥")
	baseURLFlag := flag.String("ai-base-url", "", "AI服务基础URL")
	modelFlag := flag.String("ai-model", "", "AI模型名称")
	dbHostFlag := flag.String("db-host", "", "数据库主机")
	dbPortFlag := flag.Int("db-port", 0, "数据库端口")
	dbUserFlag := flag.String("db-user", "", "数据库用户名")
	dbPassFlag := flag.String("db-pass", "", "数据库密码")
	dbNameFlag := flag.String("db-name", "", "数据库名称")
	jwtSecretFlag := flag.String("jwt-secret", "", "JWT密钥")
	jwtExpireFlag := flag.Int("jwt-expire", 0, "JWT过期时间（小时）")
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
	if *dbHostFlag != "" {
		config.DatabaseConfig.Host = *dbHostFlag
	}
	if *dbPortFlag != 0 {
		config.DatabaseConfig.Port = *dbPortFlag
	}
	if *dbUserFlag != "" {
		config.DatabaseConfig.Username = *dbUserFlag
	}
	if *dbPassFlag != "" {
		config.DatabaseConfig.Password = *dbPassFlag
	}
	if *dbNameFlag != "" {
		config.DatabaseConfig.DBName = *dbNameFlag
	}
	if *jwtSecretFlag != "" {
		config.JWTConfig.SecretKey = *jwtSecretFlag
	}
	if *jwtExpireFlag != 0 {
		config.JWTConfig.ExpireHour = *jwtExpireFlag
	}

	return config
}

// 定义一个与YAML文件结构匹配的扁平结构体
type FlatConfig struct {
	Port      int    `yaml:"port"`
	AiApiKey  string `yaml:"ai_api_key"`
	AiBaseUrl string `yaml:"ai_base_url"`
	AiModel   string `yaml:"ai_model"`
	Thinking  string `yaml:"thinking"`
}

// loadFromYaml 从yaml配置文件加载配置
func (c *Config) loadFromYaml() {
	// 读取yaml配置文件
	yamlFile, err := os.ReadFile("settings.yaml")
	if err != nil {
		// 如果文件不存在，使用默认配置
		return
	}

	// 直接解析为扁平结构体
	var flatConfig FlatConfig
	if err := yaml.Unmarshal(yamlFile, &flatConfig); err != nil {
		// 如果解析失败，使用手动解析方式
		var rawConfig map[string]interface{}
		if err := yaml.Unmarshal(yamlFile, &rawConfig); err != nil {
			return
		}

		// 手动解析配置项
		if port, ok := rawConfig["port"].(int); ok {
			c.Port = port
		}
		if apiKey, ok := rawConfig["ai_api_key"].(string); ok {
			c.AIConfig.ApiKey = apiKey
		}
		if baseURL, ok := rawConfig["ai_base_url"].(string); ok {
			c.AIConfig.BaseURL = baseURL
		}
		if model, ok := rawConfig["ai_model"].(string); ok {
			c.AIConfig.Model = model
		}
		if thinking, ok := rawConfig["thinking"].(string); ok {
			c.AIConfig.Thinking = thinking
		}
		return
	}

	// 使用扁平结构体的解析结果
	if flatConfig.Port != 0 {
		c.Port = flatConfig.Port
	}
	if flatConfig.AiApiKey != "" {
		c.AIConfig.ApiKey = flatConfig.AiApiKey
	}
	if flatConfig.AiBaseUrl != "" {
		c.AIConfig.BaseURL = flatConfig.AiBaseUrl
	}
	if flatConfig.AiModel != "" {
		c.AIConfig.Model = flatConfig.AiModel
	}
	if flatConfig.Thinking != "" {
		c.AIConfig.Thinking = flatConfig.Thinking
	}
}
