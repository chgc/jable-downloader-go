package parser

import (
	"flag"
	"os"
	"testing"
)

// resetFlags 重置 flag 狀態，避免測試間互相影響
func resetFlags(t *testing.T) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestParseArgs_Defaults(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader"}

	args := ParseArgs()

	if args.URL != "" {
		t.Errorf("expected empty URL, got %q", args.URL)
	}
	if args.Random {
		t.Error("expected Random=false")
	}
	if args.AllURLs != "" {
		t.Errorf("expected empty AllURLs, got %q", args.AllURLs)
	}
	if args.Server {
		t.Error("expected Server=false")
	}
	if args.Port != 18080 {
		t.Errorf("expected Port=18080, got %d", args.Port)
	}
}

func TestParseArgs_URL(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader", "--url", "https://jable.tv/videos/test-123/"}

	args := ParseArgs()

	if args.URL != "https://jable.tv/videos/test-123/" {
		t.Errorf("expected URL 'https://jable.tv/videos/test-123/', got %q", args.URL)
	}
	if args.Random {
		t.Error("expected Random=false when --url is set")
	}
	if args.Server {
		t.Error("expected Server=false when --url is set")
	}
}

func TestParseArgs_Random(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader", "--random"}

	args := ParseArgs()

	if !args.Random {
		t.Error("expected Random=true")
	}
	if args.URL != "" {
		t.Errorf("expected empty URL, got %q", args.URL)
	}
}

func TestParseArgs_AllURLs(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader", "--all-urls", "https://jable.tv/models/actress/"}

	args := ParseArgs()

	if args.AllURLs != "https://jable.tv/models/actress/" {
		t.Errorf("expected AllURLs 'https://jable.tv/models/actress/', got %q", args.AllURLs)
	}
}

func TestParseArgs_Server(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader", "--server"}

	args := ParseArgs()

	if !args.Server {
		t.Error("expected Server=true")
	}
}

func TestParseArgs_ServerWithPort(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader", "--server", "--port", "9090"}

	args := ParseArgs()

	if !args.Server {
		t.Error("expected Server=true")
	}
	if args.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", args.Port)
	}
}

func TestParseArgs_MultipleFlags(t *testing.T) {
	resetFlags(t)
	os.Args = []string{"jable-downloader", "--url", "https://jable.tv/videos/abc/", "--random"}

	args := ParseArgs()

	// When both --url and --random are set, --random should be true and --url should have the URL
	if args.URL != "https://jable.tv/videos/abc/" {
		t.Errorf("expected URL 'https://jable.tv/videos/abc/', got %q", args.URL)
	}
	if !args.Random {
		t.Error("expected Random=true")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    *Args
		wantErr bool
	}{
		{
			name:    "empty_args_interactive_mode",
			args:    &Args{URL: "", Random: false, AllURLs: "", Server: false, Port: 18080},
			wantErr: false,
		},
		{
			name:    "with_url",
			args:    &Args{URL: "https://jable.tv/videos/test/", Random: false, AllURLs: "", Server: false, Port: 18080},
			wantErr: false,
		},
		{
			name:    "with_random",
			args:    &Args{URL: "", Random: true, AllURLs: "", Server: false, Port: 18080},
			wantErr: false,
		},
		{
			name:    "with_all_urls",
			args:    &Args{URL: "", Random: false, AllURLs: "https://jable.tv/models/actress/", Server: false, Port: 18080},
			wantErr: false,
		},
		{
			name:    "server_mode",
			args:    &Args{URL: "", Random: false, AllURLs: "", Server: true, Port: 18080},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Just verify it doesn't panic
	resetFlags(t)
	PrintUsage()
}
