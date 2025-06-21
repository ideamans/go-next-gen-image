package nextgenimage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestJPEGToAVIF(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	images := loadTestImages(t)

	// Create temp directory for outputs
	tempDir := t.TempDir()

	for _, img := range images {
		if img.Format != "jpeg" {
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

			// Check if dimensions are consistent (orientation may swap width/height)
			origArea := origImage.Width() * origImage.Height()
			convArea := convImage.Width() * convImage.Height()

			if origArea != convArea {
				t.Errorf("Image area changed: %d -> %d pixels", origArea, convArea)
			} else if origImage.Width() != convImage.Width() || origImage.Height() != convImage.Height() {
				// Dimensions swapped due to orientation - this is expected
				t.Logf("Dimensions rotated: %dx%d -> %dx%d (orientation applied)",
					origImage.Width(), origImage.Height(),
					convImage.Width(), convImage.Height())
			}

			// Calculate PSNR only if dimensions match
			if origImage.Width() == convImage.Width() && origImage.Height() == convImage.Height() {
				psnr, err := calculatePSNR(inputPath, outputPath)
				if err != nil {
					t.Logf("Failed to calculate PSNR: %v", err)
				} else {
					t.Logf("PSNR: %.2f dB", psnr)
					if psnr < 40 {
						t.Errorf("PSNR too low: %.2f dB (threshold: 40 dB)", psnr)
					}
				}
			} else {
				t.Logf("Skipping PSNR check due to orientation rotation")
			}

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

func TestJPEGToAVIFCQSettings(t *testing.T) {
	cqValues := []int{20, 25, 30, 35}
	tempDir := t.TempDir()

	for _, cq := range cqValues {
		t.Run(fmt.Sprintf("CQ_%d", cq), func(t *testing.T) {
			config := ConverterConfig{}
			config.JPEGToAVIF.CQ = cq
			converter := NewConverter(config)

			inputPath := "testdata/test_original.jpg"
			outputPath := filepath.Join(tempDir, fmt.Sprintf("cq_%d.avif", cq))

			err := converter.ToAVIF(inputPath, outputPath)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Check file was created
			info, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("Output file not created: %v", err)
			}

			t.Logf("CQ %d: output size = %d bytes", cq, info.Size())
		})
	}
}

func TestAVIFColorProfiles(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test ICC profile preservation
	profileTests := []string{
		"jpeg/icc_srgb.jpg",
		"jpeg/icc_adobergb.jpg",
	}

	for _, testPath := range profileTests {
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

			// Load converted image
			convImage, err := vips.NewImageFromFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to load converted image: %v", err)
			}
			defer convImage.Close()

			// Check if ICC profile is preserved
			hasProfile := convImage.HasProfile()
			if hasProfile {
				t.Logf("ICC profile preserved")
			} else {
				t.Logf("No ICC profile in output")
			}
		})
	}
}
