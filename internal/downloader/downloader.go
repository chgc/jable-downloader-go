package downloader

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/grafov/m3u8"
	"github.com/jable-downloader-go/internal/crawler"
	"github.com/jable-downloader-go/internal/encoder"
	"github.com/jable-downloader-go/internal/merger"
	"github.com/jable-downloader-go/pkg/utils"
)

type Downloader struct {
	URL        string
	DirName    string
	FolderPath string
	AutoMode   bool // 自動模式（服務器模式使用）
	EncodeMode encoder.EncodeMode // 指定轉檔模式
}

func NewDownloader(url string) (*Downloader, error) {
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("無效的 URL 格式")
	}
	
	dirName := parts[len(parts)-1]
	folderPath := filepath.Join("download", dirName)
	
	return &Downloader{
		URL:        url,
		DirName:    dirName,
		FolderPath: folderPath,
		AutoMode:   false,
		EncodeMode: encoder.NoEncode, // 默認不轉檔
	}, nil
}

func (d *Downloader) Download() error {
	// 自動模式或詢問是否轉檔
	var encodeMode encoder.EncodeMode
	if d.AutoMode {
		encodeMode = d.EncodeMode // 使用預設的轉檔模式
		if encodeMode != encoder.NoEncode {
			fmt.Printf("使用轉檔模式: %d (自動模式)\n", encodeMode)
		}
	} else {
		encodeMode = d.askEncodeMode() // 互動模式詢問
	}
	
	fmt.Printf("正在下載影片: %s\n", d.URL)
	
	// 檢查是否已存在
	finalPath := filepath.Join(d.FolderPath, d.DirName+".mp4")
	if utils.FileExists(finalPath) {
		fmt.Println("番號資料夾已存在, 跳過...")
		return nil
	}
	
	// 建立資料夾
	if err := utils.EnsureDir(d.FolderPath); err != nil {
		return fmt.Errorf("建立資料夾失敗: %v", err)
	}
	
	// 使用 ChromeDP 獲取 M3U8 URL
	m3u8URL, htmlContent, err := d.getM3U8URL()
	if err != nil {
		return fmt.Errorf("獲取 M3U8 URL 失敗: %v", err)
	}
	
	fmt.Printf("m3u8url: %s\n", m3u8URL)
	
	// 解析 M3U8
	tsList, aesKey, iv, err := d.parseM3U8(m3u8URL)
	if err != nil {
		return fmt.Errorf("解析 M3U8 失敗: %v", err)
	}
	
	// 下載 TS 片段
	c, err := crawler.NewCrawler(d.FolderPath, tsList, aesKey, iv)
	if err != nil {
		return fmt.Errorf("建立爬蟲失敗: %v", err)
	}
	
	if err := c.Download(); err != nil {
		return fmt.Errorf("下載失敗: %v", err)
	}
	
	// 合併 MP4
	if err := merger.MergeTSFiles(d.FolderPath, tsList); err != nil {
		return fmt.Errorf("合併失敗: %v", err)
	}
	
	// 清理臨時檔案
	utils.DeleteFiles(d.FolderPath, d.DirName+".mp4")
	
	// 下載封面
	if err := utils.DownloadCover(htmlContent, d.FolderPath); err != nil {
		fmt.Printf("下載封面失敗: %v\n", err)
	}
	
	// 轉檔
	if err := encoder.FFmpegEncode(d.FolderPath, d.DirName, encodeMode); err != nil {
		fmt.Printf("轉檔失敗: %v\n", err)
	}
	
	return nil
}

func (d *Downloader) getM3U8URL() (string, string, error) {
	// 檢測是否在容器環境中運行
	isContainer := utils.IsRunningInContainer()
	
	// 設置 Chrome 選項
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	
	// 容器環境需要額外的選項
	if isContainer {
		opts = append(opts,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-software-rasterizer", true),
		)
		fmt.Println("檢測到容器環境，使用容器優化配置")
	}
	
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	var htmlContent string
	
	err := chromedp.Run(ctx,
		chromedp.Navigate(d.URL),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	
	if err != nil {
		return "", "", err
	}
	
	// 使用正則表達式提取 M3U8 URL
	re := regexp.MustCompile(`https://[^\s"]+\.m3u8`)
	matches := re.FindStringSubmatch(htmlContent)
	
	if len(matches) == 0 {
		return "", "", fmt.Errorf("在頁面中找不到 M3U8 URL")
	}
	
	return matches[0], htmlContent, nil
}

func (d *Downloader) parseM3U8(m3u8URL string) ([]string, []byte, []byte, error) {
	// 下載 M3U8 檔案
	resp, err := http.Get(m3u8URL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer resp.Body.Close()
	
	playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return nil, nil, nil, err
	}
	
	if listType != m3u8.MEDIA {
		return nil, nil, nil, fmt.Errorf("不支援的 M3U8 類型")
	}
	
	mediapl := playlist.(*m3u8.MediaPlaylist)
	
	// 取得基礎 URL
	baseURL := m3u8URL[:strings.LastIndex(m3u8URL, "/")]
	
	// 收集 TS URLs
	var tsList []string
	for _, segment := range mediapl.Segments {
		if segment != nil && segment.URI != "" {
			tsURL := baseURL + "/" + segment.URI
			tsList = append(tsList, tsURL)
		}
	}
	
	// 處理加密
	var aesKey []byte
	var iv []byte
	
	if mediapl.Key != nil && mediapl.Key.URI != "" {
		keyURL := baseURL + "/" + mediapl.Key.URI
		
		resp, err := http.Get(keyURL)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("獲取金鑰失敗: %v", err)
		}
		defer resp.Body.Close()
		
		aesKey, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("讀取金鑰失敗: %v", err)
		}
		
		// 處理 IV
		if mediapl.Key.IV != "" {
			ivStr := strings.TrimPrefix(mediapl.Key.IV, "0x")
			iv, err = hex.DecodeString(ivStr)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("解析 IV 失敗: %v", err)
			}
		}
	}
	
	return tsList, aesKey, iv, nil
}

func (d *Downloader) askEncodeMode() encoder.EncodeMode {
	fmt.Print("要轉檔嗎? [y/n]: ")
	var answer string
	fmt.Scanln(&answer)
	
	if strings.ToLower(answer) != "y" {
		return encoder.NoEncode
	}
	
	fmt.Print("選擇轉檔方案 [1:僅轉換格式(默認,推薦) 2:NVIDIA GPU 轉檔 3:CPU 轉檔]: ")
	var mode string
	fmt.Scanln(&mode)
	
	switch mode {
	case "2":
		return encoder.GPUEncode
	case "3":
		return encoder.CPUEncode
	default:
		return encoder.FastEncode
	}
}
