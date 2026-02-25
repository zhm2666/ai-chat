# keywords-filter

## grpc stub 生成
``` 
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/filter.proto
```
## 依赖对齐
``` 
go mod tidy
```

## 性能测试执行命令
``` 
go test --bench . .\filter_test\
go test --bench . -cpu 1  .\filter_test\
go test --bench . --run ^$  -cpu 1  .\filter_test\
go test --bench FindAll  --run ^$  -cpu 1 -outputdir ./ -cpuprofile cpu.out  .\filter_test\
```

## 性能分析工具使用
``` 
go tool pprof -web cpu.out
go tool pprof -http=:9090 cpu.out
```

## 启动脱敏服务
``` 
go run .\filter-server\main.go --config=dev.config.yaml --dict sensitive-dict.txt
```