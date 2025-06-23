package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ideamans/go-next-gen-image"
	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestRootCommand(t *testing.T) {
	// Create a new command for each test to avoid state issues
	cmd := &cobra.Command{
		Use:   "nextgenimage",
		Short: "Convert traditional web images to next-gen formats",
		Long: `nextgenimage converts traditional web image formats (JPEG, PNG, GIF) 
to next-generation formats (WebP, AVIF) following best practices.`,
		Version: version,
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")

	output, err := executeCommand(cmd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "nextgenimage converts traditional web image formats") {
		t.Errorf("Expected help text not found in output: %s", output)
	}
}

func TestWebPCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:        "help",
			args:        []string{"webp", "--help"},
			expectError: false,
		},
		{
			name:        "missing arguments",
			args:        []string{"webp"},
			expectError: true,
		},
		{
			name:          "invalid quality",
			args:          []string{"webp", "--quality", "150", "input.jpg", "output.webp"},
			expectError:   true,
			errorContains: "quality must be between 1 and 100",
		},
		{
			name:          "non-existent input file",
			args:          []string{"webp", "--quality", "80", "non-existent.jpg", "output.webp"},
			expectError:   true,
			errorContains: "input file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			webpQuality = 80
			webpTryNearLossless = false
			verbose = false
			quiet = false
			
			// Create fresh command instance
			cmd := &cobra.Command{
				Use:   "nextgenimage",
				Short: "Convert traditional web images to next-gen formats",
				Long: `nextgenimage converts traditional web image formats (JPEG, PNG, GIF) 
to next-generation formats (WebP, AVIF) following best practices.`,
				Version: version,
			}
			cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
			cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")
			
			// Create a new webp command for this test
			webpTestCmd := &cobra.Command{
				Use:   "webp <input-file> <output-file>",
				Short: "Convert image to WebP format",
				Long: `Convert JPEG, PNG, or GIF images to WebP format.
	
JPEG images are converted using lossy compression with configurable quality.
PNG images are converted using lossless compression, with optional near-lossless.
GIF images are converted to animated WebP preserving animation properties.`,
				Args: cobra.ExactArgs(2),
				RunE: runWebP,
			}
			webpTestCmd.Flags().IntVarP(&webpQuality, "quality", "q", 80, "JPEG to WebP quality (1-100)")
			webpTestCmd.Flags().BoolVar(&webpTryNearLossless, "try-near-lossless", false, "Try near-lossless compression for PNG to WebP")
			
			cmd.AddCommand(webpTestCmd)

			output, err := executeCommand(cmd, tt.args...)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if strings.Contains(tt.args[0], "help") && !strings.Contains(output, "Convert image to WebP format") {
					t.Errorf("Expected help text not found in output: %s", output)
				}
			}
		})
	}
}

func TestAVIFCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:        "help",
			args:        []string{"avif", "--help"},
			expectError: false,
		},
		{
			name:        "missing arguments",
			args:        []string{"avif"},
			expectError: true,
		},
		{
			name:          "invalid CQ",
			args:          []string{"avif", "--cq", "100", "input.jpg", "output.avif"},
			expectError:   true,
			errorContains: "CQ must be between 1 and 63",
		},
		{
			name:          "non-existent input file",
			args:          []string{"avif", "--cq", "25", "non-existent.jpg", "output.avif"},
			expectError:   true,
			errorContains: "input file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			avifCQ = 25
			verbose = false
			quiet = false
			
			// Create fresh command instance
			cmd := &cobra.Command{
				Use:   "nextgenimage",
				Short: "Convert traditional web images to next-gen formats",
				Long: `nextgenimage converts traditional web image formats (JPEG, PNG, GIF) 
to next-generation formats (WebP, AVIF) following best practices.`,
				Version: version,
			}
			cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
			cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")
			
			// Create a new avif command for this test
			avifTestCmd := &cobra.Command{
				Use:   "avif <input-file> <output-file>",
				Short: "Convert image to AVIF format",
				Long: `Convert JPEG or PNG images to AVIF format.
	
JPEG images are converted using lossy compression with configurable CQ value.
PNG images are converted using lossless compression.
GIF to AVIF conversion is not supported.`,
				Args: cobra.ExactArgs(2),
				RunE: runAVIF,
			}
			avifTestCmd.Flags().IntVar(&avifCQ, "cq", 25, "JPEG to AVIF CQ value (1-63, lower is better quality)")
			
			cmd.AddCommand(avifTestCmd)

			output, err := executeCommand(cmd, tt.args...)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if strings.Contains(tt.args[0], "help") && !strings.Contains(output, "Convert image to AVIF format") {
					t.Errorf("Expected help text not found in output: %s", output)
				}
			}
		})
	}
}

