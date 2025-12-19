package handlers

import (
	"encoding/base64"
	"io"
	"net/http"

	"phone-server/models"
	"phone-server/services"
	"phone-server/utils"

	"github.com/gin-gonic/gin"
)

// HTTPHandler HTTP接口处理器
type HTTPHandler struct {
	broker *services.Broker // 消息广播服务
}

// NewHTTPHandler 创建HTTP接口处理器实例
func NewHTTPHandler(broker *services.Broker) *HTTPHandler {
	return &HTTPHandler{
		broker: broker,
	}
}

// SendTextMessage 处理发送文本消息的HTTP请求
// @Summary 发送文本消息
// @Description 接收文本消息并通过WebSocket转发给客户端
// @Tags message
// @Accept json
// @Produce json
// @Param message body struct{Content string} true "文本内容"
// @Success 200 {object} struct{Code int; Message string}
// @Failure 400 {object} struct{Code int; Message string}
// @Router /api/message [post]
func (h *HTTPHandler) SendTextMessage(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Errorf("绑定文本消息请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数错误",
		})
		return
	}

	// 创建文本消息
	message := models.NewTextMessage(req.Content)

	// 通过消息广播服务转发消息
	h.broker.BroadcastMessage(message)

	utils.Infof("接收并转发文本消息: %s", req.Content)

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
// @Param image formData file true "图片文件"
// @Success 200 {object} struct{Code int; Message string}
// @Failure 400 {object} struct{Code int; Message string}
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

	// 将图片内容转换为base64编码
	base64Content := base64.StdEncoding.EncodeToString(fileContent)
	// 添加base64前缀，使其符合Data URL格式
	base64Content = "data:image/jpeg;base64," + base64Content

	// 创建图片消息
	message := models.NewImageMessage(base64Content)

	// 通过消息广播服务转发消息
	h.broker.BroadcastMessage(message)

	utils.Infof("接收并转发图片消息，大小: %d bytes", len(fileContent))

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "图片发送成功",
	})
}

// GetStatus 获取服务状态
// @Summary 获取服务状态
// @Description 获取当前连接数和服务状态
// @Tags status
// @Produce json
// @Success 200 {object} struct{Status string; ConnectionCount int}
// @Router /api/status [get]
func (h *HTTPHandler) GetStatus(c *gin.Context) {
	// 获取当前连接数
	connectionCount := h.broker.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"connectionCount": connectionCount,
		"message":         "服务运行正常",
	})
}
