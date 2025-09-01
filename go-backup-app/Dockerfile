# --- Builder Stage ---
# 使用包含 Go 和 Node.js 的镜像
FROM golang:1.21-alpine AS builder

# 安装 Wails 依赖
RUN apk add --no-cache build-base gcc musl-dev nodejs npm upx

WORKDIR /app

# 缓存 Go 和 Node 模块
COPY go.mod go.sum ./
RUN go mod download
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm install

# 拷贝所有源代码
COPY . .

# 构建应用 (针对 linux amd64)
# -ldflags="-w -s" 减小体积
# -nsis: 不生成 Windows 安装程序
# CGO_ENABLED=1 是 Wails 必需的
ENV CGO_ENABLED=1
RUN wails build -platform linux/amd64 -ldflags="-w -s" -nsis

# --- Final Stage ---
# 使用一个非常小的基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从 builder 阶段拷贝编译好的二进制文件
COPY --from=builder /app/build/bin/go-backup-app .

# 设置入口点
# 例如，可以添加命令行参数来支持非 GUI 模式
# CMD ["./go-backup-app", "backup", "--source", "/data/src", "--dest", "/data/backup/mybackup.gobackup"]
ENTRYPOINT ["./go-backup-app"]