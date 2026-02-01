# URL Shortener Service Makefile

.PHONY: build run test clean install-deps

# 默认目标
help:
	@echo "可用命令:"
	@echo "  make build     - 构建项目"
	@echo "  make run       - 运行服务"
	@echo "  make install-deps - 安装依赖"
	@echo "  make clean     - 清理构建文件"

# 安装依赖
install-deps:
	go mod tidy

# 构建项目
build:
	go build -o bin/url-shortener cmd/server/main.go

# 运行服务
run: install-deps
	go run cmd/server/main.go

# 构建并运行
build-run: build
	./bin/url-shortener

# 清理
clean:
	rm -rf bin/
	rm -f urls.db

# 简单测试
test:
	@echo "运行简单功能测试..."
	go run test_example.go