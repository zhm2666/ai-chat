
go run .\filter-server\main.go --config=dev.config.yaml --dict .\sensitive-dict.txt

go run .\filter-server\main.go --config=dev.config.yaml --dict sensitive-dict.txt

go run .\cmd\main.go

go run .\services\services.go

go run .\cmd\main.go --config=config.yaml

