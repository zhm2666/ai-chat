package main

import (
	"ai-chat-backend/pkg/config"
	"ai-chat-backend/pkg/log"
	"ai-chat-backend/pkg/middlewares"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"ai-chat-backend/pkg/controllers"
	"github.com/gin-gonic/gin"
)

type ChatGPTWebServer struct {
	config *config.Config
	log    log.ILogger
}

func (r *ChatGPTWebServer) Run(ctx context.Context) error {
	go r.httpServer(ctx)
	<-ctx.Done()
	return nil
}

func (r *ChatGPTWebServer) httpServer(ctx context.Context) {
	chatService, err := controllers.NewChatService(r.config, r.log)
	if err != nil {
		r.log.Fatal(err)
	}
	addr := fmt.Sprintf("%s:%d", r.config.Http.IP, r.config.Http.Port)
	r.log.InfoF("ChatGPT Web Server on: %s", addr)
	server := &http.Server{
		Addr: addr,
	}
	entry := gin.Default()
	entry.Use(middlewares.Cors())

	chat := entry.Group("/api")
	//如需权限校验，请取消注释
	//chat.Use(middlewares.Auth())
	chat.POST("/chat-process", chatService.ChatProcess)
	chat.POST("/config", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "Success",
			"data": map[string]string{
				"apiModel":   "ChatGPTAPI",
				"socksProxy": "",
			},
		})
	})
	chat.POST("/session", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status":  "Success",
			"message": "",
			"data": gin.H{
				"auth": false,
			},
		})
	})
	chat.GET("/health", func(c *gin.Context) {})

	fs := http.FileServer(http.Dir("www"))
	entry.NoRoute(func(ctx *gin.Context) {
		fs.ServeHTTP(ctx.Writer, ctx.Request)
	})
	entry.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "www/index.html")
	})

	server.Handler = entry
	go func(ctx context.Context) {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			r.log.InfoF("Server shutdown with error %v", err)
		}
	}(ctx)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		r.log.FatalF("Server listen and serve error %v", err)
	}
}

var configFile = flag.String("config", "dev.config.yaml", "")

func main() {
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()
	//初始化配置文件
	config.InitConfig(*configFile)
	cnf := config.GetConfig()

	log.SetLevel(cnf.Log.Level)
	log.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	log.SetPrintCaller(true)

	logger := log.NewLogger()
	logger.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	logger.SetLevel(cnf.Log.Level)
	logger.SetPrintCaller(true)

	webServer := ChatGPTWebServer{
		config: cnf,
		log:    logger,
	}
	webServer.Run(ctx)
	<-ctx.Done()
}
