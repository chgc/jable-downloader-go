@echo off
echo ===============================================
echo   Jable Downloader - Docker Setup
echo ===============================================
echo.

REM 检查 Docker 是否安装
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not installed!
    echo Please install Docker Desktop: https://docs.docker.com/desktop/windows/install/
    pause
    exit /b 1
)

REM 检查 Docker Compose 是否可用
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker Compose is not available!
    pause
    exit /b 1
)

echo [INFO] Building Docker image...
docker-compose build

if %errorlevel% neq 0 (
    echo [ERROR] Docker build failed!
    pause
    exit /b 1
)

echo.
echo [INFO] Starting container...
docker-compose up -d

if %errorlevel% neq 0 (
    echo [ERROR] Failed to start container!
    pause
    exit /b 1
)

echo.
echo [SUCCESS] Jable Downloader is now running!
echo.
echo API Server: http://localhost:18080
echo Health Check: http://localhost:18080/api/health
echo.
echo Commands:
echo   View logs:    docker-compose logs -f
echo   Stop server:  docker-compose down
echo   Restart:      docker-compose restart
echo.
echo Downloads will be saved to: .\download
echo ===============================================
echo.
pause
