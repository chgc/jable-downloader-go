#!/bin/bash

echo "==============================================="
echo "  Jable Downloader - API Server Launcher"
echo "==============================================="
echo ""

# 检查可执行文件是否存在
if [ ! -f "./jable-downloader" ]; then
    echo "[ERROR] jable-downloader not found!"
    echo "Please build the project first:"
    echo "  go build -o jable-downloader ./cmd/jable-downloader"
    exit 1
fi

# 默认端口
PORT=${1:-18080}

echo "[INFO] Starting API server on port $PORT..."
echo "[INFO] Chrome extension can now connect to:"
echo "       http://localhost:$PORT"
echo ""
echo "[INFO] Press Ctrl+C to stop the server"
echo "==============================================="
echo ""

./jable-downloader --server --port "$PORT"
