package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"phone-server/models"
	"phone-server/services"
	"phone-server/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HTTPHandler HTTP接口处理器
type HTTPHandler struct {
	broker    *services.Broker    // 消息广播服务
	db        *gorm.DB            // 数据库连接
	aiService *services.AIService // AI服务
}

// SendTextMessageRequest 发送文本消息请求参数
type SendTextMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// ChatWithAIRequest 与AI聊天请求参数
type ChatWithAIRequest struct {
	Type    string `json:"type" binding:"required,oneof=text image"`
	Content string `json:"content" binding:"required"`
}

// NewHTTPHandler 创建HTTP接口处理器实例
func NewHTTPHandler(broker *services.Broker, db *gorm.DB, aiService *services.AIService) *HTTPHandler {
	return &HTTPHandler{
		broker:    broker,
		db:        db,
		aiService: aiService,
	}
}

// SendTextMessage 处理发送文本消息的HTTP请求
// @Summary 发送文本消息
// @Description 接收文本消息并通过WebSocket转发给客户端
// @Tags message
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param message body SendTextMessageRequest true "文本内容"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/message [post]
func (h *HTTPHandler) SendTextMessage(c *gin.Context) {
	var req SendTextMessageRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Errorf("绑定文本消息请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数错误",
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的请求",
		})
		return
	}

	// 创建文本消息
	message := models.NewTextMessage(userID.(uint), req.Content, models.SenderTypePC)

	// 将消息存储到数据库
	if result := h.db.Create(message); result.Error != nil {
		utils.Errorf("保存文本消息失败: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "保存消息失败",
		})
		return
	}

	// 通过消息广播服务转发消息
	h.broker.BroadcastMessage(message, userID.(uint))

	utils.Infof("用户 %d 发送文本消息: %s", userID.(uint), req.Content)

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "消息发送成功",
	})
}

// SendImageMessage 处理发送图片消息的HTTP请求
// @Summary 发送图片消息
// @Description 接收图片文件并通过WebSocket转发给客户端
// @Tags message
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param image formData file true "图片文件"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/image [post]
func (h *HTTPHandler) SendImageMessage(c *gin.Context) {
	// 获取上传的文件
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		utils.Errorf("获取图片文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "获取图片文件失败",
		})
		return
	}
	defer file.Close()

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		utils.Errorf("读取图片文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "读取图片文件失败",
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的请求",
		})
		return
	}

	// 将图片内容转换为base64编码
	base64Content := base64.StdEncoding.EncodeToString(fileContent)
	// 添加base64前缀，使其符合Data URL格式
	base64Content = "data:image/jpeg;base64," + base64Content

	// 创建图片消息
	message := models.NewImageMessage(userID.(uint), base64Content, models.SenderTypePC)

	// 将消息存储到数据库
	if result := h.db.Create(message); result.Error != nil {
		utils.Errorf("保存图片消息失败: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "保存消息失败",
		})
		return
	}

	// 通过消息广播服务转发消息
	h.broker.BroadcastMessage(message, userID.(uint))

	utils.Infof("用户 %d 发送图片消息，大小: %d bytes", userID.(uint), len(fileContent))

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "图片发送成功",
	})
}

// ChatWithAI 处理与AI聊天的HTTP请求
// @Summary 与AI聊天
// @Description 接收文本或图片，获取AI回复（支持普通HTTP和SSE流式输出）
// @Tags ai
// @Accept json
// @Produce json
// @Produce text/event-stream
// @Security ApiKeyAuth
// @Param request body ChatWithAIRequest true "聊天请求" SchemaExample({"type": "text", "content": "你好"})
// @Success 200 {object} map[string]interface{} "AI回复"
// @Success 200 {string} text/event-stream "AI回复流"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/ai/chat [post]
func (h *HTTPHandler) ChatWithAI(c *gin.Context) {
	// 从上下文获取用户ID
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的请求",
		})
		return
	}

	// 绑定请求参数
	var req ChatWithAIRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Errorf("绑定AI聊天请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数错误",
		})
		return
	}

	// 创建上下文
	ctx := c.Request.Context()

	// 检查是否支持SSE流式响应
	acceptHeader := c.GetHeader("Accept")
	supportsSSE := strings.Contains(acceptHeader, "text/event-stream")

	if supportsSSE {
		// SSE流式响应模式
		// 设置响应头，支持流式输出
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
		c.Header("Access-Control-Allow-Origin", "*")

		// 定义流式响应回调函数
		streamCallback := func(chunk string) error {
			// 发送SSE格式的响应
			if _, err := c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", chunk)); err != nil {
				utils.Errorf("发送AI流式响应失败: %v", err)
				return err
			}
			c.Writer.Flush()
			return nil
		}

		// 调用AI服务
		var err error
		switch req.Type {
		case "text":
			// 文本聊天
			err = h.aiService.ChatWithText(ctx, req.Content, streamCallback)
		case "image":
			// 图片聊天
			// 提取base64内容（去掉前缀）
			base64Content := req.Content
			if strings.HasPrefix(base64Content, "data:image/") {
				// 去掉前缀
				base64Content = base64Content[strings.Index(base64Content, ",")+1:]
			}
			err = h.aiService.ChatWithImage(ctx, base64Content, "请描述这张图片", streamCallback)
		}

		if err != nil {
			utils.Errorf("AI聊天失败: %v", err)
			// 发送错误消息
			c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", "抱歉，AI服务暂时不可用，请稍后重试"))
			c.Writer.Flush()
		}
	} else {
		// 普通HTTP响应模式
		var fullResponse strings.Builder

		// 定义收集完整响应的回调函数
		collectCallback := func(chunk string) error {
			fullResponse.WriteString(chunk)
			return nil
		}

		// 调用AI服务
		var err error
		switch req.Type {
		case "text":
			// 文本聊天
			err = h.aiService.ChatWithText(ctx, req.Content, collectCallback)
		case "image":
			// 图片聊天
			// 提取base64内容（去掉前缀）
			base64Content := req.Content
			if strings.HasPrefix(base64Content, "data:image/") {
				// 去掉前缀
				base64Content = base64Content[strings.Index(base64Content, ",")+1:]
			}
			err = h.aiService.ChatWithImage(ctx, base64Content, "请描述这张图片", collectCallback)
		}

		if err != nil {
			utils.Errorf("AI聊天失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "AI请求失败，请稍后重试",
			})
			return
		}

		// 返回完整的JSON响应
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"content": fullResponse.String(),
		})
	}
}
