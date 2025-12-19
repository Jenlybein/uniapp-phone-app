package models

// MessageType 消息类型
type MessageType string

const (
	// MessageTypeText 文本消息
	MessageTypeText MessageType = "text"
	// MessageTypeImage 图片消息
	MessageTypeImage MessageType = "image"
)

// Message 统一消息格式
type Message struct {
	Type    MessageType `json:"type"`    // 消息类型：text 或 image
	Content string      `json:"content"` // 消息内容：文本内容或图片base64编码
}

// NewTextMessage 创建文本消息
func NewTextMessage(content string) *Message {
	return &Message{
		Type:    MessageTypeText,
		Content: content,
	}
}

// NewImageMessage 创建图片消息
func NewImageMessage(content string) *Message {
	return &Message{
		Type:    MessageTypeImage,
		Content: content,
	}
}
