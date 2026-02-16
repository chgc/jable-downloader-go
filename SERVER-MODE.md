# 服務器模式說明

## 📋 互動式選項處理

在服務器模式（API 模式）下，以下原本需要手動選擇的選項已自動處理：

### 1. 是否轉檔

**命令行模式（互動）：**
```
要轉檔嗎? [y/n]: _
```

**服務器模式（自動）：**
- 由 API 請求的 `convert` 參數決定
- `convert: false` → 不轉檔（默認）
- `convert: true` → 使用快速轉檔（僅轉換格式）

### 2. 轉檔方案

**命令行模式（互動）：**
```
選擇轉檔方案 [1:僅轉換格式(默認,推薦) 2:NVIDIA GPU 轉檔 3:CPU 轉檔]: _
```

**服務器模式（自動）：**
- 固定使用 **方案 1**（FastEncode - 僅轉換格式）
- 這是最快且推薦的方案
- 無損轉檔，不重新編碼

## 🔧 轉檔模式說明

| 模式 | 編號 | 說明 | 速度 | 服務器模式 |
|------|------|------|------|-----------|
| NoEncode | 0 | 不轉檔，保留原始 MP4 | N/A | convert: false |
| FastEncode | 1 | 僅轉換格式，無損 | ⚡⚡⚡ | convert: true (默認) |
| GPUEncode | 2 | NVIDIA GPU 硬體加速 | ⚡⚡ | 暫不支援 |
| CPUEncode | 3 | CPU 編碼 | ⚡ | 暫不支援 |

## 📡 API 使用示例

### 下載不轉檔（推薦，最快）

```bash
curl -X POST http://localhost:18080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://jable.tv/videos/xxx/",
    "convert": false
  }'
```

### 下載並快速轉檔

```bash
curl -X POST http://localhost:18080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://jable.tv/videos/xxx/",
    "convert": true
  }'
```

### 查看任務列表

```bash
curl http://localhost:18080/api/tasks
```

### 清除已完成的任務

```bash
# 使用 DELETE 方法
curl -X DELETE http://localhost:18080/api/tasks/clear-completed

# 或使用 POST 方法
curl -X POST http://localhost:18080/api/tasks/clear-completed
```

**響應示例：**
```json
{
  "success": true,
  "message": "Successfully cleared 5 completed task(s)",
  "cleared_count": 5
}
```

### Chrome 擴展

擴展默認使用 `convert: false`（不轉檔），確保最快的下載速度。

## 🆚 命令行模式 vs 服務器模式

### 命令行模式（互動式）

```bash
.\jable-downloader.exe --url https://jable.tv/videos/xxx/

# 會提示：
要轉檔嗎? [y/n]: y
選擇轉檔方案 [1:僅轉換格式(默認,推薦) 2:NVIDIA GPU 轉檔 3:CPU 轉檔]: 1
```

**優勢**：
- ✅ 完全控制每個選項
- ✅ 適合手動下載

### 服務器模式（自動化）

```bash
.\jable-downloader.exe --server
```

**優勢**：
- ✅ 無需人工介入
- ✅ 適合 Chrome 擴展調用
- ✅ 可批量處理
- ✅ 使用合理的默認值

## 💡 設計考量

### 為什麼服務器模式使用 FastEncode？

1. **速度最快**：不重新編碼，僅調整容器格式
2. **質量無損**：保持原始視頻質量
3. **相容性好**：適用於所有系統
4. **資源消耗低**：不需要 GPU 支援

### 為什麼默認不轉檔？

1. **更快完成**：減少處理時間
2. **保留原始**：部分用戶可能需要原始格式
3. **可選性**：用戶可在 API 請求中指定

## 🔮 未來擴展

可能在未來版本中支援更多選項：

### API 請求增強

```json
{
  "url": "https://jable.tv/videos/xxx/",
  "convert": true,
  "encode_mode": "fast",  // "none" | "fast" | "gpu" | "cpu"
  "download_cover": true,  // 是否下載封面
  "auto_cleanup": true     // 是否自動清理臨時檔案
}
```

### 全局配置文件

```yaml
# config.yaml
server:
  port: 18080
  default_convert: false
  default_encode_mode: "fast"

downloader:
  max_workers: 8
  timeout: 300
  download_cover: true
  auto_cleanup: true
```

## 📝 代碼實現

### Downloader 結構

```go
type Downloader struct {
    URL        string
    DirName    string
    FolderPath string
    AutoMode   bool             // 自動模式（服務器使用）
    EncodeMode encoder.EncodeMode // 轉檔模式
}
```

### 自動模式邏輯

```go
func (d *Downloader) Download() error {
    var encodeMode encoder.EncodeMode
    if d.AutoMode {
        // 服務器模式：使用預設值
        encodeMode = d.EncodeMode
    } else {
        // 命令行模式：詢問用戶
        encodeMode = d.askEncodeMode()
    }
    // ... 繼續下載
}
```

## ❓ 常見問題

### Q: 如何在服務器模式使用 GPU 轉檔？

A: 目前固定使用 FastEncode。若需要 GPU 轉檔，建議：
1. 先用 API 下載（不轉檔）
2. 手動用命令行工具轉檔

### Q: 可以更改服務器模式的默認轉檔方式嗎？

A: 目前需要修改代碼。未來版本會支援配置文件。

### Q: 轉檔會影響下載速度嗎？

A: 
- **FastEncode**：影響很小，主要是 I/O 時間
- **GPU/CPU Encode**：會顯著增加處理時間

### Q: 不轉檔的原始 MP4 能正常播放嗎？

A: 可以！合併後的 MP4 已經可以正常播放。轉檔主要是為了：
- 優化檔案大小
- 提高相容性
- 調整編碼格式

---

**相關文檔**：
- [extension/README.md](../extension/README.md) - Chrome 擴展使用
- [QUICKSTART.md](../QUICKSTART.md) - 快速開始
- [README.md](../README.md) - 完整功能說明
