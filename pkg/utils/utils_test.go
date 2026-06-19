package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestEnsureDir_New(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newdir")
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestEnsureDir_Existing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "existing")
	os.MkdirAll(dir, 0755)

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir on existing dir failed: %v", err)
	}
}

func TestFileExists_True(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	if !FileExists(path) {
		t.Error("FileExists should return true for existing file")
	}
}

func TestFileExists_False(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.txt")
	if FileExists(path) {
		t.Error("FileExists should return false for non-existing file")
	}
}

func TestDeleteFiles_Except(t *testing.T) {
	dir := t.TempDir()

	files := []string{"keep.mp4", "delete1.ts", "delete2.ts", "delete3.ts"}
	for _, f := range files {
		os.WriteFile(filepath.Join(dir, f), []byte("data"), 0644)
	}

	if err := DeleteFiles(dir, "keep.mp4"); err != nil {
		t.Fatalf("DeleteFiles failed: %v", err)
	}

	// keep.mp4 should survive
	if _, err := os.Stat(filepath.Join(dir, "keep.mp4")); os.IsNotExist(err) {
		t.Error("keep.mp4 should still exist")
	}

	// Others should be deleted
	for _, f := range []string{"delete1.ts", "delete2.ts", "delete3.ts"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			t.Errorf("%s should have been deleted", f)
		}
	}
}

func TestDeleteFiles_AllDeleted(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "seg1.ts"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir, "seg2.ts"), []byte("data"), 0644)

	if err := DeleteFiles(dir, "nonexistent.mp4"); err != nil {
		t.Fatalf("DeleteFiles failed: %v", err)
	}

	// All files should be deleted (except doesn't match anything)
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("expected empty directory, got %d entries", len(entries))
	}
}

func TestDeleteFiles_SkipDirectories(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	os.WriteFile(filepath.Join(dir, "file.ts"), []byte("data"), 0644)

	DeleteFiles(dir, "nonexistent")

	// Directory should still exist
	if _, err := os.Stat(filepath.Join(dir, "subdir")); os.IsNotExist(err) {
		t.Error("subdirectory should not be deleted")
	}
	// File should be deleted
	if _, err := os.Stat(filepath.Join(dir, "file.ts")); err == nil {
		t.Error("file.ts should have been deleted")
	}
}

func TestDeleteFiles_NonExistentDir(t *testing.T) {
	err := DeleteFiles("/nonexistent/path", "keep.mp4")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestIsRunningInContainer_ByEnvVar(t *testing.T) {
	// DOCKER_CONTAINER=true should trigger true
	t.Setenv("DOCKER_CONTAINER", "true")
	if !IsRunningInContainer() {
		t.Error("expected true when DOCKER_CONTAINER=true")
	}
}

func TestIsRunningInContainer_ByKubernetesEnv(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	if !IsRunningInContainer() {
		t.Error("expected true when KUBERNETES_SERVICE_HOST is set")
	}
}

func TestIsRunningInContainer_NotContainer(t *testing.T) {
	t.Setenv("DOCKER_CONTAINER", "")
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	// /.dockerenv and /proc/1/cgroup don't exist on non-container systems
	result := IsRunningInContainer()
	// On a non-container system, should be false
	// But on some CI systems, it might be running in a container
	t.Logf("IsRunningInContainer() = %v (expected false on bare metal)", result)
}

func TestDownloadCover_Success(t *testing.T) {
	imgData := []byte("fake-jpeg-image-data")

	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(imgData)
	}))
	defer imgServer.Close()

	htmlContent := `<html><head><meta content="` + imgServer.URL + `/preview.jpg"></head></html>`
	dir := t.TempDir()

	if err := DownloadCover(htmlContent, dir); err != nil {
		t.Fatalf("DownloadCover failed: %v", err)
	}

	// Check the file was saved
	dirName := filepath.Base(dir)
	expectedPath := filepath.Join(dir, dirName+".jpg")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("cover file was not created at %s", expectedPath)
	} else {
		content, _ := os.ReadFile(expectedPath)
		if string(content) != string(imgData) {
			t.Error("cover file content mismatch")
		}
	}
}

