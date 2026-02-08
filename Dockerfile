# ===================================
# 健康管理系统 - GitHub Workflows 专用 Dockerfile
# 多阶段构建，使用 Alpine 最小镜像
# ===================================

# =============
# 第一阶段：编译
# =============
FROM golang:1.25-alpine AS builder

# 安装编译依赖 (CGO 需要 gcc)
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# 复制所有源代码
COPY . .

# 更新依赖并下载
RUN go mod tidy && go mod download

# 编译为静态二进制文件
# CGO_ENABLED=1 是因为使用了 SQLite (go-sqlite3 等驱动通常需要 CGO)
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o health-manager cmd/health-manager/main.go

# =============
# 第二阶段：运行
# =============
FROM alpine:3.21

# 安装运行时必要依赖 (ca-certificates, tzdata, su-exec)
RUN apk update && apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata su-exec \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 创建并进入工作目录
WORKDIR /app

# 从构建阶段复制二进制文件到 /app
COPY --from=builder /build/health-manager .

# 复制静态资源文件夹到 /app/web/static
COPY --from=builder /build/web/static ./web/static

# 复制入口脚本
COPY entrypoint.sh ./entrypoint.sh

# 给予执行权限
RUN chmod +x /app/health-manager /app/entrypoint.sh

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置生产环境环境变量
ENV GIN_MODE=release

# 暴露 Web 服务端口
EXPOSE 8080

# 使用入口脚本启动，它会处理权限并切换用户
ENTRYPOINT ["./entrypoint.sh"]
