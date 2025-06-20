package nextgenimage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestNewConverter(t *testing.T) {
	// Test default values
	converter := NewConverter(ConverterConfig{})

	if converter.config.JPEGToWebP.Quality != 80 {
		t.Errorf("Expected default JPEG to WebP quality to be 80, got %d", converter.config.JPEGToWebP.Quality)
	}

	if converter.config.JPEGToAVIF.CQ != 25 {
		t.Errorf("Expected default JPEG to AVIF CQ to be 25, got %d", converter.config.JPEGToAVIF.CQ)
	}

	if converter.config.PNGToWebP.TryNearLossless {
		t.Error("Expected default PNG to WebP TryNearLossless to be false")
	}

	// Test custom values
	config := ConverterConfig{}
	config.JPEGToWebP.Quality = 90
	config.JPEGToAVIF.CQ = 30
	config.PNGToWebP.TryNearLossless = true

	converter2 := NewConverter(config)
	if converter2.config.JPEGToWebP.Quality != 90 {
		t.Errorf("Expected JPEG to WebP quality to be 90, got %d", converter2.config.JPEGToWebP.Quality)
	}

	if converter2.config.JPEGToAVIF.CQ != 30 {
		t.Errorf("Expected JPEG to AVIF CQ to be 30, got %d", converter2.config.JPEGToAVIF.CQ)
	}

	if !converter2.config.PNGToWebP.TryNearLossless {
		t.Error("Expected PNG to WebP TryNearLossless to be true")
	}
}

func TestFormatError(t *testing.T) {
	originalErr := errors.New("test error")
	formatErr := NewFormatError(originalErr)

	// Test Error() method
	expectedMsg := "format error: test error"
	if formatErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, formatErr.Error())
	}

	// Test Unwrap() method
	if !errors.Is(formatErr, originalErr) {
		t.Error("FormatError does not wrap the original error correctly")
	}

	// Test errors.As
	var fe *FormatError
	if !errors.As(formatErr, &fe) {
		t.Error("errors.As failed to identify FormatError")
	}

	// Test errors.Is
	if !errors.Is(formatErr, originalErr) {
		t.Error("errors.Is failed to identify wrapped error")
	}
}

func TestToWebPInvalidInput(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test non-existent file
	err := converter.ToWebP("non_existent_file.jpg", filepath.Join(tempDir, "output.webp"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test unsupported format
	txtPath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(txtPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = converter.ToWebP(txtPath, filepath.Join(tempDir, "output.webp"))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestToAVIFInvalidInput(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test non-existent file
	err := converter.ToAVIF("non_existent_file.jpg", filepath.Join(tempDir, "output.avif"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test unsupported format
	txtPath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(txtPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = converter.ToAVIF(txtPath, filepath.Join(tempDir, "output.avif"))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestLargeOutputHandling(t *testing.T) {
	tempDir := t.TempDir()

	// Create a very small test image
	smallImage, err := vips.Black(10, 10)
	if err != nil {
		t.Skip("Failed to create test image")
		return
	}
	defer smallImage.Close()

	smallPath := filepath.Join(tempDir, "small.jpg")
	params := vips.NewJpegExportParams()
	params.Quality = 100 // High quality to make it harder to compress

	// Export to buffer first, then write to file
	jpegBuffer, _, err := smallImage.ExportJpeg(params)
	if err != nil {
		t.Fatalf("Failed to export small test image: %v", err)
	}

	err = os.WriteFile(smallPath, jpegBuffer, 0644)
	if err != nil {
		t.Fatalf("Failed to write small test image: %v", err)
	}

	// Try to convert with very low quality to ensure output is smaller
	config := ConverterConfig{}
	config.JPEGToWebP.Quality = 1 // Very low quality
	converter2 := NewConverter(config)

	outputPath := filepath.Join(tempDir, "small.webp")
	err = converter2.ToWebP(smallPath, outputPath)

	// This might still succeed if WebP is more efficient
	if err != nil {
		var formatErr *FormatError
		if errors.As(err, &formatErr) {
			t.Logf("Got expected format error: %v", err)
		} else {
			t.Errorf("Unexpected error type: %v", err)
		}
	}
}

func TestConcurrentConversions(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Run multiple conversions concurrently
	inputPath := "testdata/test_original.jpg"

	// Check if test file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		t.Skip("Test file not found")
		return
	}

	done := make(chan error, 4)

	// WebP conversions
	go func() {
		err := converter.ToWebP(inputPath, filepath.Join(tempDir, "concurrent1.webp"))
		done <- err
	}()

	go func() {
		err := converter.ToWebP(inputPath, filepath.Join(tempDir, "concurrent2.webp"))
		done <- err
	}()

	// AVIF conversions
	go func() {
		err := converter.ToAVIF(inputPath, filepath.Join(tempDir, "concurrent1.avif"))
		done <- err
	}()

	go func() {
		err := converter.ToAVIF(inputPath, filepath.Join(tempDir, "concurrent2.avif"))
		done <- err
	}()

	// Wait for all conversions
	for i := 0; i < 4; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent conversion failed: %v", err)
		}
	}
}
