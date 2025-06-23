package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ideamans/go-next-gen-image"
	"github.com/spf13/cobra"
)

var (
	webpQuality         int
	webpTryNearLossless bool
)

var webpCmd = &cobra.Command{
	Use:   "webp <input-file> <output-file>",
	Short: "Convert image to WebP format",
	Long: `Convert JPEG, PNG, or GIF images to WebP format.
	
JPEG images are converted using lossy compression with configurable quality.
PNG images are converted using lossless compression, with optional near-lossless.
GIF images are converted to animated WebP preserving animation properties.`,
	Args: cobra.ExactArgs(2),
	RunE: runWebP,
}

func init() {
	webpCmd.Flags().IntVarP(&webpQuality, "quality", "q", 80, "JPEG to WebP quality (1-100)")
	webpCmd.Flags().BoolVar(&webpTryNearLossless, "try-near-lossless", false, "Try near-lossless compression for PNG to WebP")
}

func runWebP(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]

	// Validate quality
	if webpQuality < 1 || webpQuality > 100 {
		return fmt.Errorf("quality must be between 1 and 100")
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", inputPath)
		}
		return fmt.Errorf("failed to access input file: %w", err)
	}

	// Log start
	if !quiet {
		fmt.Printf("Converting %s to WebP...\n", filepath.Base(inputPath))
	}
	if verbose {
		fmt.Printf("[INFO] Input: %s\n", inputPath)
		fmt.Printf("[INFO] Output: %s\n", outputPath)
		fmt.Printf("[INFO] Quality: %d\n", webpQuality)
		if webpTryNearLossless {
			fmt.Printf("[INFO] Near-lossless: enabled\n")
		}
	}

	// Create converter with configuration
	config := nextgenimage.ConverterConfig{}
	config.JPEGToWebP.Quality = webpQuality
	config.PNGToWebP.TryNearLossless = webpTryNearLossless

	converter := nextgenimage.NewConverter(config)

	// Get file sizes for comparison
	inputInfo, _ := os.Stat(inputPath)
	inputSize := inputInfo.Size()

	// Perform conversion
	err := converter.ToWebP(inputPath, outputPath)
	if err != nil {
		var formatErr *nextgenimage.FormatError
		if errors.As(err, &formatErr) {
			if !quiet {
				fmt.Printf("✗ %s → %s (FormatError: %v)\n", 
					filepath.Base(inputPath), 
					filepath.Base(outputPath),
					err)
			}
			return formatErr
		}
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Get output file size
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to stat output file: %w", err)
	}
	outputSize := outputInfo.Size()

	// Calculate size reduction
	reduction := float64(inputSize-outputSize) / float64(inputSize) * 100

	// Log success
	if !quiet {
		fmt.Printf("✓ %s → %s (%s → %s, %.1f%%)\n",
			filepath.Base(inputPath),
			filepath.Base(outputPath),
			formatBytes(inputSize),
			formatBytes(outputSize),
			reduction)
	}

	if verbose {
		fmt.Printf("[INFO] Successfully converted: %s → %s\n", inputPath, outputPath)
	}

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}