func TestWebPConversionIntegration(t *testing.T) {
	// Skip if no test data available
	testJPEG := "../../testdata/test_original.jpg"
	if _, err := os.Stat(testJPEG); err != nil {
		t.Skip("Test data not found, skipping integration test")
	}

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.webp")

	// Reset global flags
	webpQuality = 80
	webpTryNearLossless = false
	verbose = false
	quiet = false
	
	// Create fresh command instance
	cmd := &cobra.Command{
		Use:   "nextgenimage",
		Short: "Convert traditional web images to next-gen formats",
		Long: `nextgenimage converts traditional web image formats (JPEG, PNG, GIF) 
to next-generation formats (WebP, AVIF) following best practices.`,
		Version: version,
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")
	
	// Create a new webp command for this test
	webpTestCmd := &cobra.Command{
		Use:   "webp <input-file> <output-file>",
		Short: "Convert image to WebP format",
		Long: `Convert JPEG, PNG, or GIF images to WebP format.
	
JPEG images are converted using lossy compression with configurable quality.
PNG images are converted using lossless compression, with optional near-lossless.
GIF images are converted to animated WebP preserving animation properties.`,
		Args: cobra.ExactArgs(2),
		RunE: runWebP,
	}
	webpTestCmd.Flags().IntVarP(&webpQuality, "quality", "q", 80, "JPEG to WebP quality (1-100)")
	webpTestCmd.Flags().BoolVar(&webpTryNearLossless, "try-near-lossless", false, "Try near-lossless compression for PNG to WebP")
	
	cmd.AddCommand(webpTestCmd)

	// Add --quiet flag to suppress output
	output, err := executeCommand(cmd, "--quiet", "webp", testJPEG, outputPath)

	// Allow FormatError (output size not smaller)
	if err != nil {
		var formatErr *nextgenimage.FormatError
		if !errors.As(err, &formatErr) {
			t.Errorf("Unexpected error: %v", err)
		}
		// FormatError is acceptable for this test
		return
	}

	// Check if output file was created
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("Output file was not created: %v", err)
	}

	// When using --quiet flag, there should be no output
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %s", output)
	}
}

func TestAVIFConversionIntegration(t *testing.T) {
	// Skip if no test data available
	testJPEG := "../../testdata/test_original.jpg"
	if _, err := os.Stat(testJPEG); err != nil {
		t.Skip("Test data not found, skipping integration test")
	}

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.avif")

	// Reset global flags
	avifCQ = 25
	verbose = false
	quiet = false
	
	// Create fresh command instance
	cmd := &cobra.Command{
		Use:   "nextgenimage",
		Short: "Convert traditional web images to next-gen formats",
		Long: `nextgenimage converts traditional web image formats (JPEG, PNG, GIF) 
to next-generation formats (WebP, AVIF) following best practices.`,
		Version: version,
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")
	
	// Create a new avif command for this test
	avifTestCmd := &cobra.Command{
		Use:   "avif <input-file> <output-file>",
		Short: "Convert image to AVIF format",
		Long: `Convert JPEG or PNG images to AVIF format.
	
JPEG images are converted using lossy compression with configurable CQ value.
PNG images are converted using lossless compression.
GIF to AVIF conversion is not supported.`,
		Args: cobra.ExactArgs(2),
		RunE: runAVIF,
	}
	avifTestCmd.Flags().IntVar(&avifCQ, "cq", 25, "JPEG to AVIF CQ value (1-63, lower is better quality)")
	
	cmd.AddCommand(avifTestCmd)

	// Add --quiet flag to suppress output
	output, err := executeCommand(cmd, "--quiet", "avif", testJPEG, outputPath)

	// Allow FormatError (output size not smaller)
	if err != nil {
		var formatErr *nextgenimage.FormatError
		if !errors.As(err, &formatErr) {
			t.Errorf("Unexpected error: %v", err)
		}
		// FormatError is acceptable for this test
		return
	}

	// Check if output file was created
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("Output file was not created: %v", err)
	}

	// When using --quiet flag, there should be no output
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %s", output)
	}
}