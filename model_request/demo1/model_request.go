package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// 定义请求结构体
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float32       `json:"temperature"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 定义流式响应的 Chunk 结构（简化，只解析我们需要的字段）
type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
}

func main() {
	// 1. 准备请求参数
	apiKey := "sk-sE9IjV656TK8ql6jGNjjOWN07fKCpml3RQulGSSLH4qTS0Ue" // 请替换为你的真实 API Key
	baseURL := "https://api.moonshot.cn/v1"                         // 你的目标 base URL
	url := baseURL + "/chat/completions"

	// 2. 构造请求体
	reqBody := ChatCompletionRequest{
		Model: "kimi-k2-turbo-preview",
		Messages: []ChatMessage{
			{Role: "system", Content: "你是一个加法助手，遇到加法问题请调用工具add，最后用自然语言回答用户。"},
			{Role: "user", Content: "Say Hi!"},
		},
		Stream:      true,
		Temperature: 0.3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("JSON 编码失败: %v", err)
	}

	// 3. 创建 HTTP 请求
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	// 4. 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 5. 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 6. 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 7. 检查 HTTP 响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("非 200 响应: %d, 响应内容: %s", resp.StatusCode, string(bodyBytes))
	}

	// 8. 处理流式返回（逐行读取，类似 Server-Sent Events 或每行一个 JSON）
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// 可能的格式是 "data: {...}"，我们兼容这种情况
		var chunk StreamChunk
		if bytes.HasPrefix(line, []byte("data: ")) {
			line = line[6:] // 去掉 "data: "
		}

		// 尝试解析 JSON
		if err := json.Unmarshal(line, &chunk); err != nil {
			// 如果不是合法的 JSON，可能是心跳包或空行，跳过
			continue
		}

		// 打印 delta content
		for _, choice := range chunk.Choices {
			if content := choice.Delta.Content; content != "" {
				fmt.Print(content) // 实时打印，不换行
			}
		}
	}

	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		log.Printf("读取流时出错: %v", err)
	}
}
