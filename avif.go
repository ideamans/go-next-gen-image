package nextgenimage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

// ToAVIF converts an image to AVIF format
func (c *Converter) ToAVIF(inputPath, outputPath string) error {
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

	// Detect input format
	ext := strings.ToLower(filepath.Ext(inputPath))
	var params *vips.AvifExportParams
	var outputBuffer []byte

	switch ext {
	case ".jpg", ".jpeg":
		// JPEG to AVIF: lossy conversion with CQ
		params = vips.NewAvifExportParams()
		params.Quality = c.config.JPEGToAVIF.CQ
		params.Lossless = false
		params.StripMetadata = false // Keep ICC profile

		outputBuffer, _, err = image.ExportAvif(params)
		if err != nil {
			return fmt.Errorf("failed to export avif: %w", NewFormatError(err))
		}

	case ".png":
		// PNG to AVIF: lossless conversion
		params = vips.NewAvifExportParams()
		params.Lossless = true
		params.StripMetadata = false // Keep ICC profile

		outputBuffer, _, err = image.ExportAvif(params)
		if err != nil {
			return fmt.Errorf("failed to export avif: %w", NewFormatError(err))
		}

	case ".gif":
		// GIF to AVIF: not directly supported, would need ffmpeg
		return NewFormatError(fmt.Errorf("GIF to AVIF conversion is not supported by this library"))

	default:
		return fmt.Errorf("unsupported input format: %s", ext)
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
