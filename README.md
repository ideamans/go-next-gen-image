# go-next-gen-image

A Go library for converting traditional web image formats (JPEG, PNG, GIF) to next-generation formats (WebP, AVIF) following best practices.

[![Go Reference](https://pkg.go.dev/badge/github.com/ideamans/go-next-gen-image.svg)](https://pkg.go.dev/github.com/ideamans/go-next-gen-image)
[![CI](https://github.com/ideamans/go-next-gen-image/actions/workflows/ci.yml/badge.svg)](https://github.com/ideamans/go-next-gen-image/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ideamans/go-next-gen-image)](https://goreportcard.com/report/github.com/ideamans/go-next-gen-image)
[![codecov](https://codecov.io/gh/ideamans/go-next-gen-image/branch/main/graph/badge.svg)](https://codecov.io/gh/ideamans/go-next-gen-image)

## Features

- Convert JPEG/PNG/GIF images to WebP format
- Convert JPEG/PNG images to AVIF format
- Automatic image optimization with size reduction checks
- Preserve important metadata (ICC profiles)
- Support for animated GIF to WebP conversion
- Configurable quality settings
- Thread-safe concurrent conversions

## Installation

```bash
go get github.com/ideamans/go-next-gen-image
```

### Prerequisites

This library requires libvips to be installed on your system:

**Ubuntu/Debian:**
```bash
sudo apt-get install libvips-dev
```

**macOS:**
```bash
brew install vips
```

**Windows:**
```bash
choco install vips
```

## Usage

```go
package main

import (
    "log"
    nextgenimage "github.com/ideamans/go-next-gen-image"
)

func main() {
    // Create converter with default settings
    converter := nextgenimage.NewConverter(nextgenimage.ConverterConfig{})

    // Convert JPEG to WebP
    err := converter.ToWebP("input.jpg", "output.webp")
    if err != nil {
        log.Fatal(err)
    }

    // Convert PNG to AVIF
    err = converter.ToAVIF("input.png", "output.avif")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Configuration

```go
config := nextgenimage.ConverterConfig{
    JPEGToWebP: struct {
        Quality int
    }{
        Quality: 85, // Default: 80
    },
    PNGToWebP: struct {
        TryNearLossless bool
    }{
        TryNearLossless: true, // Default: false
    },
    JPEGToAVIF: struct {
        CQ int
    }{
        CQ: 20, // Default: 25
    },
}

converter := nextgenimage.NewConverter(config)
```

## Conversion Rules

### JPEG to WebP
- Lossy compression
- Configurable quality (default: 80)
- Auto-rotation based on EXIF orientation
- Removes all metadata (EXIF, XMP, ICC)

### PNG to WebP
- Lossless compression by default
- Optional near-lossless mode for better compression
- Removes all metadata (EXIF, XMP, ICC)
- Alpha channel support

### GIF to WebP
- Lossless frame conversion
- Animation preservation (timing, loops)
- All frames converted to WebP animation

### JPEG to AVIF
- Lossy compression with CQ (Constant Quality) mode
- Configurable CQ value (default: 25, lower = better quality)
- Auto-rotation based on EXIF orientation
- Removes all metadata (EXIF, XMP, ICC)

### PNG to AVIF
- Lossless compression
- Removes all metadata (EXIF, XMP, ICC)
- Alpha channel support

### GIF to AVIF
- Not supported (returns FormatError)

## Error Handling

The library distinguishes between data-related errors and system errors:

```go
err := converter.ToWebP("input.jpg", "output.webp")
if err != nil {
    var formatErr *nextgenimage.FormatError
    if errors.As(err, &formatErr) {
        // Handle format-specific error (e.g., unsupported format, size increase)
        log.Printf("Format error: %v", err)
    } else {
        // Handle system error (e.g., file not found, permission denied)
        log.Printf("System error: %v", err)
    }
}
```

## Performance

- The library uses libvips for efficient image processing
- Supports concurrent conversions (thread-safe)
- Automatically checks output size and returns FormatError if the converted image is larger than the original

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## License

MIT License - see [LICENSE](LICENSE) file for details

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to update tests as appropriate and ensure all tests pass before submitting a PR.