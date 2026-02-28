APP_NAME    := sora2api
SERVER_BIN  := bin/$(APP_NAME)-server
CLI_BIN     := bin/$(APP_NAME)
MODULE      := github.com/DouDOU-start/go-sora2api

# 版本信息
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE        := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go 参数
GOFLAGS     := -trimpath
VERSION_PKG := main
LDFLAGS     := -s -w -X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).commit=$(COMMIT) -X $(VERSION_PKG).date=$(DATE)

# Release 平台
PLATFORMS   := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: all install build server cli web-build run dev dev-server dev-web clean clean-all fmt lint lint-web vet tidy release help

## —— 常用 ——————————————————————————————————

all: tidy lint build  ## 完整流程：tidy + lint + build

help:  ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

## —— 安装 ——————————————————————————————————

install:  ## 安装依赖（Go + 前端 npm）
	@echo "==> 安装 Go 依赖..."
	go mod download
	@echo "==> 安装前端依赖..."
	cd web && npm install
	@echo "==> 依赖安装完成"

## —— 构建 ——————————————————————————————————

build: server cli  ## 构建全部（server + cli）

server: web-build  ## 构建 server（含前端）
	@echo "==> 构建 server..."
	@mkdir -p bin
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(SERVER_BIN) ./server/

cli:  ## 构建 CLI 工具
	@echo "==> 构建 CLI..."
	@mkdir -p bin
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(CLI_BIN) ./cmd/sora2api/

web-build:  ## 构建前端并同步到 server/dist
	@echo "==> 构建前端..."
	cd web && npm install --silent && npm run build
	@rm -rf server/dist
	@cp -r web/dist server/dist
	@echo "==> 前端已同步到 server/dist"

## —— 运行 ——————————————————————————————————

run: server  ## 一键启动 server
	@echo "==> 启动 server..."
	./$(SERVER_BIN)

dev:  ## 开发模式（同时启动后端 + 前端，访问 http://localhost:5173）
	@mkdir -p server/dist && touch server/dist/index.html
	@echo "==> 启动后端 (8686) + 前端 (5173)..."
	@echo "==> 请访问 http://localhost:5173"
	@trap 'kill 0' INT TERM; \
		go run ./server/ & \
		cd web && npm run dev & \
		wait

dev-server:  ## 仅启动后端（不含前端）
	@mkdir -p server/dist && touch server/dist/index.html
	go run ./server/

dev-web:  ## 仅启动前端（Vite dev server）
	cd web && npm run dev

## —— 代码质量 ——————————————————————————————

fmt:  ## 格式化代码
	@echo "==> gofmt..."
	gofmt -s -w .
	@echo "==> goimports..."
	@command -v goimports >/dev/null 2>&1 && goimports -w . || echo "  跳过 goimports（未安装: go install golang.org/x/tools/cmd/goimports@latest）"

vet:  ## 静态分析
	go vet ./...

lint: vet  ## 代码检查（vet + golangci-lint）
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run ./... \
		|| echo "==> 跳过 golangci-lint（未安装: https://golangci-lint.run/welcome/install/）"

lint-web:  ## 前端代码检查
	cd web && npm run lint

## —— 依赖 ——————————————————————————————————

tidy:  ## 整理 Go 依赖
	go mod tidy

## —— 发布 ——————————————————————————————————

release: web-build  ## 交叉编译多平台发布包
	@echo "==> 构建发布包 $(VERSION)..."
	@rm -rf release/
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		output="release/$(APP_NAME)-server_$${os}_$${arch}"; \
		echo "  构建 $${os}/$${arch}..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $$output ./server/; \
	done
	@echo "==> 打包..."
	@cd release && for f in $(APP_NAME)-server_*; do \
		name=$${f%.*}; \
		ver=$$(echo $(VERSION) | sed 's/^v//'); \
		tarname="$(APP_NAME)_$${ver}_$${f#$(APP_NAME)-server_}"; \
		cp $$f $(APP_NAME)-server; \
		tar czf "$${tarname}.tar.gz" $(APP_NAME)-server; \
		rm $(APP_NAME)-server; \
	done
	@echo "==> 生成 checksums..."
	@cd release && sha256sum *.tar.gz > checksums.txt
	@echo "==> 发布包就绪: release/"

## —— 清理 ——————————————————————————————————

clean:  ## 清理 Go 构建产物
	rm -rf bin/

clean-all: clean  ## 清理全部（含前端构建产物）
	rm -rf server/dist/
	rm -rf web/dist/
	rm -rf release/
