package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 定义请求结构体
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 定义响应结构体
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	// API 配置
	apiKey := "sk-sE9IjV656TK8ql6jGNjjOWN07fKCpml3RQulGSSLH4qTS0Ue"
	baseURL := "https://api.moonshot.cn/v1"
	url := baseURL + "/chat/completions"

	// 构建请求数据
	requestBody := ChatCompletionRequest{
		Model: "kimi-k2-turbo-preview",
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "你是 Kimi，由 Moonshot AI 提供的人工智能助手，你更擅长中文和英文的对话。你会为用户提供安全，有帮助，准确的回答。同时，你会拒绝一切涉及恐怖主义，种族歧视，黄色暴力等问题的回答。Moonshot AI 为专有名词，不可翻译成其他语言。",
			},
			{
				Role:    "user",
				Content: "你好，我叫李雷，1+1等于多少？",
			},
		},
		Temperature: 0.6,
	}

	// 序列化 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// 创建 HTTP 请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP Error: %d, Response: %s\n", resp.StatusCode, string(body))
		return
	}

	// 解析响应
	var completionResp ChatCompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return
	}

	// 打印响应内容
	if len(completionResp.Choices) > 0 {
		fmt.Println(completionResp.Choices[0].Message.Content)
	} else {
		fmt.Println("No response from Kimi")
	}
}
