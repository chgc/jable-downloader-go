#!/bin/bash

echo "==============================================="
echo "  Jable Downloader - Docker Setup"
echo "==============================================="
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "[ERROR] Docker is not installed!"
    echo "Please install Docker first: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "[ERROR] Docker Compose is not installed!"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

echo "[INFO] Building Docker image..."
docker-compose build

if [ $? -ne 0 ]; then
    echo "[ERROR] Docker build failed!"
    exit 1
fi

echo ""
echo "[INFO] Starting container..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "[ERROR] Failed to start container!"
    exit 1
fi

echo ""
echo "[SUCCESS] Jable Downloader is now running!"
echo ""
echo "API Server: http://localhost:18080"
echo "Health Check: http://localhost:18080/api/health"
echo ""
echo "Commands:"
echo "  View logs:    docker-compose logs -f"
echo "  Stop server:  docker-compose down"
echo "  Restart:      docker-compose restart"
echo ""
echo "Downloads will be saved to: ./download"
echo "==============================================="
