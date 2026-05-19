package openai

import (
	"context"
	"fmt"
	"time"

	"new_gin/config"

	openai "github.com/sashabaranov/go-openai"
)

// Client OpenAI 客户端封装
type Client struct {
	client *openai.Client
	config *config.Config
}

// NewClient 创建新的 OpenAI 客户端
func NewClient(cfg *config.Config) *Client {
	openaiConfig := openai.DefaultConfig(cfg.OpenAI.APIKey)
	openaiConfig.BaseURL = cfg.OpenAI.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	return &Client{
		client: client,
		config: cfg,
	}
}

// CreateChatCompletionStream 创建流式聊天请求
func (c *Client) CreateChatCompletionStream(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionStream, error) {
	startTime := time.Now()
	fmt.Printf("[OpenAI] 开始请求，消息数量: %d\n", len(messages))

	req := openai.ChatCompletionRequest{
		Model:            c.config.OpenAI.Model,
		Messages:         messages,
		MaxTokens:        c.config.OpenAI.MaxTokens,
		Temperature:      c.config.OpenAI.Temperature,
		TopP:             c.config.OpenAI.TopP,
		PresencePenalty:  c.config.OpenAI.PresencePenalty,
		FrequencyPenalty: c.config.OpenAI.FrequencyPenalty,
		Stream:           true,
		ChatTemplateKwargs: map[string]any{
			"thinking": map[string]any{
				"type": "disabled",
			},
		},
	}

	fmt.Printf("[OpenAI] 请求参数: model=%s, max_tokens=%d, temperature=%.2f, top_p=%.2f\n",
		req.Model, req.MaxTokens, req.Temperature, req.TopP)

	// 使用传入的 context，支持请求取消
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("[OpenAI] 请求失败，耗时: %v, 错误: %v\n", time.Since(startTime), err)
		return nil, err
	}

	fmt.Printf("[OpenAI] 连接建立成功，耗时: %v\n", time.Since(startTime))
	return stream, nil
}

// ChatMessage OpenAI 聊天消息
type ChatMessage = openai.ChatCompletionMessage

// ChatMessageRoleSystem 系统角色
var ChatMessageRoleSystem = openai.ChatMessageRoleSystem

// ChatMessageRoleUser 用户角色
var ChatMessageRoleUser = openai.ChatMessageRoleUser

// ChatMessageRoleAssistant 助手角色
var ChatMessageRoleAssistant = openai.ChatMessageRoleAssistant

// ChatCompletionStreamResponse 流式响应
type ChatCompletionStreamResponse = openai.ChatCompletionStreamResponse
