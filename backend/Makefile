# TalkAboutIt Backend Makefile
# 提供编译、测试、集成测试和部署相关命令。

# 变量定义
BINARY_NAME := talkaboutit
BUILD_DIR := bin
GO := go
GOFLAGS := -ldflags="-s -w"

# 默认目标
.PHONY: all
all: build

# 编译二进制到 bin/talkaboutit
.PHONY: build
build:
	@echo "==> 编译 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "==> 编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行所有单元测试
.PHONY: test
test:
	@echo "==> 运行单元测试..."
	$(GO) test ./... -v -count=1

# 运行集成测试（需要 build tag: integration）
.PHONY: integration
integration:
	@echo "==> 运行集成测试..."
	$(GO) test -tags=integration ./test/ -v -count=1

# 运行评测（需要 build tag: eval）
.PHONY: eval
eval:
	@echo "==> 运行评测..."
	$(GO) test -tags=eval ./test/eval/ -v -count=1

# 清理构建产物
.PHONY: clean
clean:
	@echo "==> 清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@rm -f test/eval/baseline.json

# 格式化代码
.PHONY: fmt
fmt:
	@echo "==> 格式化代码..."
	$(GO) fmt ./...

# 静态检查
.PHONY: vet
vet:
	@echo "==> 静态检查..."
	$(GO) vet ./...

# 下载依赖
.PHONY: deps
deps:
	@echo "==> 下载依赖..."
	$(GO) mod download

# 帮助信息
.PHONY: help
help:
	@echo "TalkAboutIt Backend Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make build       编译二进制到 bin/talkaboutit"
	@echo "  make test        运行所有单元测试 (go test ./...)"
	@echo "  make integration 运行集成测试 (go test -tags=integration ./test/)"
	@echo "  make eval        运行评测框架 (go test -tags=eval ./test/eval/)"
	@echo "  make clean       清理构建产物"
	@echo "  make fmt         格式化 Go 代码"
	@echo "  make vet         运行 go vet"
	@echo "  make deps        下载依赖"
	@echo "  make help        显示本帮助信息"
