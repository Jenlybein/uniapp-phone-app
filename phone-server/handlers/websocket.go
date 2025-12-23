package handlers

import (
	"context"
	"net/http"

	"phone-server/models"
	"phone-server/services"
	"phone-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	broker    *services.Broker    // 消息广播服务
	db        *gorm.DB            // 数据库连接
	aiService *services.AIService // AI服务
	jwtSecret string              // JWT密钥
	upgrader  websocket.Upgrader  // WebSocket连接升级器
}

// NewWebSocketHandler 创建WebSocket处理器实例
func NewWebSocketHandler(broker *services.Broker, db *gorm.DB, aiService *services.AIService, jwtSecret string) *WebSocketHandler {
	return &WebSocketHandler{
		broker:    broker,
		db:        db,
		aiService: aiService,
		jwtSecret: jwtSecret,
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
	// 从查询参数获取Token
	token := c.Query("token")
	if token == "" {
		// 也可以从Authorization头获取
		token = c.GetHeader("Authorization")
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "缺少Token"})
		return
	}

	// 解析Token
	claims, err := utils.ParseToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "无效的Token: " + err.Error()})
		return
	}

	// 将HTTP连接升级为WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Errorf("WebSocket连接升级失败: %v", err)
		return
	}

	// 将客户端注册到消息广播服务
	h.broker.RegisterClient(conn, claims.UserID)

	// 启动协程处理WebSocket连接
	go h.handleConnection(conn, claims.UserID)
}

// handleConnection 处理WebSocket连接
func (h *WebSocketHandler) handleConnection(conn *websocket.Conn, userID uint) {
	defer func() {
		// 连接关闭时，将客户端从消息广播服务中注销
		h.broker.UnregisterClient(conn, userID)
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
			utils.Infof("用户 %d 收到WebSocket客户端消息: %s", userID, message)

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
				h.handleTextMessage(conn, userID, msgContent)
			case "image":
				// 处理图片消息
				h.handleImageMessage(conn, userID, msgContent)
			default:
				utils.Errorf("未知的消息类型: %s", msgType)
			}
		}
	}
}

// handleTextMessage 处理客户端发送的文本消息
func (h *WebSocketHandler) handleTextMessage(conn *websocket.Conn, userID uint, content string) {
	utils.Infof("用户 %d 处理文本消息: %s", userID, content)

	// 创建上下文
	ctx := context.Background()

	// 定义流式响应回调函数
	streamCallback := func(chunk string) error {
		// 构造响应消息
		aiResponse := &models.Message{
			Type:    "text",
			Content: chunk,
		}

		// 将响应发送回当前客户端
		return conn.WriteJSON(aiResponse)
	}

	// 调用AI服务进行文本对话（流式）
	if err := h.aiService.ChatWithText(ctx, content, streamCallback); err != nil {
		utils.Errorf("AI文本对话失败: %v", err)
		// 发送错误消息给客户端
		errorMsg := &models.Message{
			Type:    "text",
			Content: "抱歉，AI服务暂时不可用，请稍后重试",
		}
		conn.WriteJSON(errorMsg)
	}
}

// handleImageMessage 处理客户端发送的图片消息
func (h *WebSocketHandler) handleImageMessage(conn *websocket.Conn, userID uint, imageBase64 string) {
	utils.Infof("用户 %d 处理图片消息，图片大小: %d字节", userID, len(imageBase64))

	// 创建上下文
	ctx := context.Background()

	// 定义流式响应回调函数
	streamCallback := func(chunk string) error {
		// 构造响应消息
		aiResponse := &models.Message{
			Type:    "text",
			Content: chunk,
		}

		// 将响应发送回当前客户端
		return conn.WriteJSON(aiResponse)
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
		conn.WriteJSON(errorMsg)
	}
}
