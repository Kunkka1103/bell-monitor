# Makefile

# 定义变量
APP_NAME = bell-monitor
SRC = main.go
GO = go

# 默认目标
.PHONY: all
all: build

# 编译目标
.PHONY: build
build:
	$(GO) build -o $(APP_NAME) $(SRC)

# 清理目标
.PHONY: clean
clean:
	rm -f $(APP_NAME)

# 运行目标
.PHONY: run
run: build
	./$(APP_NAME)

# 重新编译目标
.PHONY: rebuild
rebuild: clean build

