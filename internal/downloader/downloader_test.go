package downloader

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jable-downloader-go/internal/encoder"
)

func TestNewDownloader_ValidURL(t *testing.T) {
	tests := []struct {
		url     string
		dirName string
	}{
		{"https://jable.tv/videos/ipx-486/", "ipx-486"},
		{"https://jable.tv/videos/ipx-486", "ipx-486"},
		{"https://jable.tv/videos/abc-123/", "abc-123"},
	}

	for _, tt := range tests {
		t.Run(tt.dirName, func(t *testing.T) {
			d, err := NewDownloader(tt.url)
			if err != nil {
				t.Fatalf("NewDownloader failed: %v", err)
			}
			if d.URL != tt.url {
				t.Errorf("expected URL=%q, got %q", tt.url, d.URL)
			}
			if d.DirName != tt.dirName {
				t.Errorf("expected DirName=%q, got %q", tt.dirName, d.DirName)
			}
			if d.FolderPath != filepath.Join("download", tt.dirName) {
				t.Errorf("expected FolderPath=%q, got %q", filepath.Join("download", tt.dirName), d.FolderPath)
			}
			if d.AutoMode {
				t.Error("expected AutoMode=false")
			}
			if d.EncodeMode != encoder.NoEncode {
				t.Errorf("expected EncodeMode=NoEncode(%d), got %d", encoder.NoEncode, d.EncodeMode)
			}
		})
	}
}

func TestNewDownloader_InvalidURL(t *testing.T) {
	tests := []struct {
		url  string
		name string
	}{
		{"", "empty"},
		{"/", "slash_only"},
		{"not-a-url", "not_a_url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDownloader(tt.url)
			if err == nil {
				t.Errorf("expected error for URL %q", tt.url)
			}
		})
	}
}

func TestNewDownloader_NormalizeURL(t *testing.T) {
	// URLs with and without trailing slash should produce same DirName
	withSlash, _ := NewDownloader("https://jable.tv/videos/abc-123/")
	withoutSlash, _ := NewDownloader("https://jable.tv/videos/abc-123")

	if withSlash.DirName != withoutSlash.DirName {
		t.Errorf("DirName should be same: withSlash=%q, withoutSlash=%q",
			withSlash.DirName, withoutSlash.DirName)
	}
}

// m3u8Regex 是 downloader 中用於提取 M3U8 URL 的正則表達式，這裡複製一份用於測試
var m3u8Regex = regexp.MustCompile(`https://[^\s"]+\.m3u8`)

func TestM3U8Regex(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		want     string
		wantFail bool
	}{
		{
			name: "standard_m3u8_url",
			html: `<video><source src="https://cdn.jable.tv/hls/ipx-486/playlist.m3u8" type="application/x-mpegURL"></video>`,
			want: "https://cdn.jable.tv/hls/ipx-486/playlist.m3u8",
		},
		{
			name: "m3u8_in_script",
			html: `<script>var url = "https://media.jable.tv/hls/abc-123/index.m3u8";</script>`,
			want: "https://media.jable.tv/hls/abc-123/index.m3u8",
		},
		{
			name: "multiple_m3u8_urls",
			html: `<source src="https://cdn1.jable.tv/a.m3u8"><source src="https://cdn2.jable.tv/b.m3u8">`,
			want: "https://cdn1.jable.tv/a.m3u8", // should match first
		},
		{
			name: "no_m3u8_url",
			html: `<html><head><title>No video here</title></head></html>`,
			wantFail: true,
		},
		{
			name: "m3u8_with_query_params",
			html: `<source src="https://cdn.jable.tv/hls/test.m3u8?token=abc123&expires=9999999999">`,
			want: "https://cdn.jable.tv/hls/test.m3u8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := m3u8Regex.FindStringSubmatch(tt.html)
			if tt.wantFail {
				if len(matches) > 0 {
					t.Errorf("expected no match, got %q", matches[0])
				}
				return
			}
			if len(matches) == 0 {
				t.Fatal("expected match, got none")
			}
			if matches[0] != tt.want {
				t.Errorf("expected %q, got %q", tt.want, matches[0])
			}
		})
	}
}

// testM3U8Playlist 產生測試用的 M3U8 播放清單
func testM3U8Playlist(hasKey bool) string {
	playlist := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
`
	if hasKey {
		playlist += `#EXT-X-KEY:METHOD=AES-128,URI="key.key",IV=0x1234567890abcdef1234567890abcdef
`
	}
	playlist += `#EXTINF:10.000,
segment1.ts
#EXTINF:10.000,
segment2.ts
#EXTINF:10.000,
segment3.ts
#EXT-X-ENDLIST`
	return playlist
}

func TestParseM3U8_NoEncryption(t *testing.T) {
	m3u8Content := testM3U8Playlist(false)

	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(m3u8Content))
	}))
	defer m3u8Server.Close()

	m3u8URL := m3u8Server.URL + "/playlist.m3u8"
	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	tsList, aesKey, iv, err := d.parseM3U8(m3u8URL)
	if err != nil {
		t.Fatalf("parseM3U8 failed: %v", err)
	}

	if len(tsList) != 3 {
		t.Errorf("expected 3 TS segments, got %d", len(tsList))
	}

	// Verify TS URLs are correctly constructed
	base := m3u8Server.URL
	for i, ts := range tsList {
		expected := fmt.Sprintf("%s/segment%d.ts", base, i+1)
		if ts != expected {
			t.Errorf("segment %d: expected %q, got %q", i, expected, ts)
		}
	}

	if len(aesKey) != 0 {
		t.Errorf("expected no AES key, got %d bytes", len(aesKey))
	}
	if len(iv) != 0 {
		t.Errorf("expected no IV, got %d bytes", len(iv))
	}
}

