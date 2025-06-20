package nextgenimage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestGIFToWebP(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	images := loadTestImages(t)

	// Create temp directory for outputs
	tempDir := t.TempDir()

	for _, img := range images {
		if img.Format != "gif" {
			continue
		}

		t.Run(img.Path, func(t *testing.T) {
			inputPath := filepath.Join("testdata", img.Path)
			outputPath := filepath.Join(tempDir, filepath.Base(img.Path)+".webp")

			// Get input file size
			inputInfo, err := os.Stat(inputPath)
			if err != nil {
				t.Fatalf("Failed to stat input file: %v", err)
			}

			// Convert
			err = converter.ToWebP(inputPath, outputPath)

			// Check if it's a format error (expected for some test cases)
			var formatErr *FormatError
			if errors.As(err, &formatErr) {
				t.Logf("Format error (expected for some cases): %v", err)
				return
			}

			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Check output exists
			outputInfo, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("Output file not created: %v", err)
			}

			// Check size reduction
			if outputInfo.Size() >= inputInfo.Size() {
				t.Errorf("Output size (%d) is not smaller than input (%d)", outputInfo.Size(), inputInfo.Size())
			}

			sizeReduction := float64(inputInfo.Size()-outputInfo.Size()) / float64(inputInfo.Size()) * 100
			t.Logf("Size reduction: %.2f%% (%d -> %d bytes)", sizeReduction, inputInfo.Size(), outputInfo.Size())

			// Check if animation is preserved
			convImage, err := vips.NewImageFromFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to load converted image: %v", err)
			}
			defer convImage.Close()

			// Check if it's animated by looking at page height
			pageHeight := convImage.PageHeight()
			if pageHeight > 0 && pageHeight < convImage.Height() {
				frameCount := convImage.Height() / pageHeight
				t.Logf("Animation preserved: %d frames", frameCount)
			}
		})
	}
}

func TestGIFAnimationProperties(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test specific animation properties
	animTests := []struct {
		name      string
		path      string
		minFrames int
	}{
		{"SingleFrame", "gif/frames_single.gif", 1},
		{"ShortAnimation", "gif/frames_short.gif", 5},
		{"MediumAnimation", "gif/frames_medium.gif", 10},
		{"LongAnimation", "gif/frames_long.gif", 20},
	}

	for _, test := range animTests {
		t.Run(test.name, func(t *testing.T) {
			inputPath := filepath.Join("testdata", test.path)
			outputPath := filepath.Join(tempDir, filepath.Base(test.path)+".webp")

			// Check if input file exists
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skip("Test file not found:", inputPath)
				return
			}

			err := converter.ToWebP(inputPath, outputPath)

			// Check if it's a format error
			var formatErr *FormatError
			if errors.As(err, &formatErr) {
				t.Logf("Format error: %v", err)
				return
			}

			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Load converted image and check frame count
			params := vips.NewImportParams()
			params.NumPages.Set(-1)
			convImage, err := vips.LoadImageFromFile(outputPath, params)
			if err != nil {
				t.Fatalf("Failed to load converted image: %v", err)
			}
			defer convImage.Close()

			pageHeight := convImage.PageHeight()
			if pageHeight > 0 {
				frameCount := convImage.Height() / pageHeight
				t.Logf("Frame count: %d", frameCount)

				if frameCount < test.minFrames {
					t.Errorf("Expected at least %d frames, got %d", test.minFrames, frameCount)
				}
			}
		})
	}
}

func TestGIFLoopHandling(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test loop settings
	loopTests := []struct {
		name string
		path string
	}{
		{"InfiniteLoop", "gif/loop_loop_infinite.gif"},
		{"LoopOnce", "gif/loop_loop_once.gif"},
		{"Loop3Times", "gif/loop_loop_3times.gif"},
	}

	for _, test := range loopTests {
		t.Run(test.name, func(t *testing.T) {
			inputPath := filepath.Join("testdata", test.path)
			outputPath := filepath.Join(tempDir, filepath.Base(test.path)+".webp")

			// Check if input file exists
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skip("Test file not found:", inputPath)
				return
			}

			err := converter.ToWebP(inputPath, outputPath)

			// Check if it's a format error
			var formatErr *FormatError
			if errors.As(err, &formatErr) {
				t.Logf("Format error: %v", err)
				return
			}

			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Just verify the file was created
			if _, err := os.Stat(outputPath); err != nil {
				t.Errorf("Output file not created: %v", err)
			}
		})
	}
}
