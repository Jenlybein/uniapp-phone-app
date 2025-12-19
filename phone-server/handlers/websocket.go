package handlers

import (
	"context"
	"net/http"
	"time"

	"phone-server/models"
	"phone-server/services"
	"phone-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	broker    *services.Broker    // 消息广播服务
	aiService *services.AIService // AI服务
	upgrader  websocket.Upgrader  // WebSocket连接升级器
}

// NewWebSocketHandler 创建WebSocket处理器实例
func NewWebSocketHandler(broker *services.Broker, aiService *services.AIService) *WebSocketHandler {
	return &WebSocketHandler{
		broker:    broker,
		aiService: aiService,
		upgrader: websocket.Upgrader{
			// 允许所有来源的跨域请求
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// HandleWebSocket 处理WebSocket连接请求
// @Summary WebSocket连接
// @Description 建立WebSocket连接，用于接收实时消息
// @Tags websocket
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 将HTTP连接升级为WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Errorf("WebSocket连接升级失败: %v", err)
		return
	}

	// 将客户端注册到消息广播服务
	h.broker.RegisterClient(conn)

	// 启动协程处理WebSocket连接
	go h.handleConnection(conn)
}

// handleConnection 处理WebSocket连接
func (h *WebSocketHandler) handleConnection(conn *websocket.Conn) {
	defer func() {
		// 连接关闭时，将客户端从消息广播服务中注销
		h.broker.UnregisterClient(conn)
	}()

	// 循环接收客户端消息
	for {
		// 读取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Errorf("WebSocket读取消息错误: %v", err)
			}
			break
		}

		// 仅处理文本消息
		if messageType == websocket.TextMessage {
			utils.Infof("收到WebSocket客户端消息: %s", message)

			// 解析客户端消息
			msgType, msgContent, err := services.ParseClientMessage(string(message))
			if err != nil {
				utils.Errorf("解析客户端消息失败: %v", err)
				continue
			}

			// 根据消息类型处理
			switch msgType {
			case "text":
				// 处理文本消息
				h.handleTextMessage(conn, msgContent)
			case "image":
				// 处理图片消息
				h.handleImageMessage(conn, msgContent)
			case "ping":
				// 处理心跳消息
				utils.Infof("收到心跳消息，来自客户端: %s", conn.RemoteAddr().String())
				// 可以发送pong消息作为响应
				pongMsg := &models.Message{
					Type:    "pong",
					Content: "",
				}
				if err := conn.WriteJSON(pongMsg); err != nil {
					utils.Errorf("发送pong消息失败: %v", err)
				}
			default:
				utils.Errorf("未知的消息类型: %s", msgType)
			}
		}
	}
}

// handleTextMessage 处理客户端发送的文本消息
func (h *WebSocketHandler) handleTextMessage(conn *websocket.Conn, content string) {
	utils.Infof("处理文本消息: %s", content)

	// 创建带超时的上下文，防止长时间阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义流式响应回调函数
	streamCallback := func(chunk string) error {
		// 构造响应消息
		aiResponse := &models.Message{
			Type:    "text",
			Content: chunk,
		}

		// 将响应发送回当前客户端
		if err := conn.WriteJSON(aiResponse); err != nil {
			utils.Errorf("发送AI文本响应失败: %v", err)
			return err
		}
		return nil
	}

	// 调用AI服务进行文本对话（流式）
	if err := h.aiService.ChatWithText(ctx, content, streamCallback); err != nil {
		utils.Errorf("AI文本对话失败: %v", err)
		// 发送错误消息给客户端
		errorMsg := &models.Message{
			Type:    "text",
			Content: "抱歉，AI服务暂时不可用，请稍后重试",
		}
		if err := conn.WriteJSON(errorMsg); err != nil {
			utils.Errorf("发送错误消息失败: %v", err)
		}
	}
}

// handleImageMessage 处理客户端发送的图片消息
func (h *WebSocketHandler) handleImageMessage(conn *websocket.Conn, imageBase64 string) {
	utils.Infof("处理图片消息，图片大小: %d字节", len(imageBase64))

	// 创建带超时的上下文，防止长时间阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义流式响应回调函数
	streamCallback := func(chunk string) error {
		// 构造响应消息
		aiResponse := &models.Message{
			Type:    "text",
			Content: chunk,
		}

		// 将响应发送回当前客户端
		if err := conn.WriteJSON(aiResponse); err != nil {
			utils.Errorf("发送AI图片响应失败: %v", err)
			return err
		}
		return nil
	}

	// 调用AI服务进行图片对话（流式）
	// 这里可以添加额外的提示文本，例如"请描述这张图片"，或者使用客户端提供的提示
	prompt := "请描述这张图片"
	if err := h.aiService.ChatWithImage(ctx, imageBase64, prompt, streamCallback); err != nil {
		utils.Errorf("AI图片对话失败: %v", err)
		// 发送错误消息给客户端
		errorMsg := &models.Message{
			Type:    "text",
			Content: "抱歉，AI服务暂时不可用，请稍后重试",
		}
		if err := conn.WriteJSON(errorMsg); err != nil {
			utils.Errorf("发送错误消息失败: %v", err)
		}
	}
}
