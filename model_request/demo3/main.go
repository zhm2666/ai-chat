package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultConfig("sk-sE9IjV656TK8ql6jGNjjOWN07fKCpml3RQulGSSLH4qTS0Ue")
	config.BaseURL = "https://api.moonshot.cn/v1"

	client := openai.NewClientWithConfig(config)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你是 Kimi，由 Moonshot AI 提供的人工智能助手，你更擅长中文和英文的对话。你会为用户提供安全，有帮助，准确的回答。同时，你会拒绝一切涉及恐怖主义，种族歧视，黄色暴力等问题的回答。Moonshot AI 为专有名词，不可翻译成其他语言。",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "你好，我叫李雷，1+1等于多少？",
		},
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       "kimi-k2-turbo-preview",
			Messages:    messages,
			Temperature: 0.6,
		},
	)

	if err != nil {
		log.Fatalf("ChatCompletion error: %v", err)
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
