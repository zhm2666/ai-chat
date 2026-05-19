package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"new_gin/internal/model"
	"new_gin/pkg/openai"

	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	client *openai.Client
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(client *openai.Client) *ChatHandler {
	return &ChatHandler{
		client: client,
	}
}

// ChatProcess SSE 流式聊天处理
func (h *ChatHandler) ChatProcess(c *gin.Context) {
	startTime := time.Now()
	fmt.Printf("[Handler] 收到请求，时间: %s\n", startTime.Format("15:04:05.000"))

	// 1. 解析请求
	var req model.ChatProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Fail",
			"message": fmt.Sprintf("请求解析失败: %v", err),
		})
		return
	}

	fmt.Printf("[Handler] 请求内容: prompt=%q, systemMessage=%q, options=%+v\n",
		req.Prompt, req.SystemMessage, req.Options)

	// 2. 构建消息列表
	messages := []openai.ChatMessage{}
	// 添加系统提示词（如果有）
	if req.SystemMessage != "" {
		messages = append(messages, openai.ChatMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemMessage,
		})
	}

	// 添加用户消息
	messages = append(messages, openai.ChatMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Prompt,
	})

	// 3. 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Header("Access-Control-Allow-Origin", "*")

	// 4. 创建流式请求（使用 Gin 的 context，支持客户端断开连接检测）
	stream, err := h.client.CreateChatCompletionStream(c.Request.Context(), messages)
	if err != nil {
		// 检查是否是客户端断开连接
		if c.Request.Context().Err() == context.Canceled {
			fmt.Printf("[Handler] 客户端断开连接\n")
		} else {
			fmt.Printf("[Handler] 创建流失败: %v\n", err)
			// 发送错误事件
			sendSSEEvent(c.Writer, "error", map[string]string{"error": err.Error()})
		}
		return
	}
	defer stream.Close()

	// 5. 处理流式响应
	totalTokens := 0
	c.Stream(func(w io.Writer) bool {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// 流结束，发送结束事件
				sendSSEEvent(w, "done", map[string]interface{}{"status": "complete", "totalTokens": totalTokens})
				fmt.Printf("[Handler] 流式响应结束，总耗时: %v\n", time.Since(startTime))
				return false
			}

			// 检查是否是客户端断开连接
			errMsg := err.Error()
			if strings.Contains(errMsg, "context canceled") || strings.Contains(errMsg, "Canceled") {
				fmt.Printf("[Handler] 客户端断开连接\n")
			} else {
				fmt.Printf("[Handler] 流式响应错误: %v\n", err)
				// 尝试发送错误事件，如果失败就算了（可能是连接已断开）
				sendSSEEventQuiet(w, "error", map[string]string{"error": errMsg})
			}
			return false
		}

		// 构建响应消息
		if len(resp.Choices) > 0 {
			delta := resp.Choices[0].Delta.Content
			if delta != "" {
				totalTokens++
				// 发送普通消息事件
				msg := model.ChatMessageResponse{
					ID:     resp.ID,
					Text:   delta, // 前端会累加
					Delta:  delta,
					Role:   openai.ChatMessageRoleAssistant,
					Detail: resp,
				}
				sendSSEEvent(w, "message", msg)
			}

			// 检查是否结束
			if resp.Choices[0].FinishReason == "stop" {
				sendSSEEvent(w, "done", map[string]interface{}{"status": "complete", "totalTokens": totalTokens})
				fmt.Printf("[Handler] 流结束（stop），总耗时: %v\n", time.Since(startTime))
				return false
			}
		}

		return true
	})
}

// sendSSEEvent 发送 SSE 事件
func sendSSEEvent(w io.Writer, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	// 立即刷新缓冲区
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// sendSSEEventQuiet 静默发送 SSE 事件（忽略写入错误）
func sendSSEEventQuiet(w io.Writer, event string, data interface{}) {
	defer func() {
		recover() // 忽略 panic
	}()
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// Config 返回配置信息
func (h *ChatHandler) Config(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"model": "kimi-k2.6",
		},
	})
}
