package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func GetMovieLinks(url string) ([]string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	var htmlContent string
	
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	
	if err != nil {
		return nil, err
	}
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}
	
	var links []string
	doc.Find("div.img-box a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})
	
	fmt.Printf("獲取到 %d 個影片\n", len(links))
	return links, nil
}
