# 快速開始指南

本指南包含兩種使用方式：

- **方式一**：命令行下載（原有功能）
- **方式二**：Chrome 擴展下載（新功能，更方便）

---

## 🌟 方式二：Chrome 擴展下載（推薦）

### 1. 構建並啟動 API 服務器

**選擇 A：直接運行**
```bash
# Windows
build.bat
start-server.bat

# Linux/Mac
./build.sh
./start-server.sh
```

**選擇 B：使用 Docker**
```bash
# Windows
docker-start.bat

# Linux/Mac
./docker-start.sh
```

### 2. 安裝 Chrome 擴展

1. 打開 Chrome：`chrome://extensions/`
2. 開啟「開發人員模式」
3. 點擊「載入未封裝項目」
4. 選擇 `extension` 資料夾

### 3. 開始使用

1. 訪問 Jable 視頻頁面
2. 點擊視頻標題下方的「📥 下載影片」按鈕
3. 完成！影片會自動下載到 `download/` 資料夾

**優勢**：
- ✅ 無需複製粘貼網址
- ✅ 一鍵下載，自動管理
- ✅ 實時顯示下載狀態
- ✅ 可以同時下載多個影片

詳細說明：[extension/README.md](extension/README.md)

---

## 📟 方式一：命令行下載

## 🚀 快速設定（5 分鐘搞定）

### Windows 使用者

1. **安裝 FFmpeg**
   - 下載：https://www.ffmpeg.org/download.html
   - 解壓縮並加入系統 PATH
   - 測試：開啟 CMD 執行 `ffmpeg -version`

2. **編譯程式**
   ```cmd
   build.bat
   ```

3. **開始使用**
   ```cmd
   jable-downloader.exe
   輸入 jable 網址: https://jable.tv/videos/xxxxx/
   ```

### Linux / macOS 使用者

1. **安裝 FFmpeg**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install ffmpeg
   
   # macOS
   brew install ffmpeg
   ```

2. **編譯程式**
   ```bash
   chmod +x build.sh
   ./build.sh
   ```

3. **開始使用**
   ```bash
   ./jable-downloader
   ```

## 📝 使用範例

### 範例 1: 下載單一影片
```bash
# 互動模式
./jable-downloader
輸入 jable 網址: https://jable.tv/videos/ipx-486/

# 或直接指定
./jable-downloader --url https://jable.tv/videos/ipx-486/
```

### 範例 2: 隨機下載
```bash
./jable-downloader --random
```

### 範例 3: 批次下載演員作品
```bash
./jable-downloader --all-urls https://jable.tv/models/some-actress/
```

## ⚙️ 轉檔選項說明

下載時會詢問：
```
要轉檔嗎? [y/n]: y
選擇轉檔方案 [1:僅轉換格式(默認,推薦) 2:NVIDIA GPU 轉檔 3:CPU 轉檔]: 1
```

- **選項 1** (推薦): 最快，不重新編碼，僅調整格式
- **選項 2**: 需要 NVIDIA 顯卡，使用硬體加速
- **選項 3**: 純 CPU 編碼，較慢但相容性最好

## 📁 下載檔案位置

所有影片下載至 `download/` 資料夾，結構如下：
```
download/
└── ipx-486/
    ├── ipx-486.mp4    # 影片檔案
    └── ipx-486.jpg    # 封面圖片
```

## ❓ 常見問題

**Q: 提示找不到 FFmpeg？**
- 確認已安裝 FFmpeg 並加入 PATH
- 測試指令：`ffmpeg -version`

**Q: ChromeDP 相關錯誤？**
- 首次執行會自動下載 Chrome，需要網路連線
- 如果失敗，手動安裝 Google Chrome 瀏覽器

**Q: 下載失敗或卡住？**
- 檢查網路連線
- 確認網址是否正確（必須是 jable.tv 網域）
- 嘗試重新執行程式

**Q: 轉檔失敗？**
- 使用選項 1（無損轉檔）最穩定
- 確認 FFmpeg 安裝正確

## 🎯 進階設定

### 調整並發數
編輯 `internal/config/config.go`：
```go
const (
    MaxWorkers = 8  // 改為 4-16 之間的值
)
```

### 自訂編譯
```bash
# 優化編譯（減小檔案大小）
go build -ldflags="-s -w" -o jable-downloader ./cmd/jable-downloader

# 交叉編譯到其他平台
GOOS=linux GOARCH=amd64 go build -o jable-downloader-linux ./cmd/jable-downloader
```

## 🆚 與 Python 版本比較

| 特性 | Python 版 | Go 版 |
|-----|----------|------|
| 執行速度 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 記憶體使用 | ~150 MB | ~50 MB |
| 安裝步驟 | 需要 Python + 套件 | 單一執行檔 |
| 啟動時間 | 2-3 秒 | 0.5 秒 |
| 依賴管理 | pip | 無（內建） |

## 💡 小技巧

1. **批次下載**：可以搭配 shell 腳本批次處理多個網址
2. **排程下載**：使用系統排程工具（Windows 工作排程器 / Linux cron）
3. **省空間**：轉檔完成後，原始 TS 片段會自動清除
4. **斷點續傳**：如果下載中斷，重新執行會跳過已下載的片段

---

**遇到問題？** 請查看完整的 [README.md](README.md) 或提交 Issue。
