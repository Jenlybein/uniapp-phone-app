package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 定义请求数据
	reqData := map[string]string{
		"type":    "text",
		"content": "你好，这是一个测试请求",
	}
	
	// 转换为JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		fmt.Printf("JSON编码失败: %v\n", err)
		return
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", "http://localhost:8080/api/ai/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6IjEyMzQ1NiIsInN1YiI6InVzZXJfYXV0aCIsImV4cCI6MTc2NjU2NzI4MiwiaWF0IjoxNzY2NDgwODgyfQ.sWwjf9esF9yRntT-_B8QzJ-qSMMtnB6jxwLMlw00t_8")
	req.Header.Set("Accept", "text/event-stream")
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应头: %v\n", resp.Header)
	
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}
	
	fmt.Printf("响应体: %s\n", string(body))
}
