# ===================================
# 血压管理系统 - GitHub Workflows 专用 Dockerfile
# 多阶段构建，使用 Alpine 最小镜像
# ===================================

# =============
# 第一阶段：编译
# =============
FROM golang:1.21-alpine AS builder

# 安装编译依赖 (CGO 需要 gcc)
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译为静态二进制文件
# CGO_ENABLED=1 是因为使用了 SQLite (go-sqlite3 等驱动通常需要 CGO)
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o blood-manager main.go

# =============
# 第二阶段：运行
# =============
FROM alpine:3.19

# 安装运行时必要依赖 (ca-certificates, tzdata)
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 创建并进入工作目录
WORKDIR /app

# 从构建阶段复制二进制文件到 /app
COPY --from=builder /build/blood-manager .

# 复制静态资源文件夹到 /app/static
COPY --from=builder /build/static ./static

# 创建数据存储目录，确保程序有权读写
RUN mkdir -p /app/data /app/config

# 给予二进制文件执行权限
RUN chmod +x /app/blood-manager

# 设置生产环境环境变量
ENV GIN_MODE=release

# 暴露 Web 服务端口
EXPOSE 8080

# 随容器启动运行
CMD ["./blood-manager"]
