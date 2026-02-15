# Jable TV Downloader - Go 版本

## 📋 目錄

- [專案簡介](#專案簡介)
- [🌟 新功能：Chrome 擴展](#-新功能chrome-擴展)
- [功能特色](#功能特色)
- [系統需求](#系統需求)

## 專案簡介

這是 [JableTVDownload](../README.md) 的 Golang 重寫版本，提供相同的功能但具有更好的效能和跨平台支援。

## 🌟 新功能：Chrome 擴展

現在支援通過 Chrome 瀏覽器擴展一鍵下載 Jable 影片！

### 快速開始

```bash
# 1. 啟動 API 服務器
./start-server.bat          # Windows
./start-server.sh           # Linux/Mac

# 或使用 Docker
./docker-start.bat          # Windows
./docker-start.sh           # Linux/Mac

# 2. 安裝 Chrome 擴展
# 打開 chrome://extensions/
# 載入 extension/ 資料夾

# 3. 訪問 Jable 視頻頁面，點擊下載按鈕即可
```

**詳細說明**：
- [Chrome 擴展完整文檔](extension/README.md)
- [快速開始指南](QUICKSTART.md)

### 使用方式對比

| 方式 | 命令行模式 | Chrome 擴展模式 |
|-----|----------|---------------|
| 啟動方式 | `./jable-downloader --url <URL>` | `./jable-downloader --server` |
| 使用場景 | 批次下載、腳本自動化 | 瀏覽時一鍵下載 |
| 操作步驟 | 複製網址 → 貼到終端 | 直接點擊按鈕 |
| 優勢 | 適合批量處理 | 方便快捷 |

## 功能特色

✨ **核心功能**
- 🎬 下載 Jable TV 影片（M3U8 串流）
- 🔐 支援 AES-128-CBC 加密解密
- ⚡ 並發下載（8 個 goroutines）
- 🎞️ FFmpeg 影片轉檔（無損/GPU/CPU）
- 🖼️ 自動下載影片封面
- 🎲 隨機推薦影片
- 📦 批次下載演員所有影片
- 🌐 **Chrome 擴展一鍵下載（NEW）**
- 🐳 **Docker 容器化部署（NEW）**
- 📡 **HTTP API 服務器（NEW）**
- 📋 **下載隊列管理（NEW）** - 依序處理下載任務，實時查看隊列狀態

🚀 **Go 版本優勢**
- 單一執行檔，無需 Python 環境
- 更快的執行速度
- 更低的記憶體佔用
- 跨平台支援（Windows/Linux/macOS）
- 無需安裝 ChromeDriver（內建 ChromeDP）

🎯 **使用模式**
- **命令行模式**：傳統的終端下載方式
- **服務器模式**：啟動 HTTP API 接受下載請求
- **擴展模式**：通過 Chrome 擴展一鍵下載

📋 **下載隊列特性**
- ⏳ 自動排隊：多個下載任務自動依序執行
- 📊 實時狀態：Extension 顯示當前下載和排隊任務
- 🔄 自動刷新：每 3 秒更新隊列狀態
- 🎯 FIFO 處理：先進先出，確保公平下載

## 系統需求

### 必要軟體
- **FFmpeg**: 用於影片轉檔
  - Windows: 從 [FFmpeg 官網](https://www.ffmpeg.org/) 下載並加入 PATH
  - Linux: `sudo apt-get install ffmpeg`
  - macOS: `brew install ffmpeg`

### 選用軟體
- **Google Chrome**: ChromeDP 會自動下載，但安裝 Chrome 可提高穩定性

## 安裝與編譯

### 方法一：從原始碼編譯

```bash
# 1. 進入專案目錄
cd jable-downloader-go

# 2. 下載依賴套件
go mod tidy

# 3. 編譯
go build -o jable-downloader.exe ./cmd/jable-downloader

# Linux/macOS
go build -o jable-downloader ./cmd/jable-downloader
```

### 方法二：交叉編譯

```bash
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o jable-downloader-windows-amd64.exe ./cmd/jable-downloader

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o jable-downloader-linux-amd64 ./cmd/jable-downloader

# macOS 64-bit
GOOS=darwin GOARCH=amd64 go build -o jable-downloader-darwin-amd64 ./cmd/jable-downloader
```

## 使用方式

### 1. 互動模式（預設）

```bash
./jable-downloader

# 輸入影片網址
輸入 jable 網址: https://jable.tv/videos/ipx-486/
```

### 2. 指定 URL 下載

```bash
./jable-downloader --url https://jable.tv/videos/ipx-486/
```

### 3. 隨機下載推薦影片

```bash
./jable-downloader --random
```

### 4. 批次下載演員所有影片

```bash
./jable-downloader --all-urls https://jable.tv/models/some-actress/
```

## 轉檔選項

下載時會詢問是否轉檔：

```
要轉檔嗎? [y/n]: y
選擇轉檔方案 [1:僅轉換格式(默認,推薦) 2:NVIDIA GPU 轉檔 3:CPU 轉檔]: 1
```

- **選項 1**: 快速無損轉檔（推薦）- 僅調整格式，不重新編碼
- **選項 2**: NVIDIA GPU 轉檔 - 使用 NVENC 硬體加速
- **選項 3**: CPU 轉檔 - 使用 x264 編碼器

## 專案結構

```
jable-downloader-go/
├── cmd/
│   └── jable-downloader/    # 主程式入口
│       └── main.go
├── internal/                 # 內部套件（不對外公開）
│   ├── config/              # 全局配置
│   ├── crawler/             # 並發下載器
│   ├── downloader/          # 下載邏輯
│   ├── encoder/             # FFmpeg 整合
│   ├── merger/              # 檔案合併
│   └── parser/              # 命令列解析
├── pkg/                     # 公開套件
│   └── utils/               # 工具函式
├── download/                # 下載目錄（自動建立）
├── go.mod                   # Go 模組定義
├── go.sum                   # 依賴套件鎖定
├── PLAN.md                  # 開發計畫
└── README.md                # 本文件
```

## 技術架構

### 使用的 Go 套件

- `github.com/chromedp/chromedp` - 瀏覽器自動化
- `github.com/PuerkitoBio/goquery` - HTML 解析
- `github.com/grafov/m3u8` - M3U8 播放列表解析
- 標準庫：`crypto/aes`, `crypto/cipher` - AES 解密
- 標準庫：`net/http` - HTTP 請求
- 標準庫：`sync` - 並發控制

### 核心技術

1. **ChromeDP**: Pure Go 實作的 Chrome DevTools Protocol，無需外部 ChromeDriver
2. **Goroutines**: 輕量級並發，8 個 worker 並行下載
3. **AES-CBC**: 標準庫實作的加密解密
4. **Worker Pool**: 生產者-消費者模式管理下載任務

## 效能比較

| 項目 | Python 版本 | Go 版本 |
|-----|-----------|---------|
| 啟動時間 | ~2-3 秒 | ~0.5 秒 |
| 記憶體佔用 | ~150-200 MB | ~50-80 MB |
| 編譯產物 | 需 Python 環境 | 單一執行檔 |
| 依賴管理 | pip + requirements.txt | go mod |

## 常見問題

### Q: 找不到 FFmpeg？
A: 請確保 FFmpeg 已安裝並加入系統 PATH。測試方式：`ffmpeg -version`

### Q: ChromeDP 無法啟動？
A: 首次執行會自動下載 Chrome，請確保網路連線正常。

### Q: 下載速度慢？
A: 可以修改 `internal/config/config.go` 的 `MaxWorkers` 增加並發數（建議不超過 16）

### Q: 轉檔失敗？
A: 確認 FFmpeg 安裝正確，選項 1（無損轉檔）最穩定。

## 開發相關

### 執行測試
```bash
go test ./...
```

### 程式碼格式化
```bash
go fmt ./...
```

### 靜態分析
```bash
go vet ./...
```

## 授權

與原 Python 版本相同，請參閱 [LICENSE](../LICENSE)

## 致謝

- 原始 Python 版本作者：hcjohn463
- Go 移植版本：基於原始專案重新實作

## 更新日誌

### v2.1.0 (2026-02-15)
- 📋 **新增下載隊列功能**
  - 服務器端依序處理下載任務（一次一個）
  - Extension 可查看當前下載和排隊任務
  - 實時狀態更新（排隊中/下載中/已完成/失敗）
  - 任務列表按時間排序，自動刷新
- 🎨 Extension UI 優化
  - 新增隊列狀態顯示區域
  - 美化任務卡片設計
  - 支持滾動查看長隊列

### v2.0.0 (2026-02-14)
- 🎉 首次發布 Golang 版本
- ✨ 完整功能對應 Python 版本
- ⚡ 效能優化和記憶體改善
- 📦 單一執行檔部署

---

**如果覺得好用，請給個 Star ⭐ 謝謝！**
