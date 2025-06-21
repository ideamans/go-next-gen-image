package nextgenimage

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// ImageType represents the type of an image file
type ImageType string

const (
	ImageTypeJPEG    ImageType = "jpeg"
	ImageTypePNG     ImageType = "png"
	ImageTypeGIF     ImageType = "gif"
	ImageTypeWebP    ImageType = "webp"
	ImageTypeAVIF    ImageType = "avif"
	ImageTypeUnknown ImageType = "unknown"
)

// Magic bytes for image format detection
var (
	jpegMagic1 = []byte{0xFF, 0xD8, 0xFF}
	pngMagic   = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	gifMagic1  = []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61} // GIF87a
	gifMagic2  = []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61} // GIF89a
	webpMagic  = []byte{0x52, 0x49, 0x46, 0x46}              // RIFF
	webpType   = []byte{0x57, 0x45, 0x42, 0x50}              // WEBP
)

// DetectImageType detects the image type from a file path
func DetectImageType(filePath string) (ImageType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ImageTypeUnknown, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return DetectImageTypeFromReader(file)
}

// DetectImageTypeFromReader detects the image type from an io.Reader
func DetectImageTypeFromReader(r io.Reader) (ImageType, error) {
	// Read enough bytes to detect all formats
	// AVIF needs more bytes due to its complex structure
	buf := make([]byte, 32)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return ImageTypeUnknown, fmt.Errorf("failed to read magic bytes: %w", err)
	}
	if n < 12 {
		return ImageTypeUnknown, fmt.Errorf("file too small to determine type")
	}

	return detectFromBytes(buf[:n])
}

// DetectImageTypeFromBytes detects the image type from a byte slice
func DetectImageTypeFromBytes(data []byte) ImageType {
	imgType, _ := detectFromBytes(data)
	return imgType
}

func detectFromBytes(buf []byte) (ImageType, error) {
	// Check PNG
	if len(buf) >= 8 && bytes.Equal(buf[:8], pngMagic) {
		return ImageTypePNG, nil
	}

	// Check JPEG
	if len(buf) >= 3 && bytes.Equal(buf[:3], jpegMagic1) {
		return ImageTypeJPEG, nil
	}

	// Check GIF
	if len(buf) >= 6 {
		if bytes.Equal(buf[:6], gifMagic1) || bytes.Equal(buf[:6], gifMagic2) {
			return ImageTypeGIF, nil
		}
	}

	// Check WebP
	if len(buf) >= 12 && bytes.Equal(buf[:4], webpMagic) && bytes.Equal(buf[8:12], webpType) {
		return ImageTypeWebP, nil
	}

	// Check AVIF
	// AVIF detection is more complex as it's based on ISO BMFF (MP4-like) structure
	if len(buf) >= 12 && isAVIF(buf) {
		return ImageTypeAVIF, nil
	}

	return ImageTypeUnknown, nil
}

// isAVIF checks if the data represents an AVIF image
func isAVIF(buf []byte) bool {
	// AVIF files start with an ftyp box
	// Check for ftyp box signature at offset 4
	if len(buf) < 12 {
		return false
	}

	// Box size (first 4 bytes) - we don't validate this
	// Box type should be "ftyp" at offset 4-7
	if !bytes.Equal(buf[4:8], []byte("ftyp")) {
		return false
	}

	// Check major brand at offset 8-11
	majorBrand := string(buf[8:12])
	
	// AVIF major brands
	avifBrands := []string{"avif", "avis"}
	for _, brand := range avifBrands {
		if majorBrand == brand {
			return true
		}
	}

	// Also check compatible brands if we have enough data
	if len(buf) >= 16 {
		// Simple check: look for "avif" or "avis" in the compatible brands
		// In a real implementation, we would parse the full ftyp box
		compatibleBrands := string(buf[16:])
		return bytes.Contains([]byte(compatibleBrands), []byte("avif")) ||
			bytes.Contains([]byte(compatibleBrands), []byte("avis"))
	}

	return false
}

// String returns the string representation of the ImageType
func (t ImageType) String() string {
	return string(t)
}

// IsSupported returns true if the image type is supported for conversion
func (t ImageType) IsSupported() bool {
	switch t {
	case ImageTypeJPEG, ImageTypePNG, ImageTypeGIF:
		return true
	default:
		return false
	}
}