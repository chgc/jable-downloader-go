package crawler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// testData 回傳一段可預測的測試資料（長度為 16 的倍數以符合 AES-CBC）
func testData() []byte {
	return []byte("Hello! This is test data for jable downloader!!!")
}

func pad(data []byte) []byte {
	padding := 16 - len(data)%16
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}

func unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding > 16 {
		return data
	}
	return data[:len(data)-padding]
}

func TestNewCrawler_WithoutKey(t *testing.T) {
	c, err := NewCrawler(t.TempDir(), []string{"http://example.com/seg1.ts"}, nil, nil)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}
	if c == nil {
		t.Fatal("NewCrawler returned nil")
	}
	if c.cipher != nil {
		t.Error("cipher should be nil when no key provided")
	}
	if c.total != 1 {
		t.Errorf("expected total=1, got %d", c.total)
	}
}

func TestNewCrawler_WithKey(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	rand.Read(key)
	rand.Read(iv)

	c, err := NewCrawler(t.TempDir(), []string{"http://example.com/seg1.ts"}, key, iv)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}
	if c == nil {
		t.Fatal("NewCrawler returned nil")
	}
	if c.cipher == nil {
		t.Error("cipher should not be nil when key is provided")
	}
}

func TestNewCrawler_InvalidKey(t *testing.T) {
	// AES-128 requires 16-byte key
	_, err := NewCrawler(t.TempDir(), []string{}, []byte{1, 2, 3}, make([]byte, 16))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestDownload_Success(t *testing.T) {
	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock-ts-content"))
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	c, err := NewCrawler(dir, []string{tsServer.URL + "/seg1.ts", tsServer.URL + "/seg2.ts"}, nil, nil)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}

	if err := c.Download(); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify files were created
	expectedFiles := []string{"seg1.mp4", "seg2.mp4"}
	for _, f := range expectedFiles {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", f)
		}
	}
}

func TestDownload_WithAESDecryption(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	rand.Read(key)
	rand.Read(iv)

	plaintext := []byte("encrypted-ts-content-block")
	padded := pad(plaintext)

	// Encrypt
	block, _ := aes.NewCipher(key)
	encrypter := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padded))
	encrypter.CryptBlocks(ciphertext, padded)

	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(ciphertext)
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	c, err := NewCrawler(dir, []string{tsServer.URL + "/enc.ts"}, key, iv)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}

	if err := c.Download(); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Read the decrypted file
	content, err := os.ReadFile(filepath.Join(dir, "enc.mp4"))
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	content = unpad(content)
	if string(content) != string(plaintext) {
		t.Errorf("decrypted content mismatch:\n got:  %q\n want: %q", string(content), string(plaintext))
	}
}

func TestDownload_HTTPError(t *testing.T) {
	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	c, err := NewCrawler(dir, []string{tsServer.URL + "/missing.ts"}, nil, nil)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}

	// Should not return error (errors are reported individually)
	if err := c.Download(); err != nil {
		t.Fatalf("Download should not return error on HTTP 404, got: %v", err)
	}
}

func TestDownload_SkipExisting(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content"))
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	// Create the file beforehand to trigger skip
	os.WriteFile(filepath.Join(dir, "seg1.mp4"), []byte("existing"), 0644)

	c, err := NewCrawler(dir, []string{tsServer.URL + "/seg1.ts"}, nil, nil)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}

	if err := c.Download(); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	mu.Lock()
	if callCount > 0 {
		t.Errorf("expected 0 HTTP calls (file skipped), got %d", callCount)
	}
	mu.Unlock()
}

func TestDownload_MultipleWorkers(t *testing.T) {
	var mu sync.Mutex
	downloaded := make(map[string]bool)

	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		downloaded[r.URL.Path] = true
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content"))
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	urls := []string{
		tsServer.URL + "/seg1.ts",
		tsServer.URL + "/seg2.ts",
		tsServer.URL + "/seg3.ts",
		tsServer.URL + "/seg4.ts",
		tsServer.URL + "/seg5.ts",
	}
	c, err := NewCrawler(dir, urls, nil, nil)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}

	if err := c.Download(); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify all files created
	expectedFiles := []string{"seg1.mp4", "seg2.mp4", "seg3.mp4", "seg4.mp4", "seg5.mp4"}
	for _, f := range expectedFiles {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", f)
		}
	}
}

func TestDownload_ConcurrentSafety(t *testing.T) {
	// Test that multiple crawlers can run independently
	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content"))
	}))
	defer tsServer.Close()

	dir1 := t.TempDir()
	dir2 := t.TempDir()
	urls := []string{tsServer.URL + "/seg1.ts", tsServer.URL + "/seg2.ts"}

	c1, _ := NewCrawler(dir1, urls, nil, nil)
	c2, _ := NewCrawler(dir2, urls, nil, nil)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); c1.Download() }()
	go func() { defer wg.Done(); c2.Download() }()
	wg.Wait()

	// Both should have created files
	for _, f := range []string{"seg1.mp4", "seg2.mp4"} {
		if _, err := os.Stat(filepath.Join(dir1, f)); os.IsNotExist(err) {
			t.Errorf("c1: expected file %s", f)
		}
		if _, err := os.Stat(filepath.Join(dir2, f)); os.IsNotExist(err) {
			t.Errorf("c2: expected file %s", f)
		}
	}
}

func TestProgressTracking(t *testing.T) {
	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	urls := []string{tsServer.URL + "/a.ts", tsServer.URL + "/b.ts", tsServer.URL + "/c.ts"}
	c, err := NewCrawler(dir, urls, nil, nil)
	if err != nil {
		t.Fatalf("NewCrawler failed: %v", err)
	}

	if err := c.Download(); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// After download, progress should equal total
	if c.progress != c.total {
		t.Errorf("expected progress=%d (total=%d), got %d", c.total, c.total, c.progress)
	}
}

func TestUpdateProgress(t *testing.T) {
	c := &Crawler{
		downloadList: []string{"http://example.com/a.ts", "http://example.com/b.ts"},
		total:        2,
		progress:     0,
	}

	c.updateProgress("http://example.com/a.ts", false)
	if c.progress != 1 {
		t.Errorf("expected progress=1, got %d", c.progress)
	}

	c.updateProgress("http://example.com/b.ts", true)
	if c.progress != 2 {
		t.Errorf("expected progress=2, got %d", c.progress)
	}
}

func TestDownloadOne_WritesCorrectContent(t *testing.T) {
	expectedContent := []byte("specific-test-content-12345")

	tsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(expectedContent)
	}))
	defer tsServer.Close()

	dir := t.TempDir()
	c, _ := NewCrawler(dir, []string{tsServer.URL + "/test.ts"}, nil, nil)

	c.downloadOne(tsServer.URL + "/test.ts")

	content, err := os.ReadFile(filepath.Join(dir, "test.mp4"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if string(content) != string(expectedContent) {
		t.Errorf("content mismatch:\n got:  %q\n want: %q", string(content), string(expectedContent))
	}
}
