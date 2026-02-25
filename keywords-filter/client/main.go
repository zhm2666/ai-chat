package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"keywords-filter/proto"
	"log"
)

var addr = "localhost:50053"

func main() {
	callSensitive()
	callKeyword()
}
func callSensitive() {
	addr := "localhost:50053"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	client := proto.NewFilterClient(conn)
	ctx := AppendBearerTokenToContext(context.Background(), "22sapguumyqcang1chubdev1ozhome256487dialoud")
	in := &proto.FilterReq{
		Text: "词语 AV PK PX 中的字符替换成指定的字符，这里的字符指的是rune字符，比如*就是'*'",
	}
	fmt.Println(client.Validate(ctx, in))
	fmt.Println(client.FindAll(ctx, in))
}
func callKeyword() {
	addr := "localhost:50054"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	client := proto.NewFilterClient(conn)
	ctx := AppendBearerTokenToContext(context.Background(), "ev1ozhome256487dialoud22sapguumyqcang1chubd")
	in := &proto.FilterReq{
		Text: "词语 golang defer recover docker 中的字符替换成指定的字符，这里的字符指的是rune字符，比如*就是'*'",
	}
	fmt.Println(client.Validate(ctx, in))
	fmt.Println(client.FindAll(ctx, in))
}
func AppendBearerTokenToContext(ctx context.Context, accessToken string) context.Context {
	md := metadata.Pairs("Authorization", "Bearer "+accessToken)
	return metadata.NewOutgoingContext(ctx, md)
}
