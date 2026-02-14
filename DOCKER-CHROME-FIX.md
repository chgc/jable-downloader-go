# Docker 容器 Chrome 錯誤修正

## 問題描述

在 Docker 容器中運行時出現錯誤：
```
chrome failed to start
```

## 原因

Docker 容器是隔離的沙盒環境，Chrome/Chromium 默認的安全設置在容器中無法正常工作，需要特殊的啟動參數。

## 解決方案

### 1. 自動檢測容器環境

添加了 `IsRunningInContainer()` 函數，自動檢測是否在容器中運行：

```go
func IsRunningInContainer() bool {
    // 檢查 /.dockerenv 文件
    if _, err := os.Stat("/.dockerenv"); err == nil {
        return true
    }
    
    // 檢查 /proc/1/cgroup
    if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
        if strings.Contains(string(data), "docker") {
            return true
        }
    }
    
    // 檢查環境變量
    if os.Getenv("DOCKER_CONTAINER") == "true" {
        return true
    }
    
    return false
}
```

### 2. 容器優化的 Chrome 配置

在容器環境中自動添加必要的啟動參數：

```go
opts := append(chromedp.DefaultExecAllocatorOptions[:],
    chromedp.Flag("disable-gpu", true),
    chromedp.Flag("disable-dev-shm-usage", true),
    chromedp.Flag("disable-setuid-sandbox", true),
    chromedp.Flag("disable-extensions", true),
)

if isContainer {
    opts = append(opts,
        chromedp.Flag("no-sandbox", true),         // 必須！
        chromedp.Flag("headless", true),           // 無頭模式
        chromedp.Flag("disable-software-rasterizer", true),
    )
}
```

### 3. Dockerfile 設置環境變量

```dockerfile
ENV DOCKER_CONTAINER=true
```

## 關鍵參數說明

| 參數 | 說明 | 必要性 |
|------|------|--------|
| `--no-sandbox` | 禁用沙盒模式 | ✅ 容器必須 |
| `--disable-gpu` | 禁用 GPU | ✅ 推薦 |
| `--disable-dev-shm-usage` | 不使用 /dev/shm | ✅ 推薦 |
| `--disable-setuid-sandbox` | 禁用 setuid 沙盒 | ✅ 容器必須 |
| `--headless` | 無頭模式 | ✅ 容器推薦 |

## 使用方法

### 重新構建 Docker 鏡像

```bash
# 停止舊容器
docker-compose down

# 重新構建並啟動
docker-compose up -d --build

# 查看日誌
docker-compose logs -f
```

### 測試

```bash
# 發送測試請求
curl -X POST http://localhost:18080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://jable.tv/videos/xxx/",
    "convert": false
  }'
```

如果成功，應該會看到：
```json
{
  "success": true,
  "message": "Download task created",
  "task_id": "task_xxxxx"
}
```

並且容器日誌中會顯示：
```
檢測到容器環境，使用容器優化配置
正在下載影片: https://jable.tv/videos/xxx/
```

## 故障排查

### 1. 仍然出現 Chrome 錯誤

檢查容器日誌：
```bash
docker-compose logs jable-downloader
```

確認是否看到：
```
檢測到容器環境，使用容器優化配置
```

如果沒有，手動設置環境變量：
```yaml
# docker-compose.yml
environment:
  - DOCKER_CONTAINER=true
```

### 2. 檢查 Chromium 是否安裝

進入容器檢查：
```bash
docker-compose exec jable-downloader sh
which chromium-browser
chromium-browser --version
```

### 3. 手動測試 Chrome

在容器中手動運行：
```bash
docker-compose exec jable-downloader sh
chromium-browser --no-sandbox --headless --disable-gpu --dump-dom https://www.google.com
```

如果這能成功，說明 Chrome 配置正確。

### 4. 增加共享內存

如果仍有問題，可能需要增加共享內存：

```yaml
# docker-compose.yml
services:
  jable-downloader:
    shm_size: '2gb'  # 增加共享內存
```

或使用：
```yaml
volumes:
  - /dev/shm:/dev/shm  # 掛載主機的共享內存
```

## 為什麼需要 --no-sandbox？

Chrome 的沙盒模式需要特殊的內核權限（如 CAP_SYS_ADMIN），這在 Docker 容器中通常是被限制的。有兩種解決方案：

**方案 1：禁用沙盒（推薦）**
```
--no-sandbox
```

**方案 2：給容器特權（不推薦，不安全）**
```yaml
# docker-compose.yml
privileged: true
```

我們選擇方案 1，因為：
- ✅ 更安全
- ✅ 不需要額外權限
- ✅ 在隔離的容器環境中已足夠安全

## 本地運行 vs 容器運行

| 環境 | Chrome 配置 | 是否需要 --no-sandbox |
|------|-------------|---------------------|
| Windows 本地 | 標準配置 | ❌ 否 |
| Linux 本地 | 標準配置 | ❌ 否 |
| Docker 容器 | 優化配置 | ✅ 是 |
| Kubernetes | 優化配置 | ✅ 是 |

程序會自動檢測環境並使用正確的配置！

## 相關資源

- [ChromeDP Docker 示例](https://github.com/chromedp/chromedp/blob/master/examples/docker/Dockerfile)
- [Chrome Headless Docker](https://github.com/Zenika/alpine-chrome)
- [Puppeteer Troubleshooting](https://github.com/puppeteer/puppeteer/blob/main/docs/troubleshooting.md#running-puppeteer-in-docker)

---

**已修正並測試通過！** 🎉
