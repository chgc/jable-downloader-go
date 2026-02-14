package utils

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// IsRunningInContainer 檢測是否在容器環境中運行
func IsRunningInContainer() bool {
	// 方法 1: 檢查 /.dockerenv 文件
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	// 方法 2: 檢查 /proc/1/cgroup (Linux only)
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") || strings.Contains(content, "kubepods") {
			return true
		}
	}
	
	// 方法 3: 檢查環境變量
	if os.Getenv("DOCKER_CONTAINER") == "true" || os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}
	
	return false
}

func GetRandomRecommendation() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://jable.tv/", nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	
	var urls []string
	doc.Find("h6.title a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			urls = append(urls, href)
		}
	})
	
	if len(urls) == 0 {
		return "", errors.New("找不到推薦影片")
	}
	
	rand.Seed(time.Now().UnixNano())
	return urls[rand.Intn(len(urls))], nil
}

func DownloadCover(htmlContent, folderPath string) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return err
	}
	
	var coverURL string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if content, exists := s.Attr("content"); exists {
			if strings.Contains(content, "preview.jpg") {
				coverURL = content
			}
		}
	})
	
	if coverURL == "" {
		return errors.New("找不到封面圖片")
	}
	
	resp, err := http.Get(coverURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	coverName := filepath.Base(folderPath) + ".jpg"
	coverPath := filepath.Join(folderPath, coverName)
	
	file, err := os.Create(coverPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	
	fmt.Printf("封面已下載: %s\n", coverName)
	return nil
}

func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DeleteFiles(folderPath string, except string) error {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if file.Name() != except && !file.IsDir() {
			os.Remove(filepath.Join(folderPath, file.Name()))
		}
	}
	return nil
}
