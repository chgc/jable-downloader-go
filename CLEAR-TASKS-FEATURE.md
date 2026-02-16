# 清除已完成任務功能

## 📋 功能概述

新增了清除已完成任務的功能，允許用户在 Chrome 擴展和 API 端點中清理已完成或失敗的下載任務。

## ✨ 新功能

### 1. API 端點

新增了 `/api/tasks/clear-completed` 端點來清除已完成的任務。

**請求方式：**
- `DELETE /api/tasks/clear-completed` (推薦)
- `POST /api/tasks/clear-completed` (兼容)

**響應示例：**
```json
{
  "success": true,
  "message": "Successfully cleared 5 completed task(s)",
  "cleared_count": 5
}
```

**功能特點：**
- ✅ 只刪除 `completed` 和 `failed` 狀態的任務
- ✅ 保留 `queued` 和 `downloading` 狀態的任務
- ✅ 不會刪除當前正在處理的任務
- ✅ 支援 CORS，可從瀏覽器調用

### 2. Chrome 擴展集成

在 Chrome 擴展的彈出窗口中新增了清除按鈕。

**使用方式：**
1. 點擊擴展圖標打開彈出窗口
2. 查看下載隊列
3. 點擊「🗑️ 清除已完成任務」按鈕
4. 已完成和失敗的任務將被清除

**智能顯示：**
- 當沒有已完成的任務時，清除按鈕會自動隱藏
- 清除後會顯示清除數量的提示訊息
- 自動刷新任務列表

## 🔧 使用示例

### 命令行測試

使用提供的測試腳本：

```bash
# Windows
.\test-clear-tasks.bat

# 或手動測試
curl -X DELETE http://localhost:18080/api/tasks/clear-completed
```

### Chrome 擴展

1. 確保 API 服務器正在運行
2. 點擊瀏覽器工具欄的擴展圖標
3. 在彈出窗口中查看任務隊列
4. 點擊紅色的清除按鈕

## 📊 技術細節

### 後端實現

**文件：** `internal/server/server.go`

**新增函數：**
- `handleClearCompletedTasks()` - 處理清除請求
- `ClearCompletedResponse` - 響應結構

**邏輯：**
```go
// 遍歷所有任務
for taskID, task := range s.tasks {
    // 只刪除已完成或失敗的任務，且不是當前處理中的任務
    if (task.Status == "completed" || task.Status == "failed") && 
       taskID != currentTaskID {
        delete(s.tasks, taskID)
        clearedCount++
    }
}
```

### 前端實現

**文件：** `extension/popup.html`, `extension/popup.js`

**新增元素：**
- 清除按鈕（紅色主題）
- 智能顯示控制

**功能：**
```javascript
// 計算已完成的任務數
const completedCount = tasks.filter(t => 
  t.status === 'completed' || t.status === 'failed'
).length;

// 控制按鈕顯示
clearCompletedBtn.style.display = completedCount > 0 ? 'block' : 'none';

// 發送清除請求
const response = await fetch(`${currentApiUrl}/api/tasks/clear-completed`, {
  method: 'DELETE',
  signal: AbortSignal.timeout(5000)
});
```

## 🎯 使用場景

1. **定期清理**：長期使用後積累大量已完成任務，需要清理界面
2. **錯誤恢復**：批量清除失敗的任務，重新開始
3. **隱私保護**：清除下載記錄
4. **性能優化**：減少內存中的任務數據

## 🔒 安全考慮

- ✅ 不會刪除正在進行的任務
- ✅ 不會影響下載隊列
- ✅ 不會刪除已下載的文件
- ✅ 操作是原子性的（使用 mutex 保護）

## 📝 更新的文件

1. **後端**
   - `internal/server/server.go` - 新增 API 處理函數

2. **前端**
   - `extension/popup.html` - 新增清除按鈕
   - `extension/popup.js` - 新增清除邏輯和按鈕控制

3. **文檔**
   - `SERVER-MODE.md` - 更新 API 使用說明
   - `extension/README.md` - 更新擴展功能說明
   - `CLEAR-TASKS-FEATURE.md` - 本功能說明文檔

4. **測試**
   - `test-clear-tasks.bat` - Windows 測試腳本

## 🚀 快速開始

1. **啟動服務器**
```bash
.\jable-downloader.exe --server
```

2. **測試 API**
```bash
.\test-clear-tasks.bat
```

3. **使用擴展**
   - 打開 Chrome 擴展
   - 點擊清除按鈕

## 💡 提示

- 清除操作是不可逆的
- 建議在確認任務已完成後再清除
- 清除不會影響已下載的視頻文件
- 擴展會自動刷新任務列表

---

**更新日期：** 2026-02-16  
**版本：** v1.1.0
