package crawler

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jable-downloader-go/internal/config"
)

type Crawler struct {
	client      *http.Client
	cipher      cipher.BlockMode
	folderPath  string
	downloadList []string
	mu          sync.Mutex
	progress    int
	total       int
}

func NewCrawler(folderPath string, tsList []string, aesKey []byte, iv []byte) (*Crawler, error) {
	c := &Crawler{
		client:       &http.Client{Timeout: 30 * time.Second},
		folderPath:   folderPath,
		downloadList: make([]string, len(tsList)),
		total:        len(tsList),
	}
	
	copy(c.downloadList, tsList)
	
	// 如果有 AES 金鑰，建立解密器
	if len(aesKey) > 0 {
		block, err := aes.NewCipher(aesKey)
		if err != nil {
			return nil, fmt.Errorf("建立 AES cipher 失敗: %v", err)
		}
		c.cipher = cipher.NewCBCDecrypter(block, iv[:16])
	}
	
	return c, nil
}

func (c *Crawler) Download() error {
	startTime := time.Now()
	fmt.Printf("開始下載 %d 個檔案..\n", c.total)
	fmt.Printf("預計等待時間: %.2f 分鐘 (視影片長度與網路速度而定)\n", float64(c.total)/150)
	
	var wg sync.WaitGroup
	jobs := make(chan string, c.total)
	
	// 啟動 worker pool
	for i := 0; i < config.MaxWorkers; i++ {
		wg.Add(1)
		go c.worker(&wg, jobs)
	}
	
	// 發送任務
	for _, url := range c.downloadList {
		jobs <- url
	}
	close(jobs)
	
	// 等待完成
	wg.Wait()
	
	elapsed := time.Since(startTime)
	fmt.Printf("\n花費 %.2f 分鐘爬取完成!\n", elapsed.Minutes())
	
	return nil
}

func (c *Crawler) worker(wg *sync.WaitGroup, jobs <-chan string) {
	defer wg.Done()
	
	for url := range jobs {
		c.downloadOne(url)
	}
}

func (c *Crawler) downloadOne(url string) {
	fileName := filepath.Base(url)
	fileName = fileName[:len(fileName)-3] + ".mp4"
	savePath := filepath.Join(c.folderPath, fileName)
	
	// 檢查是否已下載
	if _, err := os.Stat(savePath); err == nil {
		c.updateProgress(url, true)
		return
	}
	
	// 下載
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("\n建立請求失敗 %s: %v\n", fileName, err)
		return
	}
	
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("\n下載失敗 %s: %v\n", fileName, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("\n下載失敗 %s: status code %d\n", fileName, resp.StatusCode)
		return
	}
	
	// 讀取內容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("\n讀取失敗 %s: %v\n", fileName, err)
		return
	}
	
	// 解密
	if c.cipher != nil {
		decrypted := make([]byte, len(content))
		c.cipher.CryptBlocks(decrypted, content)
		content = decrypted
	}
	
	// 寫入檔案
	file, err := os.Create(savePath)
	if err != nil {
		fmt.Printf("\n建立檔案失敗 %s: %v\n", fileName, err)
		return
	}
	defer file.Close()
	
	if _, err := file.Write(content); err != nil {
		fmt.Printf("\n寫入檔案失敗 %s: %v\n", fileName, err)
		return
	}
	
	c.updateProgress(url, false)
}

func (c *Crawler) updateProgress(url string, skipped bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.progress++
	remaining := c.total - c.progress
	
	fileName := filepath.Base(url)
	if skipped {
		fmt.Printf("\r當前目標: %s 已下載, 故跳過...剩餘 %d 個", fileName, remaining)
	} else {
		fmt.Printf("\r當前下載: %s, 剩餘 %d 個", fileName, remaining)
	}
}
