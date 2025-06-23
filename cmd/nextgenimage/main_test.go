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
			cmd.AddCommand(webpCmd)

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
			cmd.AddCommand(avifCmd)

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
	
	// Reset flags for this test
	quiet = true // Suppress output for cleaner test
	cmd.AddCommand(webpCmd)

	output, err := executeCommand(cmd, "webp", testJPEG, outputPath)

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

	// If no error, output should contain success indicator
	if !quiet && !strings.Contains(output, "✓") {
		t.Errorf("Expected success indicator in output: %s", output)
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
	
	// Reset flags for this test
	quiet = true // Suppress output for cleaner test
	cmd.AddCommand(avifCmd)

	output, err := executeCommand(cmd, "avif", testJPEG, outputPath)

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

	// If no error, output should contain success indicator
	if !quiet && !strings.Contains(output, "✓") {
		t.Errorf("Expected success indicator in output: %s", output)
	}
}