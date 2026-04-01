# ============================================
# Spot VM 自动管理服务 - Makefile
# ============================================

# 变量定义
APP_NAME     := spot-manager
MODULE       := gitee.com/dinglide/spot-vm
MAIN_PKG     := ./cmd/spot-manager
BUILD_DIR    := build
GO           := go
GOFLAGS      := -v

# 版本信息（从 git 获取）
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT       := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME   := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# 编译标志
LDFLAGS      := -s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildTime=$(BUILD_TIME)

# 目标平台（交叉编译用）
GOOS         ?= $(shell go env GOOS)
GOARCH       ?= $(shell go env GOARCH)

# ============================================
# 默认目标
# ============================================
.PHONY: all
all: tidy build

# ============================================
# 开发相关
# ============================================

## build: 编译二进制文件
.PHONY: build
build:
	@echo "🔨 编译 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)
	@echo "✅ 编译完成: $(BUILD_DIR)/$(APP_NAME)"

## build-linux: 交叉编译 Linux amd64 版本（用于部署到云服务器）
.PHONY: build-linux
build-linux:
	@echo "🔨 交叉编译 Linux amd64 版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PKG)
	@echo "✅ 编译完成: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

## run: 编译并运行
.PHONY: run
run: build
	@echo "🚀 启动 $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

## dev: 直接运行（不编译，适合开发调试）
.PHONY: dev
dev:
	$(GO) run $(MAIN_PKG)

# ============================================
# 代码质量
# ============================================

## tidy: 整理依赖
.PHONY: tidy
tidy:
	@echo "📦 整理依赖..."
	$(GO) mod tidy

## fmt: 格式化代码
.PHONY: fmt
fmt:
	@echo "🎨 格式化代码..."
	$(GO) fmt ./...

## vet: 静态分析
.PHONY: vet
vet:
	@echo "🔍 静态分析..."
	$(GO) vet ./...

## lint: 代码检查（需要安装 golangci-lint）
.PHONY: lint
lint:
	@echo "🧹 代码检查..."
	@which golangci-lint > /dev/null 2>&1 || { echo "⚠️  请先安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run ./...

## check: 运行所有代码质量检查
.PHONY: check
check: fmt vet lint

# ============================================
# 测试
# ============================================

## test: 运行测试
.PHONY: test
test:
	@echo "🧪 运行测试..."
	$(GO) test ./... -v

## test-cover: 运行测试并生成覆盖率报告
.PHONY: test-cover
test-cover:
	@echo "🧪 运行测试（含覆盖率）..."
	@mkdir -p $(BUILD_DIR)
	$(GO) test ./... -coverprofile=$(BUILD_DIR)/coverage.out
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "✅ 覆盖率报告: $(BUILD_DIR)/coverage.html"

# ============================================
# 清理
# ============================================

## clean: 清理编译产物
.PHONY: clean
clean:
	@echo "🧹 清理编译产物..."
	rm -rf $(BUILD_DIR)
	@echo "✅ 清理完成"

# ============================================
# 部署辅助
# ============================================

## deploy: 编译 Linux 版本并通过 SCP 部署到远程服务器
## 用法: make deploy HOST=1.2.3.4 [USER=root] [REMOTE_DIR=/opt/spot-manager]
HOST       ?=
USER       ?= root
REMOTE_DIR ?= /opt/spot-manager

.PHONY: deploy
deploy: build-linux
ifndef HOST
	$(error ❌ 请指定 HOST，用法: make deploy HOST=1.2.3.4)
endif
	@echo "🚀 部署到 $(USER)@$(HOST):$(REMOTE_DIR)..."
	ssh $(USER)@$(HOST) "mkdir -p $(REMOTE_DIR)"
	scp $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(USER)@$(HOST):$(REMOTE_DIR)/$(APP_NAME)
	@test -f .env && scp .env $(USER)@$(HOST):$(REMOTE_DIR)/.env || echo "⚠️  .env 文件不存在，跳过"
	ssh $(USER)@$(HOST) "chmod +x $(REMOTE_DIR)/$(APP_NAME)"
	@echo "✅ 部署完成"
	@echo "💡 在远程服务器上运行: cd $(REMOTE_DIR) && nohup ./$(APP_NAME) > /var/log/$(APP_NAME).log 2>&1 &"

# ============================================
# 帮助
# ============================================

## help: 显示帮助信息
.PHONY: help
help:
	@echo ""
	@echo "$(APP_NAME) - Spot VM 自动管理服务"
	@echo ""
	@echo "用法: make [目标]"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | sort
	@echo ""
