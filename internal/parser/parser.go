package parser

import (
	"flag"
	"fmt"
)

type Args struct {
	URL     string
	Random  bool
	AllURLs string
	Server  bool
	Port    int
}

func ParseArgs() *Args {
	args := &Args{}
	
	flag.StringVar(&args.URL, "url", "", "Jable TV URL to download")
	flag.BoolVar(&args.Random, "random", false, "Download random recommended video")
	flag.StringVar(&args.AllURLs, "all-urls", "", "Jable URL contains multiple videos")
	flag.BoolVar(&args.Server, "server", false, "Start HTTP API server mode")
	flag.IntVar(&args.Port, "port", 18080, "HTTP API server port (default: 18080)")
	
	flag.Parse()
	
	return args
}

func (a *Args) Validate() error {
	if a.URL == "" && !a.Random && a.AllURLs == "" {
		return nil // 互動模式
	}
	return nil
}

func PrintUsage() {
	fmt.Println("Jable TV Downloader - Go Version")
	fmt.Println("\n使用方式:")
	flag.PrintDefaults()
}
