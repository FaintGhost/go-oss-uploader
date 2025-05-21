# 使用多阶段构建
# 第一阶段：用于构建Go应用程序
FROM golang:1.24-alpine AS builder

# 安装必要的依赖项
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY ./backend/go.mod ./backend/go.sum ./

# 下载所有依赖项
RUN go mod download

# 复制源代码
COPY ./backend/ ./

# 构建应用程序
RUN go build -o uploader .

# 第二阶段：创建最终镜像
FROM alpine:latest

# 安装必要的系统依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 从builder阶段复制编译好的应用
COPY --from=builder /app/uploader .
COPY --from=builder /app/test.html .

# 创建必要的目录
RUN mkdir -p /app/temp /app/checkpoint /app/logs

# 暴露应用端口
EXPOSE 5050

# 设置容器启动命令
CMD ["./uploader", "--verbose"]