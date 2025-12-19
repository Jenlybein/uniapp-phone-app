package services

import (
	"encoding/json"
	"sync"

	"phone-server/models"
	"phone-server/utils"

	"github.com/gorilla/websocket"
)

// Broker 消息广播服务
type Broker struct {
	clients    map[*websocket.Conn]bool // 客户端连接集合
	clientsMux sync.Mutex               // 保护clients的互斥锁
	register   chan *websocket.Conn     // 注册客户端的通道
	unregister chan *websocket.Conn     // 注销客户端的通道
	broadcast  chan *models.Message     // 广播消息的通道
}

// NewBroker 创建消息广播服务实例
func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan *models.Message),
	}
}

// Start 启动消息广播服务
func (b *Broker) Start() {
	for {
		select {
		// 注册新客户端
		case client := <-b.register:
			b.clientsMux.Lock()
			b.clients[client] = true
			b.clientsMux.Unlock()
			utils.Infof("ws客户端:%s已连接，当前连接数: %d", client.RemoteAddr().String(), len(b.clients))

		// 注销客户端
		case client := <-b.unregister:
			b.clientsMux.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				client.Close()
				utils.Infof("ws客户端:%s已断开连接，当前连接数: %d", client.RemoteAddr().String(), len(b.clients))
			}
			b.clientsMux.Unlock()

		// 广播消息给所有客户端
		case message := <-b.broadcast:
			b.clientsMux.Lock()

			// 将消息序列化为JSON
			msgBytes, err := json.Marshal(message)
			if err != nil {
				utils.Errorf("消息序列化失败: %v", err)
				b.clientsMux.Unlock()
				continue
			}

			// 发送消息给所有客户端
			for client := range b.clients {
				if err := client.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					utils.Errorf("发送消息失败: %v", err)
					// 发送失败，关闭连接并从集合中移除
					client.Close()
					delete(b.clients, client)
				}
			}
			utils.Infof("已广播消息，类型: %s，当前连接数: %d", message.Type, len(b.clients))
			b.clientsMux.Unlock()
		}
	}
}

// RegisterClient 注册WebSocket客户端
func (b *Broker) RegisterClient(client *websocket.Conn) {
	b.register <- client
}

// UnregisterClient 注销WebSocket客户端
func (b *Broker) UnregisterClient(client *websocket.Conn) {
	b.unregister <- client
}

// BroadcastMessage 广播消息
func (b *Broker) BroadcastMessage(message *models.Message) {
	b.broadcast <- message
}

// GetClientCount 获取当前客户端连接数
func (b *Broker) GetClientCount() int {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	return len(b.clients)
}
