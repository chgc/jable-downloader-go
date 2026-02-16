@echo off
chcp 65001 >nul
echo ==========================================
echo 測試清除已完成任務功能
echo ==========================================
echo.

echo 1. 查看當前任務列表
curl -s http://localhost:18080/api/tasks | jq .
echo.
echo.

echo 2. 清除已完成的任務 (使用 DELETE 方法)
curl -s -X DELETE http://localhost:18080/api/tasks/clear-completed | jq .
echo.
echo.

echo 3. 再次查看任務列表 (驗證已清除)
curl -s http://localhost:18080/api/tasks | jq .
echo.
echo.

echo ==========================================
echo 測試完成！
echo ==========================================
