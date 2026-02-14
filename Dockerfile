# 多階段構建：構建和運行環境分離
FROM golang:1.24-alpine AS builder

# 安裝構建依賴
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# 複製 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jable-downloader ./cmd/jable-downloader

# 運行階段：使用更小的基礎鏡像
FROM alpine:latest

# 安裝運行依賴
RUN apk add --no-cache \
    ffmpeg \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont

# 創建非 root 用戶
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# 從構建階段複製執行檔
COPY --from=builder /app/jable-downloader .

# 創建下載目錄
RUN mkdir -p /app/download && chown -R appuser:appgroup /app

# 切換到非 root 用戶
USER appuser

# 設置環境變量
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    DOCKER_CONTAINER=true

# 暴露 API 端口
EXPOSE 18080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:18080/api/health || exit 1

# 啟動服務
CMD ["./jable-downloader", "--server", "--port", "18080"]
