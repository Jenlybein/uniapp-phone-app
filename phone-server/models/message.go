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

