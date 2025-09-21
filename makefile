# Makefile
.PHONY: build test lint clean install

BINARY_NAME=genGormStructV2
VERSION=2.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_HASH=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitHash=$(GIT_HASH)"
X_CMD=.

build:
	@echo "构建 $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ${X_CMD_DIR}

build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ${X_CMD_DIR}

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ${X_CMD_DIR}

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ${X_CMD_DIR}

test:
	@echo "运行测试..."
	go test -v -cover ./...

lint:
	@echo "代码检查..."
	golangci-lint run

install:
	go install ${X_CMD_DIR}

clean:
	@echo "清理构建文件..."
	rm -rf bin/ output/ coverage.out

run:
	go run ${X_CMD_DIR}/main.go

help:
	@echo "可用命令:"
	@echo "  build      - 构建二进制文件"
	@echo "  build-all  - 构建多平台二进制文件"
	@echo "  test       - 运行测试"
	@echo "  lint       - 代码检查"
	@echo "  install    - 安装到 GOPATH"
	@echo "  clean      - 清理文件"
	@echo "  run        - 运行脚本"
