@echo off
echo ========================================
echo   Jable TV Downloader - Build Script
echo ========================================
echo.

echo 正在編譯程式...
go build -ldflags="-s -w" -o jable-downloader.exe .\cmd\jable-downloader

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo   編譯成功！
    echo ========================================
    echo.
    echo 執行檔位置: jable-downloader.exe
    echo.
    echo 使用方式:
    echo   1. 互動模式:           jable-downloader.exe
    echo   2. 指定網址:           jable-downloader.exe --url [網址]
    echo   3. 隨機下載:           jable-downloader.exe --random
    echo   4. 批次下載:           jable-downloader.exe --all-urls [演員頁面]
    echo.
    pause
) else (
    echo.
    echo ========================================
    echo   編譯失敗！
    echo ========================================
    echo.
    echo 請檢查錯誤訊息並確保已安裝 Go 環境
    pause
    exit /b 1
)
