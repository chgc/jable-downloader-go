package merger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func MergeTSFiles(folderPath string, tsList []string) error {
	startTime := time.Now()
	fmt.Println("開始合成影片..")

	// 建立 FFmpeg concat 清單檔
	listPath := filepath.Join(folderPath, "filelist.txt")
	var lines []string
	for _, tsURL := range tsList {
		fileName := filepath.Base(tsURL)
		fileName = fileName[:len(fileName)-3] + ".mp4"
		fullPath := filepath.Join(folderPath, fileName)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("%s 不存在，跳過\n", fileName)
			continue
		}
		// FFmpeg concat 清單路徑相對於 filelist.txt 所在目錄
		lines = append(lines, fmt.Sprintf("file '%s'", fileName))
	}

	if err := os.WriteFile(listPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("無法建立 filelist.txt: %v", err)
	}
	defer os.Remove(listPath)

	videoName := filepath.Base(folderPath)
	outputPath := filepath.Join(folderPath, videoName+".mp4")

	// 使用 FFmpeg concat demuxer 直接合成為瀏覽器可播放的 MP4
	// -c copy      無損重新封裝，速度快
	// -bsf:a aac_adtstoasc  將 TS 的 ADTS AAC 轉為 MP4 所需的 ASC 格式
	// -movflags +faststart  將 moov atom 移到檔案開頭，允許邊下載邊播放
	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-movflags", "+faststart",
		outputPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 合成失敗: %v", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("花費 %.2f 秒合成影片\n", elapsed.Seconds())
	fmt.Println("下載完成!")

	return nil
}
