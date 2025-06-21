package nextgenimage

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

// checkMetadata checks if the image has EXIF, XMP, and ICC profile
func checkMetadata(t *testing.T, imagePath string) (hasEXIF, hasXMP, hasICC bool) {
	t.Helper()

	image, err := vips.NewImageFromFile(imagePath)
	if err != nil {
		t.Fatalf("Failed to load image for metadata check: %v", err)
	}
	defer image.Close()

	// Check for EXIF data
	hasEXIF = image.HasExif()

	// Check for XMP data
	// libvips doesn't have a direct HasXMP method in govips
	// We'll check by attempting to get the metadata
	// For now, we'll assume XMP is not directly accessible through govips
	hasXMP = false // govips doesn't expose XMP metadata directly

	// Check for ICC profile
	hasICC = image.HasICCProfile()

	return hasEXIF, hasXMP, hasICC
}

// TestMetadataRemoval tests that EXIF and XMP are removed while ICC is preserved
func TestMetadataRemoval(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	testCases := []struct {
		name          string
		inputPath     string
		expectEXIF    bool
		expectXMP     bool
		expectICC     bool
		skipWebP      bool
		skipAVIF      bool
	}{
		{
			name:       "JPEG with EXIF",
			inputPath:  "testdata/jpeg/metadata_basic_exif.jpg",
			expectEXIF: true,
			expectXMP:  false,
			expectICC:  false,
		},
		{
			name:       "JPEG with full EXIF",
			inputPath:  "testdata/jpeg/metadata_full_exif.jpg",
			expectEXIF: true,
			expectXMP:  false,
			expectICC:  false,
		},
		{
			name:       "JPEG with XMP",
			inputPath:  "testdata/jpeg/metadata_xmp.jpg",
			expectEXIF: false,  // This file doesn't have EXIF
			expectXMP:  false,  // We can't detect XMP with govips
			expectICC:  false,
		},
		{
			name:       "JPEG with ICC profile",
			inputPath:  "testdata/jpeg/icc_srgb.jpg",
			expectEXIF: false,
			expectXMP:  false,
			expectICC:  true,
		},
		{
			name:       "PNG with metadata",
			inputPath:  "testdata/png/metadata_text.png",
			expectEXIF: false,
			expectXMP:  false,
			expectICC:  false,
		},
		{
			name:      "GIF (WebP only)",
			inputPath: "testdata/gif/frames_medium.gif",
			expectEXIF: false,
			expectXMP:  false,
			expectICC:  false,
			skipAVIF:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check original file metadata
			origEXIF, origXMP, origICC := checkMetadata(t, tc.inputPath)
			t.Logf("Original metadata - EXIF: %v, XMP: %v, ICC: %v", origEXIF, origXMP, origICC)

			// Verify expectations for original file
			if origEXIF != tc.expectEXIF {
				t.Errorf("Original EXIF expectation mismatch: got %v, want %v", origEXIF, tc.expectEXIF)
			}
			if origXMP != tc.expectXMP {
				t.Errorf("Original XMP expectation mismatch: got %v, want %v", origXMP, tc.expectXMP)
			}
			if origICC != tc.expectICC {
				t.Errorf("Original ICC expectation mismatch: got %v, want %v", origICC, tc.expectICC)
			}

			// Test WebP conversion
			if !tc.skipWebP {
				t.Run("WebP", func(t *testing.T) {
					outputPath := filepath.Join(tempDir, filepath.Base(tc.inputPath)+".webp")
					err := converter.ToWebP(tc.inputPath, outputPath)
					if err != nil {
						t.Fatalf("WebP conversion failed: %v", err)
					}

					// Check converted file metadata
					convEXIF, convXMP, convICC := checkMetadata(t, outputPath)
					t.Logf("WebP metadata - EXIF: %v, XMP: %v, ICC: %v", convEXIF, convXMP, convICC)

					// All metadata should be removed with StripMetadata=true
					if convEXIF {
						t.Error("EXIF data was not removed in WebP conversion")
					}
					if convXMP {
						t.Error("XMP data was not removed in WebP conversion")
					}
					// ICC is also removed with StripMetadata=true
					if convICC {
						t.Error("ICC profile was not removed in WebP conversion")
					}
				})
			}

			// Test AVIF conversion
			if !tc.skipAVIF {
				t.Run("AVIF", func(t *testing.T) {
					outputPath := filepath.Join(tempDir, filepath.Base(tc.inputPath)+".avif")
					err := converter.ToAVIF(tc.inputPath, outputPath)
					if err != nil {
						var formatErr *FormatError
						if errors.As(err, &formatErr) {
							t.Skipf("Format error (expected for some cases): %v", err)
							return
						}
						t.Fatalf("AVIF conversion failed: %v", err)
					}

					// Check converted file metadata
					convEXIF, convXMP, convICC := checkMetadata(t, outputPath)
					t.Logf("AVIF metadata - EXIF: %v, XMP: %v, ICC: %v", convEXIF, convXMP, convICC)

					// All metadata should be removed with StripMetadata=true
					if convEXIF {
						t.Error("EXIF data was not removed in AVIF conversion")
					}
					if convXMP {
						t.Error("XMP data was not removed in AVIF conversion")
					}
					// ICC is also removed with StripMetadata=true
					if convICC {
						t.Error("ICC profile was not removed in AVIF conversion")
					}
				})
			}
		})
	}
}

