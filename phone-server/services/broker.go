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
	clients    map[uint]map[*websocket.Conn]bool // 客户端连接集合，按用户ID分组
	clientsMux sync.Mutex                        // 保护clients的互斥锁
	register   chan struct {
		client *websocket.Conn
		userID uint
	}                                            // 注册客户端的通道
	unregister chan struct {
		client *websocket.Conn
		userID uint
	}                                            // 注销客户端的通道
	broadcast  chan struct {
		message *models.Message
		userID  uint
	}                                            // 广播消息的通道
}

// NewBroker 创建消息广播服务实例
func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[uint]map[*websocket.Conn]bool),
		register:   make(chan struct{ client *websocket.Conn; userID uint }),
		unregister: make(chan struct{ client *websocket.Conn; userID uint }),
		broadcast:  make(chan struct{ message *models.Message; userID uint }),
	}
}

// Start 启动消息广播服务
func (b *Broker) Start() {
	for {
		select {
		// 注册新客户端
		case reg := <-b.register:
			b.clientsMux.Lock()
			// 如果用户ID对应的客户端映射不存在，则创建
			if _, ok := b.clients[reg.userID]; !ok {
				b.clients[reg.userID] = make(map[*websocket.Conn]bool)
			}
			// 将客户端添加到用户ID对应的映射中
			b.clients[reg.userID][reg.client] = true
			b.clientsMux.Unlock()
			utils.Infof("用户 %d 的ws客户端:%s已连接", reg.userID, reg.client.RemoteAddr().String())

		// 注销客户端
		case unreg := <-b.unregister:
			b.clientsMux.Lock()
			// 检查用户ID对应的客户端映射是否存在
			if clientMap, ok := b.clients[unreg.userID]; ok {
				// 检查客户端是否存在于映射中
				if _, ok := clientMap[unreg.client]; ok {
					// 移除客户端并关闭连接
					delete(clientMap, unreg.client)
					unreg.client.Close()
					utils.Infof("用户 %d 的ws客户端:%s已断开连接", unreg.userID, unreg.client.RemoteAddr().String())
					// 如果用户ID对应的客户端映射为空，则删除该映射
					if len(clientMap) == 0 {
						delete(b.clients, unreg.userID)
					}
				}
			}
			b.clientsMux.Unlock()

		// 广播消息给特定用户的所有客户端
		case msg := <-b.broadcast:
			b.clientsMux.Lock()

			// 将消息序列化为JSON
			msgBytes, err := json.Marshal(msg.message)
			if err != nil {
				utils.Errorf("消息序列化失败: %v", err)
				b.clientsMux.Unlock()
				continue
			}

			// 发送消息给特定用户的所有客户端
			if clientMap, ok := b.clients[msg.userID]; ok {
				for client := range clientMap {
					if err := client.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
						utils.Errorf("发送消息失败: %v", err)
						// 发送失败，关闭连接并从集合中移除
						client.Close()
						delete(clientMap, client)
					}
				}
				utils.Infof("已向用户 %d 广播消息，类型: %s，客户端数: %d", msg.userID, msg.message.Type, len(clientMap))
			}
			b.clientsMux.Unlock()
		}
	}
}

// RegisterClient 注册WebSocket客户端
func (b *Broker) RegisterClient(client *websocket.Conn, userID uint) {
	b.register <- struct{ client *websocket.Conn; userID uint }{client: client, userID: userID}
}

// UnregisterClient 注销WebSocket客户端
func (b *Broker) UnregisterClient(client *websocket.Conn, userID uint) {
	b.unregister <- struct{ client *websocket.Conn; userID uint }{client: client, userID: userID}
}

// BroadcastMessage 广播消息给特定用户的所有客户端
func (b *Broker) BroadcastMessage(message *models.Message, userID uint) {
	b.broadcast <- struct{ message *models.Message; userID uint }{message: message, userID: userID}
}

// GetClientCount 获取当前客户端连接数
func (b *Broker) GetClientCount() int {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	total := 0
	for _, clientMap := range b.clients {
		total += len(clientMap)
	}
	return total
}

// GetClientCountByUserID 获取特定用户的客户端连接数
func (b *Broker) GetClientCountByUserID(userID uint) int {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	if clientMap, ok := b.clients[userID]; ok {
		return len(clientMap)
	}
	return 0
}
