package nextgenimage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidbyttow/govips/v2/vips"
)

// ToWebP converts an image to WebP format
func (c *Converter) ToWebP(inputPath, outputPath string) error {
	// Check input file
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}

	// Load image
	image, err := vips.NewImageFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", NewFormatError(err))
	}
	defer image.Close()

	// Auto-rotate based on EXIF orientation
	err = image.AutoRotate()
	if err != nil {
		return fmt.Errorf("failed to auto-rotate: %w", NewFormatError(err))
	}

	// Detect input format using magic bytes
	imgType, err := DetectImageType(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect image type: %w", err)
	}

	if !imgType.IsSupported() {
		return fmt.Errorf("unsupported image format: %s", imgType)
	}

	var params *vips.WebpExportParams
	var outputBuffer []byte

	switch imgType {
	case ImageTypeJPEG:
		// JPEG to WebP: lossy conversion
		params = vips.NewWebpExportParams()
		params.Quality = c.config.JPEGToWebP.Quality
		params.Lossless = false
		params.StripMetadata = true // Strip all metadata during export

		outputBuffer, _, err = image.ExportWebp(params)
		if err != nil {
			return fmt.Errorf("failed to export webp: %w", NewFormatError(err))
		}

	case ImageTypePNG:
		// PNG to WebP: lossless conversion
		params = vips.NewWebpExportParams()
		params.Lossless = true
		params.StripMetadata = true // Strip all metadata during export

		outputBuffer, _, err = image.ExportWebp(params)
		if err != nil {
			return fmt.Errorf("failed to export webp: %w", NewFormatError(err))
		}

		// Try near-lossless if configured
		if c.config.PNGToWebP.TryNearLossless {
			nearLosslessParams := vips.NewWebpExportParams()
			nearLosslessParams.Lossless = false
			nearLosslessParams.NearLossless = true
			nearLosslessParams.Quality = 100
			nearLosslessParams.StripMetadata = true // Strip all metadata during export

			nearLosslessBuffer, _, err2 := image.ExportWebp(nearLosslessParams)
			if err2 == nil && len(nearLosslessBuffer) < len(outputBuffer) {
				outputBuffer = nearLosslessBuffer
			}
		}

	case ImageTypeGIF:
		// GIF to WebP: animated webp conversion
		// Load as animated image
		animParams := vips.NewImportParams()
		animParams.NumPages.Set(-1) // Load all pages/frames

		animImage, err := vips.LoadImageFromFile(inputPath, animParams)
		if err != nil {
			return fmt.Errorf("failed to load animated gif: %w", NewFormatError(err))
		}
		defer animImage.Close()

		// Export as animated WebP
		params = vips.NewWebpExportParams()
		params.Lossless = true // GIF frames are lossless
		params.StripMetadata = true // Strip all metadata during export

		// Get page height for animation
		pageHeight := animImage.PageHeight()
		// Note: The current govips version doesn't expose animation delay/loop
		// parameters directly. The animation will be preserved but timing
		// may use defaults if pageHeight > 0 (animated image)
		_ = pageHeight // Mark as intentionally unused

		outputBuffer, _, err = animImage.ExportWebp(params)
		if err != nil {
			return fmt.Errorf("failed to export animated webp: %w", NewFormatError(err))
		}

	}

	// Check if output is smaller than input
	if int64(len(outputBuffer)) >= inputInfo.Size() {
		return NewFormatError(fmt.Errorf("output file size (%d) is not smaller than input (%d)", len(outputBuffer), inputInfo.Size()))
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, outputBuffer, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
