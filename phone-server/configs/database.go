package configs

import (
	"fmt"
	"log"
	"phone-server/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表结构
	if err := migrateDatabase(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	log.Println("数据库连接成功")
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
