package configs

import (
	"fmt"
	"phone-server/models"
	"phone-server/utils"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *Config) (*gorm.DB, error) {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DatabaseConfig.Username,
		cfg.DatabaseConfig.Password,
		cfg.DatabaseConfig.Host,
		cfg.DatabaseConfig.Port,
		cfg.DatabaseConfig.DBName,
	)

	utils.DatabaseInfof("正在连接数据库: %s:%d/%s, 用户名: %s",
		cfg.DatabaseConfig.Host, cfg.DatabaseConfig.Port, cfg.DatabaseConfig.DBName, cfg.DatabaseConfig.Username)

	// 配置GORM日志记录器
	gormLogger := logger.New(
		&gormLoggerWrapper{},
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
			Colorful:                  false,       // 禁用彩色输出
		},
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		utils.DatabaseErrorf("连接数据库失败: %v", err)
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表结构
	utils.DatabaseInfof("正在执行数据库表结构迁移...")
	if err := migrateDatabase(db); err != nil {
		utils.DatabaseErrorf("数据库迁移失败: %v", err)
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	utils.DatabaseInfof("数据库连接成功")
	return db, nil
}

// migrateDatabase 执行数据库表结构迁移
func migrateDatabase(db *gorm.DB) error {
	// 自动迁移所有模型
	return db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.Message{},
		&models.AIResult{},
	)
}

// gormLoggerWrapper GORM日志包装器，用于将GORM日志输出到自定义日志系统
type gormLoggerWrapper struct{}

// Printf 实现GORM的logger.Writer接口
func (l *gormLoggerWrapper) Printf(format string, args ...interface{}) {
	// 将GORM日志转换为我们的日志系统输出
	msg := fmt.Sprintf(format, args...)

	// 根据日志内容判断日志级别
	if strings.Contains(msg, "ERROR") || strings.Contains(msg, "Failed") {
		utils.DatabaseErrorf("[GORM] %s", msg)
	} else if strings.Contains(msg, "WARN") || strings.Contains(msg, "Warning") {
		utils.DatabaseWarnf("[GORM] %s", msg)
	} else {
		// 只记录SQL查询和慢查询
		if strings.Contains(msg, "SELECT") || strings.Contains(msg, "INSERT") ||
			strings.Contains(msg, "UPDATE") || strings.Contains(msg, "DELETE") ||
			strings.Contains(msg, "SLOW QUERY") {
			utils.DatabaseInfof("[GORM] %s", msg)
		}
	}
}
