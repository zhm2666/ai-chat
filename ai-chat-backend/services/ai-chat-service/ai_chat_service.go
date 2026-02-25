package ai_chat_service

import (
	"ai-chat-backend/pkg/config"
	grpc_client "ai-chat-backend/services/grpc-client"
	"sync"
)

var pool grpc_client.ClientPool
var once sync.Once

type Client struct {
	grpc_client.DefaultClient
}

func GetAiChatServiceClientPool() grpc_client.ClientPool {
	once.Do(func() {
		cnf := config.GetConfig()
		c := &Client{}
		pool = c.GetPool(cnf.DependOn.AiChatService.Address)
	})
	return pool
}
