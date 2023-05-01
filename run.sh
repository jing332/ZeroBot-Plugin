go version
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy
go build -o ZeroBot-Plugin
#go run main.go