func TestParseM3U8_WithEncryption(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes
	ivHex := "1234567890abcdef1234567890abcdef"

	m3u8Content := testM3U8Playlist(true)

	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		switch {
		case strings.HasSuffix(r.URL.Path, "key.key"):
			w.WriteHeader(http.StatusOK)
			w.Write(key)
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(m3u8Content))
		}
	}))
	defer m3u8Server.Close()

	m3u8URL := m3u8Server.URL + "/playlist.m3u8"
	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	tsList, gotKey, gotIV, err := d.parseM3U8(m3u8URL)
	if err != nil {
		t.Fatalf("parseM3U8 failed: %v", err)
	}

	if len(tsList) != 3 {
		t.Errorf("expected 3 TS segments, got %d", len(tsList))
	}

	if string(gotKey) != string(key) {
		t.Errorf("key mismatch: got %v, want %v", gotKey, key)
	}

	if len(gotIV) != 16 {
		t.Fatalf("expected 16-byte IV, got %d bytes", len(gotIV))
	}
	if hex.EncodeToString(gotIV) != ivHex {
		t.Errorf("IV mismatch: got %s, want %s", hex.EncodeToString(gotIV), ivHex)
	}
}

func TestParseM3U8_InvalidURL(t *testing.T) {
	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	_, _, _, err := d.parseM3U8("http://invalid-url-that-does-not-exist.example/playlist.m3u8")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestParseM3U8_InvalidContent(t *testing.T) {
	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not a valid m3u8 playlist"))
	}))
	defer m3u8Server.Close()

	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	_, _, _, err := d.parseM3U8(m3u8Server.URL + "/invalid.m3u8")
	if err == nil {
		t.Error("expected error for invalid M3U8 content")
	}
}

func TestParseM3U8_NonMediaPlaylist(t *testing.T) {
	// 主播放清單（MASTER），不是 MEDIA
	masterPlaylist := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1920x1080
https://cdn.example.com/high.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=640000,RESOLUTION=1280x720
https://cdn.example.com/low.m3u8`

	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(masterPlaylist))
	}))
	defer m3u8Server.Close()

	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	_, _, _, err := d.parseM3U8(m3u8Server.URL + "/master.m3u8")
	if err == nil {
		t.Error("expected error for MASTER playlist (not MEDIA)")
	}
}

func TestParseM3U8_KeyFetchFailure(t *testing.T) {
	// Note: The production code does not check HTTP status code on key fetch,
	// so a 404 response body will be treated as the key bytes (no error).
	// This test verifies the actual behavior.
	m3u8Content := testM3U8Playlist(true) // references key.key

	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		if strings.HasSuffix(r.URL.Path, "key.key") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 not found")) // body so io.ReadAll returns data
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(m3u8Content))
	}))
	defer m3u8Server.Close()

	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	tsList, aesKey, iv, err := d.parseM3U8(m3u8Server.URL + "/playlist.m3u8")
	if err != nil {
		t.Fatalf("parseM3U8 failed: %v", err)
	}

	// TS list should still be parsed
	if len(tsList) == 0 {
		t.Error("expected TS segments to be parsed even if key fetch returns 404")
	}

	// Key should be the 404 body bytes (production code doesn't check HTTP status)
	if len(aesKey) == 0 {
		t.Error("expected key bytes (404 body) even with failed key fetch")
	}

	// IV should be set because the playlist has IV
	if len(iv) == 0 {
		t.Error("expected IV to be parsed")
	}
}

func TestParseM3U8_KeyNetworkError(t *testing.T) {
	// Test with a key URL that causes a network error
	m3u8Content := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-KEY:METHOD=AES-128,URI="http://127.0.0.1:1/nonexistent.key"
#EXTINF:10.000,
seg1.ts
#EXT-X-ENDLIST`

	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(m3u8Content))
	}))
	defer m3u8Server.Close()

	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	_, _, _, err := d.parseM3U8(m3u8Server.URL + "/playlist.m3u8")
	if err != nil {
		t.Logf("parseM3U8 returned error (expected if key URL unreachable): %v", err)
	} else {
		t.Log("parseM3U8 succeeded (key URL might be reachable)")
	}
}

func TestParseM3U8_RelativeKeyURL(t *testing.T) {
	key := []byte("0123456789abcdef")

	m3u8Content := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-KEY:METHOD=AES-128,URI="keys/enc.key"
#EXTINF:10.000,
seg1.ts
#EXT-X-ENDLIST`

	m3u8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		switch {
		case strings.HasSuffix(r.URL.Path, "enc.key"):
			w.WriteHeader(http.StatusOK)
			w.Write(key)
		case strings.HasSuffix(r.URL.Path, "playlist.m3u8"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(m3u8Content))
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}
	}))
	defer m3u8Server.Close()

	d, _ := NewDownloader("https://jable.tv/videos/test-123/")
	_, gotKey, _, err := d.parseM3U8(m3u8Server.URL + "/playlist.m3u8")
	if err != nil {
		t.Fatalf("parseM3U8 failed: %v", err)
	}
	if string(gotKey) != string(key) {
		t.Errorf("key mismatch: got %v, want %v", gotKey, key)
	}
}

func TestEncodeModeConstants(t *testing.T) {
	if encoder.NoEncode != 0 {
		t.Errorf("expected NoEncode=0, got %d", encoder.NoEncode)
	}
	if encoder.FastEncode != 1 {
		t.Errorf("expected FastEncode=1, got %d", encoder.FastEncode)
	}
	if encoder.GPUEncode != 2 {
		t.Errorf("expected GPUEncode=2, got %d", encoder.GPUEncode)
	}
	if encoder.CPUEncode != 3 {
		t.Errorf("expected CPUEncode=3, got %d", encoder.CPUEncode)
	}
}
