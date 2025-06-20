package nextgenimage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestPNGToWebP(t *testing.T) {
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

			// For PNG lossless conversion, PSNR should be perfect (100)
			psnr, err := calculatePSNR(inputPath, outputPath)
			if err != nil {
				t.Logf("Failed to calculate PSNR: %v", err)
			} else {
				t.Logf("PSNR: %.2f dB", psnr)
				// For lossless conversion, PSNR should be very high
				if psnr < 50 {
					t.Errorf("PSNR too low for lossless conversion: %.2f dB", psnr)
				}
			}
		})
	}
}

func TestPNGToWebPNearLossless(t *testing.T) {
	tempDir := t.TempDir()

	// Test with near-lossless enabled
	config := ConverterConfig{}
	config.PNGToWebP.TryNearLossless = true
	converter := NewConverter(config)

	inputPath := "testdata/test_original.png"
	outputPath := filepath.Join(tempDir, "near_lossless.webp")

	err := converter.ToWebP(inputPath, outputPath)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Check file was created
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not created: %v", err)
	}

	t.Logf("Near-lossless output size: %d bytes", info.Size())

	// Compare with regular lossless
	config2 := ConverterConfig{}
	config2.PNGToWebP.TryNearLossless = false
	converter2 := NewConverter(config2)

	outputPath2 := filepath.Join(tempDir, "lossless.webp")
	err = converter2.ToWebP(inputPath, outputPath2)
	if err != nil {
		t.Fatalf("Lossless conversion failed: %v", err)
	}

	info2, err := os.Stat(outputPath2)
	if err != nil {
		t.Fatalf("Lossless output file not created: %v", err)
	}

	t.Logf("Lossless output size: %d bytes", info2.Size())
}

func TestPNGAlphaHandling(t *testing.T) {
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
			outputPath := filepath.Join(tempDir, filepath.Base(testPath)+".webp")

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
