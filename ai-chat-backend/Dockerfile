FROM quay.io/0voice/golang:1.20 as stage1
RUN go env -w GOPROXY=https://goproxy.cn,https://proxy.golang.com.cn,direct
ADD ./ /src/ai-chat-backend
WORKDIR /src/ai-chat-backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ai-chat-backend cmd/*.go

FROM quay.io/0voice/alpine:3.18 as stage2
ADD ./curl-amd64 /usr/bin/curl
RUN chmod +x /usr/bin/curl

MAINTAINER nick
WORKDIR /app/
ADD ./dev.config.yaml /app/config.yaml
COPY ./www /app/www
COPY --from=stage1 /src/ai-chat-backend/ai-chat-backend /app/ai-chat-backend
# 指定入口程序
ENTRYPOINT ["./ai-chat-backend"]
# 指定容器的启动命令或者入口程序的参数
CMD ["--config=config.yaml"]