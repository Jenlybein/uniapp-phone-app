package models

import (
	"time"

	"gorm.io/gorm"
)

// AIResult AI结果模型
type AIResult struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	MessageID uint           `gorm:"uniqueIndex;not null" json:"message_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
	Message   Message        `gorm:"foreignKey:MessageID" json:"-"`
}
