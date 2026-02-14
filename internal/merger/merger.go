package merger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func MergeTSFiles(folderPath string, tsList []string) error {
	startTime := time.Now()
	fmt.Println("開始合成影片..")
	
	videoName := filepath.Base(folderPath)
	outputPath := filepath.Join(folderPath, videoName+".mp4")
	
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("無法建立輸出檔案: %v", err)
	}
	defer output.Close()
	
	for i, tsURL := range tsList {
		fileName := filepath.Base(tsURL)
		fileName = fileName[:len(fileName)-3] + ".mp4"
		fullPath := filepath.Join(folderPath, fileName)
		
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("%s 失敗\n", fileName)
			continue
		}
		
		input, err := os.Open(fullPath)
		if err != nil {
			fmt.Printf("無法開啟 %s: %v\n", fileName, err)
			continue
		}
		
		_, err = io.Copy(output, input)
		input.Close()
		
		if err != nil {
			fmt.Printf("合併 %s 失敗: %v\n", fileName, err)
			continue
		}
		
		if (i+1)%100 == 0 {
			fmt.Printf("已合併 %d/%d 個片段\n", i+1, len(tsList))
		}
	}
	
	elapsed := time.Since(startTime)
	fmt.Printf("花費 %.2f 秒合成影片\n", elapsed.Seconds())
	fmt.Println("下載完成!")
	
	return nil
}
