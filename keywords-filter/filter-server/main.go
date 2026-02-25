package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"keywords-filter/filter-server/interceptor"
	"keywords-filter/filter-server/server"
	"keywords-filter/pkg/config"
	"keywords-filter/pkg/filter"
	"keywords-filter/pkg/log"
	"keywords-filter/proto"
	"net"
)

var (
	configFile   = flag.String("config", "dev.config.yaml", "")
	dictFilePath = flag.String("dict", "dict.txt", "")
	formatDict   = flag.Bool("format", false, "")
)

func main() {
	flag.Parse()
	if *formatDict {
		filter.OverwriteDict(*dictFilePath)
		return
	}

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

	filter.InitFilter(*dictFilePath)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cnf.Server.IP, cnf.Server.Port))
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryAuthInterceptor), grpc.StreamInterceptor(interceptor.StreamAuthInterceptor))
	service := server.NewFilterService(filter.GetFilter())
	proto.RegisterFilterServer(s, service)

	healthCheckSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthCheckSrv)

	if err = s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
