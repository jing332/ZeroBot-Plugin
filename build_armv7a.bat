set GOARCH=arm
set GOOS=linux
set CGO_ENABLED=0

go version
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy

go build -ldflags="-s -w" -o zbp_arm