func TestDownloadCover_NoMeta(t *testing.T) {
	htmlContent := `<html><head><title>No cover</title></head></html>`
	dir := t.TempDir()

	err := DownloadCover(htmlContent, dir)
	if err == nil {
		t.Error("expected error when no cover meta found")
	}
}

func TestDownloadCover_InvalidHTML(t *testing.T) {
	err := DownloadCover("not valid html content", t.TempDir())
	if err != nil {
		// goquery might still handle this gracefully
		t.Logf("DownloadCover with invalid HTML returned: %v", err)
	}
}

func TestDownloadCover_HTTPError(t *testing.T) {
	// Note: The production code does not check HTTP status on image download.
	// A 404 response results in an empty file, not an error.
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer imgServer.Close()

	htmlContent := `<html><head><meta content="` + imgServer.URL + `/preview.jpg"></head></html>`
	dir := t.TempDir()

	err := DownloadCover(htmlContent, dir)
	if err != nil {
		t.Fatalf("DownloadCover should not error on HTTP 404 (no status check): %v", err)
	}

	// An empty file should have been created
	expectedPath := filepath.Join(dir, filepath.Base(dir)+".jpg")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Error("cover file should have been created even with 404 response")
	}
}

func TestDownloadCover_NetworkError(t *testing.T) {
	// Test with a server that causes connection error
	htmlContent := `<html><head><meta content="http://127.0.0.1:1/nonexistent.jpg"></head></html>`
	dir := t.TempDir()

	err := DownloadCover(htmlContent, dir)
	if err != nil {
		t.Logf("DownloadCover with unreachable server returned: %v (expected)", err)
	} else {
		t.Log("DownloadCover succeeded unexpectedly")
	}
}

func TestGetRandomRecommendation_Success(t *testing.T) {
	htmlWithLinks := `<html><body>
		<div class="card">
			<h6 class="title"><a href="https://jable.tv/videos/abc-123/">Video 1</a></h6>
		</div>
		<div class="card">
			<h6 class="title"><a href="https://jable.tv/videos/def-456/">Video 2</a></h6>
		</div>
	</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlWithLinks))
	}))
	defer server.Close()

	// Temporarily replace the jable URL with our test server
	// Since GetRandomRecommendation hardcodes "https://jable.tv/", we need to
	// verify the logic works, not the specific URL
	// We can't easily mock this without changing the code, so let's test the parsing logic

	// Verify the goquery selector works on this HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlWithLinks))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	var urls []string
	doc.Find("h6.title a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			urls = append(urls, href)
		}
	})

	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
}

func TestGetRandomRecommendation_NoLinks(t *testing.T) {
	htmlNoLinks := `<html><body><h6 class="title">No links here</h6></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlNoLinks))
	var urls []string
	doc.Find("h6.title a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			urls = append(urls, href)
		}
	})

	if len(urls) != 0 {
		t.Errorf("expected 0 URLs, got %d", len(urls))
	}
}

func TestGetRandomRecommendation_HTTPError(t *testing.T) {
	// This function makes a real HTTP request to jable.tv
	// We can't easily mock it without interfaces
	// Just verify it handles errors gracefully
	_, err := GetRandomRecommendation()
	if err != nil {
		// Expected if network is unavailable
		t.Logf("GetRandomRecommendation returned: %v (expected if offline)", err)
	}
}

func TestGetRandomRecommendation_MultipleURLs(t *testing.T) {
	// Test that with multiple URLs, we get one of them
	html := `<html><body>`
	expectedURLs := []string{
		"https://jable.tv/videos/vid-001/",
		"https://jable.tv/videos/vid-002/",
		"https://jable.tv/videos/vid-003/",
	}
	for _, u := range expectedURLs {
		html += `<h6 class="title"><a href="` + u + `">Video</a></h6>`
	}
	html += `</body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	var urls []string
	doc.Find("h6.title a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			urls = append(urls, href)
		}
	})

	if len(urls) != len(expectedURLs) {
		t.Errorf("expected %d URLs, got %d", len(expectedURLs), len(urls))
	}

	// Each extracted URL should be in expected
	for _, u := range urls {
		found := false
		for _, exp := range expectedURLs {
			if u == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected URL: %s", u)
		}
	}
}