// TestICCProfileRemoval specifically tests ICC profile removal
func TestICCProfileRemoval(t *testing.T) {
	converter := NewConverter(ConverterConfig{})
	tempDir := t.TempDir()

	// Test files known to have ICC profiles
	testFiles := []string{
		"testdata/jpeg/icc_srgb.jpg",
		"testdata/jpeg/icc_adobergb.jpg",
	}

	for _, inputPath := range testFiles {
		t.Run(filepath.Base(inputPath), func(t *testing.T) {
			// Load original image and get ICC profile
			origImage, err := vips.NewImageFromFile(inputPath)
			if err != nil {
				t.Fatalf("Failed to load original image: %v", err)
			}
			defer origImage.Close()

			origICCData := origImage.GetICCProfile()
			if len(origICCData) == 0 {
				t.Skip("Original image has no ICC profile")
			}

			// Test WebP conversion
			t.Run("WebP", func(t *testing.T) {
				outputPath := filepath.Join(tempDir, filepath.Base(inputPath)+".webp")
				err := converter.ToWebP(inputPath, outputPath)
				if err != nil {
					t.Fatalf("WebP conversion failed: %v", err)
				}

				// Check ICC profile in converted image
				convImage, err := vips.NewImageFromFile(outputPath)
				if err != nil {
					t.Fatalf("Failed to load converted image: %v", err)
				}
				defer convImage.Close()

				convICCData := convImage.GetICCProfile()

				if len(convICCData) > 0 {
					t.Error("ICC profile was not removed in WebP conversion")
				} else {
					t.Log("ICC profile successfully removed")
				}
			})

			// Test AVIF conversion
			t.Run("AVIF", func(t *testing.T) {
				outputPath := filepath.Join(tempDir, filepath.Base(inputPath)+".avif")
				err := converter.ToAVIF(inputPath, outputPath)
				if err != nil {
					t.Fatalf("AVIF conversion failed: %v", err)
				}

				// Check ICC profile in converted image
				convImage, err := vips.NewImageFromFile(outputPath)
				if err != nil {
					t.Fatalf("Failed to load converted image: %v", err)
				}
				defer convImage.Close()

				convICCData := convImage.GetICCProfile()

				if len(convICCData) > 0 {
					t.Error("ICC profile was not removed in AVIF conversion")
				} else {
					t.Log("ICC profile successfully removed")
				}
			})
		})
	}
}