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
	clientIP := c.ClientIP()
	utils.Infofc(c.Request.Context(), "[WS] 收到WebSocket连接请求，客户端IP: %s", clientIP)

	// 从查询参数获取Token
	token := c.Query("token")
	if token == "" {
		// 也可以从Authorization头获取
		token = c.GetHeader("Authorization")
		utils.Debugfc(c.Request.Context(), "[WS] 从Authorization头获取Token: %s", token)
	} else {
		utils.Debugfc(c.Request.Context(), "[WS] 从查询参数获取Token: %s", token)
	}

	if token == "" {
		utils.Warnfc(c.Request.Context(), "[WS] 缺少Token，客户端IP: %s", clientIP)
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "缺少Token"})
		return
	}

	// 解析Token
	claims, err := utils.ParseToken(token, h.jwtSecret)
	if err != nil {
		utils.Errorfc(c.Request.Context(), "[WS] 无效的Token: %v, 客户端IP: %s", err, clientIP)
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "无效的Token: " + err.Error()})
		return
	}
	utils.Infofc(c.Request.Context(), "[WS] Token验证成功，用户ID: %d, 客户端IP: %s", claims.UserID, clientIP)

	// 将HTTP连接升级为WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Errorfc(c.Request.Context(), "[WS] WebSocket连接升级失败: %v, 用户ID: %d, 客户端IP: %s", err, claims.UserID, clientIP)
		return
	}
	utils.Infofc(c.Request.Context(), "[WS] WebSocket连接升级成功，用户ID: %d, 客户端IP: %s", claims.UserID, clientIP)

	// 将客户端注册到消息广播服务
	h.broker.RegisterClient(conn, claims.UserID)
	utils.Infofc(c.Request.Context(), "[WS] 客户端已注册到消息广播服务，用户ID: %d, 客户端IP: %s", claims.UserID, clientIP)

	// 启动协程处理WebSocket连接
	go h.handleConnection(conn, claims.UserID, clientIP)
}

// handleConnection 处理WebSocket连接
func (h *WebSocketHandler) handleConnection(conn *websocket.Conn, userID uint, clientIP string) {
	utils.Infof("[WS] WebSocket连接建立成功，用户ID: %d, 客户端IP: %s", userID, clientIP)

	defer func() {
		// 连接关闭时，将客户端从消息广播服务中注销
		h.broker.UnregisterClient(conn, userID)
		utils.Infof("[WS] WebSocket连接关闭，用户ID: %d, 客户端IP: %s", userID, clientIP)
	}()

	// 循环接收客户端消息
	for {
		// 读取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Errorf("[WS] WebSocket读取消息错误: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
			} else {
				utils.Infof("[WS] WebSocket连接正常关闭: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
			}
			break
		}

		// 仅处理文本消息
		if messageType == websocket.TextMessage {
			utils.Infof("[WS] 用户 %d 收到WebSocket客户端消息: %s, 客户端IP: %s", userID, string(message), clientIP)

			// 解析客户端消息
			msgType, msgContent, err := utils.ParseClientMessage(string(message))
			if err != nil {
				utils.Errorf("[WS] 解析客户端消息失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
				continue
			}
			utils.Debugfc(context.Background(), "[WS] 解析客户端消息成功，类型: %s, 内容: %s, 用户ID: %d, 客户端IP: %s", msgType, msgContent, userID, clientIP)

			// 根据消息类型处理
			switch msgType {
			case "text":
				// 处理文本消息
				h.handleTextMessage(conn, userID, msgContent, clientIP)
			case "image":
				// 处理图片消息
				h.handleImageMessage(conn, userID, msgContent, clientIP)
			default:
				utils.Errorf("[WS] 未知的消息类型: %s, 用户ID: %d, 客户端IP: %s", msgType, userID, clientIP)
			}
		} else {
			utils.Warnf("[WS] 收到非文本消息，类型: %d, 用户ID: %d, 客户端IP: %s", messageType, userID, clientIP)
		}
	}
}

// handleTextMessage 处理客户端发送的文本消息
func (h *WebSocketHandler) handleTextMessage(conn *websocket.Conn, userID uint, content string, clientIP string) {
	utils.Infof("[WS] 用户 %d 处理文本消息: %s, 客户端IP: %s", userID, content, clientIP)

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
		if err := conn.WriteJSON(aiResponse); err != nil {
			utils.Errorf("[WS] 发送AI文本响应失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
			return err
		}
		utils.Debugfc(ctx, "[WS] 发送AI文本响应成功，用户ID: %d, 客户端IP: %s, 响应内容: %s", userID, clientIP, chunk)
		return nil
	}

	// 调用AI服务进行文本对话（流式）
	if err := h.aiService.ChatWithText(ctx, content, streamCallback); err != nil {
		utils.Errorf("[WS] AI文本对话失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
		// 发送错误消息给客户端
		errorMsg := &models.Message{
			Type:    "text",
			Content: "抱歉，AI服务暂时不可用，请稍后重试",
		}
		if err := conn.WriteJSON(errorMsg); err != nil {
			utils.Errorf("[WS] 发送错误消息失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
		}
	}
}

// handleImageMessage 处理客户端发送的图片消息
func (h *WebSocketHandler) handleImageMessage(conn *websocket.Conn, userID uint, imageBase64 string, clientIP string) {
	utils.Infof("[WS] 用户 %d 处理图片消息，图片大小: %d字节, 客户端IP: %s", userID, len(imageBase64), clientIP)

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
		if err := conn.WriteJSON(aiResponse); err != nil {
			utils.Errorf("[WS] 发送AI图片响应失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
			return err
		}
		utils.Debugfc(ctx, "[WS] 发送AI图片响应成功，用户ID: %d, 客户端IP: %s, 响应内容: %s", userID, clientIP, chunk)
		return nil
	}

	// 调用AI服务进行图片对话（流式）
	// 这里可以添加额外的提示文本，例如"请描述这张图片"，或者使用客户端提供的提示
	prompt := "请描述这张图片"
	if err := h.aiService.ChatWithImage(ctx, imageBase64, prompt, streamCallback); err != nil {
		utils.Errorf("[WS] AI图片对话失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
		// 发送错误消息给客户端
		errorMsg := &models.Message{
			Type:    "text",
			Content: "抱歉，AI服务暂时不可用，请稍后重试",
		}
		if err := conn.WriteJSON(errorMsg); err != nil {
			utils.Errorf("[WS] 发送错误消息失败: %v, 用户ID: %d, 客户端IP: %s", err, userID, clientIP)
		}
	}
}
