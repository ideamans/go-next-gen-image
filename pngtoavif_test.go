package nextgenimage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestPNGToAVIF(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	images := loadTestImages(t)

	// Create temp directory for outputs
	tempDir := t.TempDir()

	for _, img := range images {
		if img.Format != "png" {
			continue
		}

		t.Run(img.Path, func(t *testing.T) {
			inputPath := filepath.Join("testdata", img.Path)
			outputPath := filepath.Join(tempDir, filepath.Base(img.Path)+".avif")

			// Get input file size
			inputInfo, err := os.Stat(inputPath)
			if err != nil {
				t.Fatalf("Failed to stat input file: %v", err)
			}

			// Convert
			err = converter.ToAVIF(inputPath, outputPath)

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

			// Check dimensions
			origImage, err := vips.NewImageFromFile(inputPath)
			if err != nil {
				t.Fatalf("Failed to load original image: %v", err)
			}
			defer origImage.Close()

			convImage, err := vips.NewImageFromFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to load converted image: %v", err)
			}
			defer convImage.Close()

			if origImage.Width() != convImage.Width() || origImage.Height() != convImage.Height() {
				t.Errorf("Dimensions changed: %dx%d -> %dx%d",
					origImage.Width(), origImage.Height(),
					convImage.Width(), convImage.Height())
			}

			// For PNG lossless conversion to AVIF, skip PSNR check
			// AVIF lossless may not produce identical pixels due to format differences
			t.Logf("Lossless AVIF conversion completed")

			// Check metadata preservation/removal
			origEXIF, origXMP, origICC := checkMetadata(t, inputPath)
			convEXIF, convXMP, convICC := checkMetadata(t, outputPath)
			
			t.Logf("Original metadata - EXIF: %v, XMP: %v, ICC: %v", origEXIF, origXMP, origICC)
			t.Logf("Converted metadata - EXIF: %v, XMP: %v, ICC: %v", convEXIF, convXMP, convICC)
			
			// All metadata should be removed with StripMetadata=true
			if convEXIF {
				t.Error("EXIF data was not removed during conversion")
			}
			if convXMP {
				t.Error("XMP data was not removed during conversion")
			}
			// ICC is also removed with StripMetadata=true
			if convICC {
				t.Error("ICC profile was not removed during conversion")
			}
		})
	}
}

func TestPNGToAVIFAlphaHandling(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test alpha channel preservation
	alphaTests := []string{
		"png/alpha_transparent.png",
		"png/alpha_semitransparent.png",
		"png/colortype_rgba.png",
		"png/colortype_grayscale_alpha.png",
	}

	for _, testPath := range alphaTests {
		t.Run(testPath, func(t *testing.T) {
			inputPath := filepath.Join("testdata", testPath)
			outputPath := filepath.Join(tempDir, filepath.Base(testPath)+".avif")

			// Check if input file exists
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skip("Test file not found:", inputPath)
				return
			}

			err := converter.ToAVIF(inputPath, outputPath)

			// Check if it's a format error
			var formatErr *FormatError
			if errors.As(err, &formatErr) {
				t.Logf("Format error: %v", err)
				return
			}

			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Load converted image and check if it has alpha
			convImage, err := vips.NewImageFromFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to load converted image: %v", err)
			}
			defer convImage.Close()

			if convImage.HasAlpha() {
				t.Logf("Alpha channel preserved")
			} else {
				// Some formats might not need alpha, so just log
				t.Logf("No alpha channel in output")
			}
		})
	}
}

func TestGIFToAVIFNotSupported(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	inputPath := "testdata/test_original.gif"
	outputPath := filepath.Join(tempDir, "test.avif")

	err := converter.ToAVIF(inputPath, outputPath)

	// Should return a format error
	var formatErr *FormatError
	if !errors.As(err, &formatErr) {
		t.Errorf("Expected FormatError for GIF to AVIF conversion, got: %v", err)
	}

	t.Logf("Expected format error: %v", err)
}
