package main

import (
	"fmt"
	"log"

	"new_gin/config"
	"new_gin/internal/handler"
	"new_gin/pkg/openai"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建 OpenAI 客户端
	client := openai.NewClient(cfg)

	// 3. 创建聊天处理器
	chatHandler := handler.NewChatHandler(client)

	// 4. 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 5. 创建 Gin 引擎
	r := gin.Default()

	// 6. 注册路由
	api := r.Group("/api")
	{
		// 流式聊天接口
		api.POST("/chat-process", chatHandler.ChatProcess)
		// 配置接口
		api.POST("/config", chatHandler.Config)
	}

	// 7. 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("服务器启动，监听地址: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
