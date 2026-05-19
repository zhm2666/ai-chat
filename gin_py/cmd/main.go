package main

import (
	"gin_py/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 初始化处理器
	chatHandler := handler.NewChatHandler()

	// 注册路由
	r.POST("/api/chat-process", chatHandler.ChatProcess)
	// 如果有配置接口，也可以注册
	// r.GET("/api/config", chatHandler.Config)

	// 启动服务
	if err := r.Run(":7080"); err != nil {
		panic(err)
	}
}
