package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"phone-server/utils"

	openai "github.com/sashabaranov/go-openai"
)

// AIService AI服务
type AIService struct {
	client *openai.Client
	model  string
}

// NewAIService 创建AI服务实例
func NewAIService(apiKey string, baseURL string, model string) *AIService {
	config := openai.DefaultConfig(apiKey)
	// 如果需要使用豆包API，可以设置baseURL
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)
	return &AIService{
		client: client,
		model:  model,
	}
}

// StreamResponseFunc 流式响应回调函数类型
type StreamResponseFunc func(chunk string) error

// ChatWithText 与AI进行文本对话（流式）
func (s *AIService) ChatWithText(ctx context.Context, content string, streamCallback StreamResponseFunc) error {
	utils.Infof("发送文本到AI: %s", content)

	// 创建流式请求
	req := openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
		Stream: true,
	}

	// 发送流式请求
	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		utils.Errorf("AI流式请求失败: %v", err)
		return err
	}
	defer stream.Close()

	// 处理流式响应
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// 流式响应结束
				utils.Infof("AI流式响应结束")
				return nil
			}
			utils.Errorf("AI流式响应接收失败: %v", err)
			return err
		}

		// 提取当前响应的文本
		chunk := response.Choices[0].Delta.Content
		if chunk != "" {
			// 调用回调函数处理当前响应块
			if err := streamCallback(chunk); err != nil {
				utils.Errorf("流式响应回调处理失败: %v", err)
				return err
			}
		}
	}
}

// ChatWithImage 与AI进行图片对话（流式）
func (s *AIService) ChatWithImage(ctx context.Context, imageBase64 string, content string, streamCallback StreamResponseFunc) error {
	utils.Infof("发送图片和文本到AI")

	// 创建流式请求
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4VisionPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: content,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
						},
					},
				},
			},
		},
		Stream: true,
	}

	// 发送流式请求
	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		utils.Errorf("AI流式请求失败: %v", err)
		return err
	}
	defer stream.Close()

	// 处理流式响应
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// 流式响应结束
				utils.Infof("AI流式响应结束")
				return nil
			}
			utils.Errorf("AI流式响应接收失败: %v", err)
			return err
		}

		// 提取当前响应的文本
		chunk := response.Choices[0].Delta.Content
		if chunk != "" {
			// 调用回调函数处理当前响应块
			if err := streamCallback(chunk); err != nil {
				utils.Errorf("流式响应回调处理失败: %v", err)
				return err
			}
		}
	}
}

// ParseClientMessage 解析客户端消息
func ParseClientMessage(message string) (string, string, error) {
	// 简单的消息解析，实际应用中可以根据需要实现更复杂的解析逻辑
	// 假设消息格式为：{"type":"text","content":"xxx"} 或 {"type":"image","content":"base64..."}
	var msg struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		utils.Errorf("解析客户端消息失败: %v", err)
		return "", "", err
	}

	return msg.Type, msg.Content, nil
}
