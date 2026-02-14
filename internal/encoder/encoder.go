package encoder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type EncodeMode int

const (
	NoEncode EncodeMode = iota
	FastEncode
	GPUEncode
	CPUEncode
)

func FFmpegEncode(folderPath, fileName string, mode EncodeMode) error {
	if mode == NoEncode {
		return nil
	}
	
	originalPath := filepath.Join(folderPath, fileName+".mp4")
	tempPath := filepath.Join(folderPath, "f_"+fileName+".mp4")
	
	var cmd *exec.Cmd
	
	switch mode {
	case FastEncode:
		// 快速無損轉檔
		cmd = exec.Command("ffmpeg",
			"-i", originalPath,
			"-c", "copy",
			"-bsf:a", "aac_adtstoasc",
			"-movflags", "+faststart",
			tempPath,
		)
	case GPUEncode:
		// NVIDIA GPU 轉檔
		cmd = exec.Command("ffmpeg",
			"-i", originalPath,
			"-c:v", "h264_nvenc",
			"-b:v", "10000K",
			"-threads", "5",
			tempPath,
		)
	case CPUEncode:
		// CPU 轉檔
		cmd = exec.Command("ffmpeg",
			"-i", originalPath,
			"-c:v", "libx264",
			"-b:v", "3M",
			"-threads", "5",
			"-preset", "superfast",
			tempPath,
		)
	default:
		return fmt.Errorf("不支援的轉檔模式")
	}
	
	fmt.Printf("開始轉檔 (模式: %d)...\n", mode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("轉檔失敗: %v", err)
	}
	
	// 刪除原始檔案並重命名
	if err := os.Remove(originalPath); err != nil {
		return fmt.Errorf("無法刪除原始檔案: %v", err)
	}
	
	if err := os.Rename(tempPath, originalPath); err != nil {
		return fmt.Errorf("無法重命名檔案: %v", err)
	}
	
	fmt.Println("轉檔成功!")
	return nil
}
