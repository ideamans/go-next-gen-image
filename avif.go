package nextgenimage

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Remove metadata except ICC profile and orientation
	// RemoveMetadata() keeps ICC profile, orientation, and page info by default
	err = image.RemoveMetadata()
	if err != nil {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}

	// Detect input format using magic bytes
	imgType, err := DetectImageType(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect image type: %w", err)
	}

	if !imgType.IsSupported() {
		return fmt.Errorf("unsupported image format: %s", imgType)
	}

	// GIF to AVIF is not supported
	if imgType == ImageTypeGIF {
		return NewFormatError(fmt.Errorf("GIF to AVIF conversion is not supported"))
	}

	var params *vips.AvifExportParams
	var outputBuffer []byte

	switch imgType {
	case ImageTypeJPEG:
		// JPEG to AVIF: lossy conversion with CQ
		params = vips.NewAvifExportParams()
		params.Quality = c.config.JPEGToAVIF.CQ
		params.Lossless = false
		params.StripMetadata = false // Already removed unwanted metadata, keep ICC

		outputBuffer, _, err = image.ExportAvif(params)
		if err != nil {
			return fmt.Errorf("failed to export avif: %w", NewFormatError(err))
		}

	case ImageTypePNG:
		// PNG to AVIF: lossless conversion
		params = vips.NewAvifExportParams()
		params.Lossless = true
		params.StripMetadata = false // Already removed unwanted metadata, keep ICC

		outputBuffer, _, err = image.ExportAvif(params)
		if err != nil {
			return fmt.Errorf("failed to export avif: %w", NewFormatError(err))
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
