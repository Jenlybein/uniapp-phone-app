package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"phone-server/utils"
)

// AIService AI服务
type AIService struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	model    string
	thinking string
}

// NewAIService 创建AI服务实例
func NewAIService(apiKey string, baseURL string, model string, thinking string) *AIService {
	client := &http.Client{}
	return &AIService{
		client:   client,
		apiKey:   apiKey,
		baseURL:  baseURL,
		model:    model,
		thinking: thinking,
	}
}

// StreamResponseFunc 流式响应回调函数类型
type StreamResponseFunc func(chunk string) error

// SSEMessage 表示SSE消息结构（仅保留实际使用的Data字段）
type SSEMessage struct {
	Data string `json:"data,omitempty"`
}

// ChatWithText 与AI进行文本对话（流式）
func (s *AIService) ChatWithText(ctx context.Context, content string, streamCallback StreamResponseFunc) error {
	// 记录请求开始时间
	startTime := time.Now()

	utils.Infofc(ctx, "[AI_REQUEST] 开始发送文本到AI，内容: %s", content)

	// 构建请求URL
	url := fmt.Sprintf("%s/chat/completions", s.baseURL)
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": content,
			},
		},
		"thinking": map[string]string{
			"type": s.thinking,
		},
		"stream": true,
	}

	// 转换为JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		utils.Errorf("JSON编码失败: %v", err)
		return err
	}

	// 记录请求详细信息
	utils.Debugfc(ctx, "[AI_REQUEST] 请求URL: %s, 请求体: %s", url, string(jsonData))

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Errorf("创建请求失败: %v", err)
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		utils.Errorfc(ctx, "[AI_REQUEST] 发送请求失败: %v, 耗时: %v", err, time.Since(startTime))
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		utils.Errorfc(ctx, "[AI_REQUEST] 请求失败，状态码: %d, 耗时: %v", resp.StatusCode, time.Since(startTime))
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 记录响应状态
	utils.Infofc(ctx, "[AI_REQUEST] 请求成功，状态码: %d, 耗时: %v", resp.StatusCode, time.Since(startTime))

	// 处理SSE响应
	utils.Infofc(ctx, "[AI_RESPONSE] 开始接收AI流式响应")

	// 创建SSE解析器
	reader := bufio.NewReader(resp.Body)
	var buffer string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 流式响应结束
				utils.Infofc(ctx, "[AI_RESPONSE] AI流式响应结束，总耗时: %v", time.Since(startTime))
				break
			}
			utils.Errorfc(ctx, "[AI_RESPONSE] 读取响应失败: %v", err)
			return err
		}

		// 添加到缓冲区
		buffer += line

		// 检查完整SSE事件（以两个换行符结束）
		if strings.HasSuffix(buffer, "\n\n") {
			// 解析SSE事件
			lines := strings.Split(buffer, "\n")

			var sseMessage SSEMessage
			for _, l := range lines {
				if l == "" {
					continue
				}

				parts := strings.SplitN(l, ": ", 2)
				if len(parts) != 2 {
					continue
				}

				key := parts[0]
				value := parts[1]

				// 只处理data字段，其他字段暂时不需要
				if key == "data" {
					sseMessage.Data = value
				}
			}

			// 处理SSE数据
			if sseMessage.Data != "" {
				if sseMessage.Data == "[DONE]" {
					// 流式结束标记
					break
				}

				// 解析AI响应数据
				var aiResponse struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(sseMessage.Data), &aiResponse); err != nil {
					utils.Errorfc(ctx, "[AI_RESPONSE] 解析AI响应失败: %v", err)
					continue
				}

				// 提取内容并调用回调
				if len(aiResponse.Choices) > 0 {
					chunk := aiResponse.Choices[0].Delta.Content
					if chunk != "" {
						if err := streamCallback(chunk); err != nil {
							utils.Errorfc(ctx, "[AI_RESPONSE] 流式响应回调处理失败: %v", err)
							return err
						}
					}
				}
			}

			// 重置缓冲区
			buffer = ""
		}
	}

	return nil
}

// ChatWithImage 与AI进行图片对话（流式）
func (s *AIService) ChatWithImage(ctx context.Context, imageBase64 string, content string, streamCallback StreamResponseFunc) error {
	// 记录请求开始时间
	startTime := time.Now()

	utils.Infofc(ctx, "[AI_REQUEST] 开始发送图片和文本到AI")

	// 构建请求URL
	url := fmt.Sprintf("%s/chat/completions", s.baseURL)
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"multi_content": []map[string]interface{}{
					{
						"type": "text",
						"text": content,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
						},
					},
				},
			},
		},
		"stream": true,
	}

	// 添加thinking参数
	if s.thinking != "" {
		reqBody["thinking"] = s.thinking
		utils.Infofc(ctx, "[AI_REQUEST] 设置AI思考模式: %s", s.thinking)
	}

	// 转换为JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		utils.Errorf("JSON编码失败: %v", err)
		return err
	}

	// 记录请求详细信息
	utils.Debugfc(ctx, "[AI_REQUEST] 请求URL: %s, 请求体: %s", url, string(jsonData))

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Errorfc(ctx, "[AI_REQUEST] 创建请求失败: %v", err)
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		utils.Errorfc(ctx, "[AI_REQUEST] 发送请求失败: %v, 耗时: %v", err, time.Since(startTime))
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		utils.Errorfc(ctx, "[AI_REQUEST] 请求失败，状态码: %d, 耗时: %v", resp.StatusCode, time.Since(startTime))
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 记录响应状态
	utils.Infofc(ctx, "[AI_REQUEST] 请求成功，状态码: %d, 耗时: %v", resp.StatusCode, time.Since(startTime))

	// 处理SSE响应
	utils.Infofc(ctx, "[AI_RESPONSE] 开始接收AI流式响应")

	// 创建SSE解析器
	reader := bufio.NewReader(resp.Body)
	var buffer string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 流式响应结束
				utils.Infofc(ctx, "[AI_RESPONSE] AI流式响应结束，总耗时: %v", time.Since(startTime))
				break
			}
			utils.Errorfc(ctx, "[AI_RESPONSE] 读取响应失败: %v", err)
			return err
		}

		// 添加到缓冲区
		buffer += line

		// 检查完整SSE事件（以两个换行符结束）
		if strings.HasSuffix(buffer, "\n\n") {
			// 解析SSE事件
			lines := strings.Split(buffer, "\n")

			var sseMessage SSEMessage
			for _, l := range lines {
				if l == "" {
					continue
				}

				parts := strings.SplitN(l, ": ", 2)
				if len(parts) != 2 {
					continue
				}

				key := parts[0]
				value := parts[1]

				// 只处理data字段，其他字段暂时不需要
				if key == "data" {
					sseMessage.Data = value
				}
			}

			// 处理SSE数据
			if sseMessage.Data != "" {
				if sseMessage.Data == "[DONE]" {
					// 流式结束标记
					break
				}

				// 解析AI响应数据
				var aiResponse struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(sseMessage.Data), &aiResponse); err != nil {
					utils.Errorfc(ctx, "[AI_RESPONSE] 解析AI响应失败: %v", err)
					continue
				}

				// 提取内容并调用回调
				if len(aiResponse.Choices) > 0 {
					chunk := aiResponse.Choices[0].Delta.Content
					if chunk != "" {
						if err := streamCallback(chunk); err != nil {
							utils.Errorfc(ctx, "[AI_RESPONSE] 流式响应回调处理失败: %v", err)
							return err
						}
					}
				}
			}

			// 重置缓冲区
			buffer = ""
		}
	}

	return nil
}


