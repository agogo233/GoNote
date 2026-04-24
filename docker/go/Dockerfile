# ==================== 构建阶段 ====================
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
# Install dependencies
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum 文件，利用 Docker 缓存加速构建
# Copy go mod files
COPY go/go.mod go/go.sum ./
RUN go mod download

# 复制源代码
# Copy source code
COPY go/ .

# 编译应用（优化编译参数减小二进制体积）
# - CGO_ENABLED=0: 禁用 CGO，生成静态链接的二进制文件
# - -ldflags="-s -w": 去除调试信息和符号表
# - -trimpath: 去除文件路径信息
# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o gonote ./go/cmd/server

# ==================== 运行阶段 ====================
# Runtime stage
FROM alpine:latest

WORKDIR /app

# 安装运行时依赖
# - ca-certificates: HTTPS 证书
# - tzdata: 时区数据
# - wget: 健康检查
# Install ca-certificates for HTTPS and wget for healthcheck
RUN apk --no-cache add ca-certificates tzdata wget

# 从构建阶段复制二进制文件
# Copy binary from builder
COPY --from=builder /app/gonote .

# 复制前端和静态资源文件
# Copy frontend and other static files
COPY shared/frontend ./frontend
COPY shared/themes ./themes
COPY shared/locales ./locales
COPY go/VERSION ./VERSION
COPY go/config.yaml ./config.yaml

# 创建数据目录
# Create data directory
RUN mkdir -p /app/data

# 设置环境变量
# Set environment variables
ENV PORT=9000

# 暴露端口
# Expose port
EXPOSE 9000

# 运行应用
# Run the application
CMD ["./gonote"]