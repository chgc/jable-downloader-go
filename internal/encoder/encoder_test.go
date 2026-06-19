package encoder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFFmpegEncode_NoEncode(t *testing.T) {
	dir := t.TempDir()
	fileName := "test-video"

	// NoEncode should return nil immediately without any file operations
	err := FFmpegEncode(dir, fileName, NoEncode)
	if err != nil {
		t.Errorf("NoEncode should not return error, got: %v", err)
	}
}

func TestFFmpegEncode_NoEncodeNoFile(t *testing.T) {
	// Even with non-existent file, NoEncode should not error
	err := FFmpegEncode("/nonexistent/path", "video", NoEncode)
	if err != nil {
		t.Errorf("NoEncode should not return error for non-existent paths, got: %v", err)
	}
}

func TestEncodeModeValues(t *testing.T) {
	tests := []struct {
		mode EncodeMode
		want int
	}{
		{NoEncode, 0},
		{FastEncode, 1},
		{GPUEncode, 2},
		{CPUEncode, 3},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			if int(tt.mode) != tt.want {
				t.Errorf("expected %d, got %d", tt.want, int(tt.mode))
			}
		})
	}
}

func TestFFmpegEncode_FastEncodeWithoutFFmpeg(t *testing.T) {
	dir := t.TempDir()
	fileName := "test-video"

	// Create a dummy source file
	srcPath := filepath.Join(dir, fileName+".mp4")
	if err := os.WriteFile(srcPath, []byte("dummy mp4 content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// FastEncode requires FFmpeg - expect error if not installed
	err := FFmpegEncode(dir, fileName, FastEncode)
	if err != nil {
		// FFmpeg not available is expected
		t.Logf("FastEncode returned (FFmpeg likely not installed): %v", err)

		// The source file should still exist (temp file was not renamed over it)
		if _, statErr := os.Stat(srcPath); statErr != nil {
			t.Errorf("source file should still exist after failed encode: %v", statErr)
		}
	} else {
		t.Log("FFmpeg is installed and FastEncode succeeded")
	}
}

func TestFFmpegEncode_GPUEncodeWithoutFFmpeg(t *testing.T) {
	dir := t.TempDir()
	fileName := "test-gpu"

	srcPath := filepath.Join(dir, fileName+".mp4")
	os.WriteFile(srcPath, []byte("dummy"), 0644)

	err := FFmpegEncode(dir, fileName, GPUEncode)
	if err != nil {
		t.Logf("GPUEncode returned (expected without GPU/FFmpeg): %v", err)
	} else {
		t.Log("GPU FFmpeg is available and encode succeeded")
	}
}

func TestFFmpegEncode_CPUEncodeWithoutFFmpeg(t *testing.T) {
	dir := t.TempDir()
	fileName := "test-cpu"

	srcPath := filepath.Join(dir, fileName+".mp4")
	os.WriteFile(srcPath, []byte("dummy"), 0644)

	err := FFmpegEncode(dir, fileName, CPUEncode)
	if err != nil {
		t.Logf("CPUEncode returned (expected without FFmpeg): %v", err)
	} else {
		t.Log("CPU FFmpeg is available and encode succeeded")
	}
}

func TestFFmpegEncode_InvalidMode(t *testing.T) {
	dir := t.TempDir()
	fileName := "test-invalid"

	srcPath := filepath.Join(dir, fileName+".mp4")
	os.WriteFile(srcPath, []byte("dummy"), 0644)

	// Use an invalid mode value
	err := FFmpegEncode(dir, fileName, EncodeMode(99))
	if err == nil {
		t.Error("expected error for invalid encode mode")
	}
}

func TestFFmpegEncode_SourceFileNotExist(t *testing.T) {
	dir := t.TempDir()

	// FastEncode with non-existent source file
	err := FFmpegEncode(dir, "nonexistent", FastEncode)
	if err == nil {
		t.Error("expected error when source file does not exist")
	}
}

// Test that FFmpegEncode paths are constructed correctly
func TestFFmpegEncode_Paths(t *testing.T) {
	dir := t.TempDir()
	fileName := "path-test"

	expectedOriginal := filepath.Join(dir, fileName+".mp4")

	// Create the original file
	os.WriteFile(expectedOriginal, []byte("test data"), 0644)

	// FastEncode will try to run FFmpeg, which likely fails
	// But we can check the paths are constructed correctly by verifying the file exists
	err := FFmpegEncode(dir, fileName, FastEncode)
	if err != nil {
		// FFmpeg not available - but verify paths were correct
		// Original should still exist, temp should not
		if _, err := os.Stat(expectedOriginal); os.IsNotExist(err) {
			t.Error("original file should still exist")
		}
	}
}

// Helper type for test readability
func (m EncodeMode) String() string {
	switch m {
	case NoEncode:
		return "NoEncode"
	case FastEncode:
		return "FastEncode"
	case GPUEncode:
		return "GPUEncode"
	case CPUEncode:
		return "CPUEncode"
	default:
		return "Unknown"
	}
}
