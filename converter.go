package nextgenimage

import (
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
)

// FormatError represents errors related to data or format problems
type FormatError struct {
	err error
}

func NewFormatError(err error) *FormatError {
	return &FormatError{err: err}
}

func (e *FormatError) Error() string {
	return fmt.Sprintf("format error: %v", e.err)
}

func (e *FormatError) Unwrap() error {
	return e.err
}

// ConverterConfig holds configuration for image conversion
type ConverterConfig struct {
	JPEGToWebP struct {
		Quality int // Default: 80
	}
	PNGToWebP struct {
		TryNearLossless bool // Default: false
	}
	JPEGToAVIF struct {
		CQ int // Default: 25
	}
}

// Converter handles image format conversions
type Converter struct {
	config ConverterConfig
}

// NewConverter creates a new converter instance
func NewConverter(config ConverterConfig) *Converter {
	// Set defaults
	if config.JPEGToWebP.Quality == 0 {
		config.JPEGToWebP.Quality = 80
	}
	if config.JPEGToAVIF.CQ == 0 {
		config.JPEGToAVIF.CQ = 25
	}
	return &Converter{config: config}
}

func init() {
	// Initialize vips once
	vips.Startup(nil)
}
