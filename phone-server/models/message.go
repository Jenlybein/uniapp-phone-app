package models

import (
	"time"

	"gorm.io/gorm"
)

// MessageType 消息类型
type MessageType string

const (
	// MessageTypeText 文本消息
	MessageTypeText MessageType = "text"
	// MessageTypeImage 图片消息
	MessageTypeImage MessageType = "image"
)

// SenderType 发送者类型
type SenderType string

const (
	// SenderTypePC PC端发送
	SenderTypePC SenderType = "pc"
	// SenderTypeServer 服务器发送
	SenderTypeServer SenderType = "server"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Device 设备模型
type Device struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	DeviceType   string         `gorm:"size:10;not null" json:"device_type"` // pc 或 phone
	DeviceToken  string         `gorm:"size:255;not null" json:"device_token"`
	Status       string         `gorm:"size:10;not null;default:'offline'" json:"status"` // online 或 offline
	LastActiveAt time.Time      `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	User         User           `gorm:"foreignKey:UserID" json:"-"`
}

// Message 消息模型
type Message struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"`
	Type       MessageType    `gorm:"size:10;not null" json:"type"` // text 或 image
	Content    string         `gorm:"type:text;not null" json:"content"`
	Sender     SenderType     `gorm:"size:10;not null" json:"sender"` // pc 或 server
	IsSelected bool           `gorm:"not null;default:false" json:"is_selected"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	User       User           `gorm:"foreignKey:UserID" json:"-"`
}

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

// NewTextMessage 创建文本消息
func NewTextMessage(userID uint, content string, sender SenderType) *Message {
	return &Message{
		UserID:  userID,
		Type:    MessageTypeText,
		Content: content,
		Sender:  sender,
	}
}

// NewImageMessage 创建图片消息
func NewImageMessage(userID uint, content string, sender SenderType) *Message {
	return &Message{
		UserID:  userID,
		Type:    MessageTypeImage,
		Content: content,
		Sender:  sender,
	}
}

