# TalkAboutIt Backend Dockerfile
# 多阶段构建：编译阶段 → 运行阶段

# ==================== 编译阶段 ====================
FROM golang:1.23-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 先复制 go.mod 和 go.sum，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译静态二进制（禁用 CGO，避免依赖 glibc）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o bin/talkaboutit ./cmd/server

# ==================== 运行阶段 ====================
FROM alpine:3.20

# 安装运行时依赖（ca-certificates 用于 HTTPS 请求）
RUN apk add --no-cache ca-certificates

WORKDIR /app

# 从编译阶段复制二进制
COPY --from=builder /app/bin/talkaboutit /app/talkaboutit

# 复制 personas 目录和数据目录结构
COPY --from=builder /app/personas /app/personas
RUN mkdir -p /app/data

# 暴露服务端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/personas || exit 1

# 运行服务
ENTRYPOINT ["/app/talkaboutit"]
