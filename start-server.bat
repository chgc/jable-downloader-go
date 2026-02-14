@echo off
echo ===============================================
echo   Jable Downloader - API Server Launcher
echo ===============================================
echo.

REM 检查可执行文件是否存在
if not exist "jable-downloader.exe" (
    echo [ERROR] jable-downloader.exe not found!
    echo Please build the project first:
    echo   go build -o jable-downloader.exe ./cmd/jable-downloader
    pause
    exit /b 1
)

REM 默认端口
set PORT=18080
if not "%1"=="" set PORT=%1

echo [INFO] Starting API server on port %PORT%...
echo [INFO] Chrome extension can now connect to:
echo        http://localhost:%PORT%
echo.
echo [INFO] Press Ctrl+C to stop the server
echo ===============================================
echo.

jable-downloader.exe --server --port %PORT%

pause
