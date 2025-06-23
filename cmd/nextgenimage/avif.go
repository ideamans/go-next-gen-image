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
	avifCQ int
)

var avifCmd = &cobra.Command{
	Use:   "avif <input-file> <output-file>",
	Short: "Convert image to AVIF format",
	Long: `Convert JPEG or PNG images to AVIF format.
	
JPEG images are converted using lossy compression with configurable CQ value.
PNG images are converted using lossless compression.
GIF to AVIF conversion is not supported.`,
	Args: cobra.ExactArgs(2),
	RunE: runAVIF,
}

func init() {
	avifCmd.Flags().IntVar(&avifCQ, "cq", 25, "JPEG to AVIF CQ value (1-63, lower is better quality)")
}

func runAVIF(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]

	// Validate CQ
	if avifCQ < 1 || avifCQ > 63 {
		return fmt.Errorf("CQ must be between 1 and 63")
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
		fmt.Printf("Converting %s to AVIF...\n", filepath.Base(inputPath))
	}
	if verbose {
		fmt.Printf("[INFO] Input: %s\n", inputPath)
		fmt.Printf("[INFO] Output: %s\n", outputPath)
		fmt.Printf("[INFO] CQ: %d\n", avifCQ)
	}

	// Create converter with configuration
	config := nextgenimage.ConverterConfig{}
	config.JPEGToAVIF.CQ = avifCQ

	converter := nextgenimage.NewConverter(config)

	// Get file sizes for comparison
	inputInfo, _ := os.Stat(inputPath)
	inputSize := inputInfo.Size()

	// Perform conversion
	err := converter.ToAVIF(inputPath, outputPath)
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