# Python 轉 Golang 專案執行計畫

## 專案概述
將 JableTVDownload Python 專案轉換為 Golang 版本。這是一個用於下載 Jable TV 影片的命令列工具，支援 M3U8 串流下載、AES 解密、多執行緒下載、影片合併和 FFmpeg 轉檔功能。

## 原始專案功能分析

### 核心功能模組
1. **主程式 (main.py)**: 程式入口點，處理命令列參數
2. **下載模組 (download.py)**: 核心下載邏輯，使用 Selenium 取得 M3U8 網址
3. **爬蟲模組 (crawler.py)**: 並發下載 TS 片段（8 個執行緒）
4. **合併模組 (merge.py)**: 將 TS 片段合併成完整 MP4
5. **編碼模組 (encode.py)**: FFmpeg 轉檔支援（無損/GPU/CPU）
6. **封面模組 (cover.py)**: 下載影片封面圖片
7. **電影清單模組 (movies.py)**: 批次下載演員所有影片
8. **參數模組 (args.py)**: 命令列參數解析和隨機推薦
9. **配置模組 (config.py)**: HTTP headers 配置
10. **刪除模組 (delete.py)**: 清理臨時檔案

### 主要技術特性
- M3U8 串流解析與下載
- AES-128-CBC 解密
- 並發下載（8 執行緒）
- Selenium WebDriver 自動化
- FFmpeg 影片處理
- 命令列參數支援

### 使用方式
- 互動式輸入網址
- 指定 URL: `--url`
- 隨機下載: `--random True`
- 批次下載: `--all-urls`

## Golang 轉換工作計畫

### 第一階段：專案結構建立
- [x] 建立 `jable-downloader-go` 資料夾
- [ ] 初始化 Go module (`go.mod`)
- [x] 設計專案目錄結構
  ```
  jable-downloader-go/
  ├── cmd/
  │   └── jable-downloader/
  │       └── main.go
  ├── internal/
  │   ├── config/
  │   ├── crawler/
  │   ├── downloader/
  │   ├── encoder/
  │   ├── merger/
  │   └── parser/
  ├── pkg/
  │   └── utils/
  ├── go.mod
  ├── go.sum
  └── README.md
  ```

### 第二階段：核心模組實作
- [x] **config 套件**: HTTP headers 和全局配置
- [x] **parser 套件**: 命令列參數解析（使用 flag）
- [x] **downloader 套件**: 
  - ChromeDP 整合
  - M3U8 解析
  - AES 解密邏輯
- [x] **crawler 套件**: 
  - 並發下載 TS 片段（goroutines）
  - 下載進度追蹤
- [x] **merger 套件**: TS 片段合併
- [x] **encoder 套件**: FFmpeg 整合
- [x] **utils 套件**: 
  - 檔案操作
  - 封面下載
  - 隨機推薦

### 第三階段：主程式整合
- [x] 實作主程式邏輯
- [x] 整合所有模組
- [x] 錯誤處理和日誌
- [x] 使用者互動界面

### 第四階段：測試與優化
- [x] 功能測試（編譯成功）
- [x] 錯誤處理完善
- [ ] 實際下載測試（需使用者測試）
- [ ] 跨平台測試（已提供編譯腳本）

### 第五階段：文檔與部署
- [x] 撰寫 README.md（繁體中文）
- [x] 使用說明文檔 (QUICKSTART.md)
- [x] 編譯腳本（Windows .bat / Linux .sh）
- [x] 交叉編譯說明
- [x] 專案總結 (SUMMARY.md)

## 技術選型

### Go 第三方套件
1. **命令列**: `github.com/spf13/cobra` 或標準 `flag`
2. **HTTP 請求**: 標準 `net/http` + `github.com/go-resty/resty/v2`
3. **HTML 解析**: `github.com/PuerkitoBio/goquery`
4. **M3U8 解析**: `github.com/grafov/m3u8`
5. **AES 解密**: 標準 `crypto/aes`, `crypto/cipher`
6. **瀏覽器自動化**: `github.com/chromedp/chromedp`
7. **並發控制**: 標準 goroutines + channels + `sync`
8. **進度條**: `github.com/schollz/progressbar/v3`

## 關鍵技術實作重點

### 1. Selenium 替代方案
- 使用 ChromeDP（pure Go）替代 Selenium
- 無需外部 ChromeDriver
- 更輕量級的瀏覽器自動化

### 2. 並發下載
- 使用 goroutines 替代 Python threading
- Worker pool 模式控制並發數
- Channel 管理任務分發和結果收集

### 3. AES 解密
- 使用 Go 標準庫 `crypto/aes`
- CBC 模式實作
- IV 處理

### 4. FFmpeg 整合
- 使用 `os/exec` 調用 FFmpeg
- 三種轉檔模式：無損/GPU/CPU

### 5. 檔案操作
- 使用標準 `os`, `io`, `path/filepath`
- 確保跨平台路徑處理

## 預期成果
- 功能完全對應 Python 版本
- 更好的執行效能（Go 編譯優勢）
- 單一可執行檔，無需依賴環境
- 跨平台支援
- 更低的記憶體佔用

## 注意事項
1. 保持與原 Python 版本相同的使用者體驗
2. 命令列參數保持一致
3. 下載資料夾結構保持一致
4. 錯誤訊息使用繁體中文
5. 確保 FFmpeg 外部依賴說明清楚
