#!/bin/bash

echo "========================================"
echo "  Jable TV Downloader - Build Script"
echo "========================================"
echo ""

echo "正在編譯程式..."
go build -ldflags="-s -w" -o jable-downloader ./cmd/jable-downloader

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "  編譯成功！"
    echo "========================================"
    echo ""
    echo "執行檔位置: jable-downloader"
    echo ""
    echo "使用方式:"
    echo "  1. 互動模式:           ./jable-downloader"
    echo "  2. 指定網址:           ./jable-downloader --url [網址]"
    echo "  3. 隨機下載:           ./jable-downloader --random"
    echo "  4. 批次下載:           ./jable-downloader --all-urls [演員頁面]"
    echo ""
    chmod +x jable-downloader
else
    echo ""
    echo "========================================"
    echo "  編譯失敗！"
    echo "========================================"
    echo ""
    echo "請檢查錯誤訊息並確保已安裝 Go 環境"
    exit 1
fi
