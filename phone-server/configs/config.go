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

// LogConfig 日志配置结构体
type LogConfig struct {
	Level           string `yaml:"log_level"`            // 日志级别：DEBUG/INFO/WARN/ERROR/FATAL
	FilePath        string `yaml:"log_file_path"`        // 日志文件路径
	MaxSize         int    `yaml:"log_max_size"`         // 日志文件最大大小（MB）
	MaxDays         int    `yaml:"log_max_days"`         // 日志文件保留天数
	Compress        bool   `yaml:"log_compress"`         // 是否压缩日志文件
	ConsoleServer   bool   `yaml:"log_console_server"`   // 是否将基本日志信息输出到控制台
	ConsoleDatabase bool   `yaml:"log_console_database"` // 是否将数据库操作日志信息输出到控制台
}

// Config 服务器配置结构体
type Config struct {
	Port           int            `yaml:"port"` // 服务器端口
	AIConfig       AIConfig       // AI服务配置
	DatabaseConfig DatabaseConfig // 数据库配置
	JWTConfig      JWTConfig      // JWT配置
	LogConfig      LogConfig      // 日志配置
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
		LogConfig: LogConfig{
			Level:           "INFO",  // 默认日志级别
			FilePath:        "logs/", // 默认日志文件路径
			MaxSize:         100,     // 默认日志文件最大大小100MB
			MaxDays:         7,       // 默认日志文件保留7天
			Compress:        false,   // 默认不压缩日志文件
			ConsoleServer:   true,    // 默认将基本日志信息输出到控制台
			ConsoleDatabase: true,    // 默认将数据库操作日志信息输出到控制台
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
	Port int `yaml:"port"`
	// 数据库配置
	DbHost     string `yaml:"db_host"`
	DbPort     int    `yaml:"db_port"`
	DbUsername string `yaml:"db_username"`
	DbPassword string `yaml:"db_password"`
	DbName     string `yaml:"db_name"`
	// AI配置
	AiApiKey  string `yaml:"ai_api_key"`
	AiBaseUrl string `yaml:"ai_base_url"`
	AiModel   string `yaml:"ai_model"`
	Thinking  string `yaml:"thinking"`
	// 日志配置
	LogLevel           string `yaml:"log_level"`
	LogFilePath        string `yaml:"log_file_path"`
	LogMaxSize         int    `yaml:"log_max_size"`
	LogMaxDays         int    `yaml:"log_max_days"`
	LogCompress        bool   `yaml:"log_compress"`
	LogConsoleServer   bool   `yaml:"log_console_server"`
	LogConsoleDatabase bool   `yaml:"log_console_database"`
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
		// 数据库配置
		if dbHost, ok := rawConfig["db_host"].(string); ok {
			c.DatabaseConfig.Host = dbHost
		}
		if dbPort, ok := rawConfig["db_port"].(int); ok {
			c.DatabaseConfig.Port = dbPort
		}
		if dbUsername, ok := rawConfig["db_username"].(string); ok {
			c.DatabaseConfig.Username = dbUsername
		}
		if dbPassword, ok := rawConfig["db_password"].(string); ok {
			c.DatabaseConfig.Password = dbPassword
		}
		if dbName, ok := rawConfig["db_name"].(string); ok {
			c.DatabaseConfig.DBName = dbName
		}
		// AI配置
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
		// 日志配置
		if logLevel, ok := rawConfig["log_level"].(string); ok {
			c.LogConfig.Level = logLevel
		}
		if logFilePath, ok := rawConfig["log_file_path"].(string); ok {
			c.LogConfig.FilePath = logFilePath
		}
		if logMaxSize, ok := rawConfig["log_max_size"].(int); ok {
			c.LogConfig.MaxSize = logMaxSize
		}
		if logMaxDays, ok := rawConfig["log_max_days"].(int); ok {
			c.LogConfig.MaxDays = logMaxDays
		}
		if logCompress, ok := rawConfig["log_compress"].(bool); ok {
			c.LogConfig.Compress = logCompress
		}
		if logConsoleServer, ok := rawConfig["log_console_server"].(bool); ok {
			c.LogConfig.ConsoleServer = logConsoleServer
		}
		if logConsoleDatabase, ok := rawConfig["log_console_database"].(bool); ok {
			c.LogConfig.ConsoleDatabase = logConsoleDatabase
		}
		return
	}

	// 使用扁平结构体的解析结果
	if flatConfig.Port != 0 {
		c.Port = flatConfig.Port
	}
	// 数据库配置
	if flatConfig.DbHost != "" {
		c.DatabaseConfig.Host = flatConfig.DbHost
	}
	if flatConfig.DbPort != 0 {
		c.DatabaseConfig.Port = flatConfig.DbPort
	}
	if flatConfig.DbUsername != "" {
		c.DatabaseConfig.Username = flatConfig.DbUsername
	}
	if flatConfig.DbPassword != "" {
		c.DatabaseConfig.Password = flatConfig.DbPassword
	}
	if flatConfig.DbName != "" {
		c.DatabaseConfig.DBName = flatConfig.DbName
	}
	// AI配置
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
	// 日志配置
	if flatConfig.LogLevel != "" {
		c.LogConfig.Level = flatConfig.LogLevel
	}
	if flatConfig.LogFilePath != "" {
		c.LogConfig.FilePath = flatConfig.LogFilePath
	}
	if flatConfig.LogMaxSize != 0 {
		c.LogConfig.MaxSize = flatConfig.LogMaxSize
	}
	if flatConfig.LogMaxDays != 0 {
		c.LogConfig.MaxDays = flatConfig.LogMaxDays
	}
	// 布尔值不需要检查，直接赋值
	c.LogConfig.Compress = flatConfig.LogCompress
	c.LogConfig.ConsoleServer = flatConfig.LogConsoleServer
	c.LogConfig.ConsoleDatabase = flatConfig.LogConsoleDatabase
}
