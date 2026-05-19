package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gin_py/internal/model"
	"github.com/gin-gonic/gin"
)

const pythonServerURL = "http://127.0.0.1:8081/api/chat-process"

type ChatHandler struct{}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{}
}

// ChatProcess 处理流式请求（修正版）
func (h *ChatHandler) ChatProcess(c *gin.Context) {
	startTime := time.Now()
	fmt.Printf("[Handler] 收到请求: %s\n", startTime.Format("15:04:05.000"))

	var req model.ChatProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Fail", "message": err.Error()})
		return
	}

	// 构建 OpenAI 格式请求体
	messages := []map[string]string{}
	if req.SystemMessage != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.SystemMessage})
	}
	messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})

	openAIReq := map[string]interface{}{
		"model":    "kimi-k2.6",
		"messages": messages,
		"stream":   true,
	}
	bodyBytes, _ := json.Marshal(openAIReq)

	// 创建上游请求（使用 request context，支持取消）
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", pythonServerURL, bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 过滤逐跳头
	for k, v := range c.Request.Header {
		if !isHopHeader(k) && k != "Content-Length" {
			httpReq.Header[k] = v
		}
	}

	// 重要：Timeout = 0 禁用超时
	client := &http.Client{Timeout: 0}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("[Handler] 请求 Python 失败: %v\n", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "上游服务不可用"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", errBody)
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Header("Access-Control-Allow-Origin", "*")

	// 转换并转发流
	h.forwardStream(c, resp.Body, startTime)
}

// forwardStream 解析 Python 的原始 SSE 并转换为前端事件
func (h *ChatHandler) forwardStream(c *gin.Context, body io.ReadCloser, startTime time.Time) {
	reader := bufio.NewReader(body)
	totalTokens := 0

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			fmt.Printf("[Handler] 客户端断开连接\n")
			return false
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				sendSSEEvent(w, "done", map[string]interface{}{"status": "complete", "totalTokens": totalTokens})
				fmt.Printf("[Handler] 流式响应结束，总耗时: %v\n", time.Since(startTime))
			} else {
				sendSSEEventQuiet(w, "error", map[string]string{"error": err.Error()})
			}
			return false
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			return true // 忽略非 data 行
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			sendSSEEvent(w, "done", map[string]interface{}{"status": "complete", "totalTokens": totalTokens})
			fmt.Printf("[Handler] 收到 [DONE]\n")
			return false
		}

		var chunk struct {
			ID      string `json:"id"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			fmt.Printf("[Handler] JSON 解析失败: %v\n", err)
			return true
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			totalTokens++
			msg := model.ChatMessageResponse{
				ID:    chunk.ID,
				Text:  chunk.Choices[0].Delta.Content,
				Delta: chunk.Choices[0].Delta.Content,
				Role:  "assistant",
				Detail: map[string]interface{}{
					"raw": chunk,
				},
			}
			sendSSEEvent(w, "message", msg)
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason == "stop" {
			sendSSEEvent(w, "done", map[string]interface{}{"status": "complete", "totalTokens": totalTokens})
			fmt.Printf("[Handler] 收到 stop\n")
			return false
		}

		return true
	})
}

func sendSSEEvent(w io.Writer, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func sendSSEEventQuiet(w io.Writer, event string, data interface{}) {
	defer func() { recover() }()
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func isHopHeader(header string) bool {
	hopHeaders := map[string]bool{
		"Connection": true, "Keep-Alive": true, "Proxy-Connection": true,
		"Transfer-Encoding": true, "Upgrade": true, "TE": true, "Trailer": true,
	}
	return hopHeaders[header]
}

func (h *ChatHandler) Config(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "Success", "data": gin.H{"model": "kimi-k2.6"}})
}
