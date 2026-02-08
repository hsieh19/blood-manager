#!/bin/sh

# 确保数据目录存在
mkdir -p /app/data

# 确保日志文件存在
touch /app/app.log

# 关键步骤：修复挂载卷后的权限问题
# 如果用户使用 -v 挂载了本地目录，挂载后的文件夹权限可能属于 root
# 我们在运行时强制将其所有权转交给 appuser
echo "Initialising: Ensuring appuser has ownership of /app/data and /app/app.log"
chown -R appuser:appgroup /app/data /app/app.log

# 切换到非 root 用户并启动程序
# su-exec 类似于 sudo，但更轻量，专门用于 Docker
echo "Starting application as appuser..."
exec su-exec appuser:appgroup ./health-manager
