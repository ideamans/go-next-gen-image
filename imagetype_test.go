package nextgenimage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectImageTypeFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ImageType
	}{
		{
			name:     "JPEG with FFD8FF",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			expected: ImageTypeJPEG,
		},
		{
			name:     "PNG",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00},
			expected: ImageTypePNG,
		},
		{
			name:     "GIF87a",
			data:     []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61, 0x01, 0x00, 0x01, 0x00},
			expected: ImageTypeGIF,
		},
		{
			name:     "GIF89a",
			data:     []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00},
			expected: ImageTypeGIF,
		},
		{
			name:     "WebP",
			data:     []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expected: ImageTypeWebP,
		},
		{
			name: "AVIF with avif brand",
			data: []byte{
				0x00, 0x00, 0x00, 0x20, // box size
				0x66, 0x74, 0x79, 0x70, // ftyp
				0x61, 0x76, 0x69, 0x66, // avif major brand
				0x00, 0x00, 0x00, 0x00, // minor version
			},
			expected: ImageTypeAVIF,
		},
		{
			name: "AVIF with avis brand",
			data: []byte{
				0x00, 0x00, 0x00, 0x20, // box size
				0x66, 0x74, 0x79, 0x70, // ftyp
				0x61, 0x76, 0x69, 0x73, // avis major brand
				0x00, 0x00, 0x00, 0x00, // minor version
			},
			expected: ImageTypeAVIF,
		},
		{
			name: "AVIF with mif1 major brand and avif compatible",
			data: []byte{
				0x00, 0x00, 0x00, 0x24, // box size
				0x66, 0x74, 0x79, 0x70, // ftyp
				0x6D, 0x69, 0x66, 0x31, // mif1 major brand
				0x00, 0x00, 0x00, 0x00, // minor version
				0x61, 0x76, 0x69, 0x66, // avif compatible brand
				0x6D, 0x69, 0x66, 0x31, // mif1 compatible brand
			},
			expected: ImageTypeAVIF,
		},
		{
			name:     "Unknown - too small",
			data:     []byte{0x00, 0x01},
			expected: ImageTypeUnknown,
		},
		{
			name:     "Unknown - random data",
			data:     []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B},
			expected: ImageTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectImageTypeFromBytes(tt.data)
			if result != tt.expected {
				t.Errorf("DetectImageTypeFromBytes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectImageTypeFromReader(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ImageType
		wantErr  bool
	}{
		{
			name:     "Valid JPEG",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			expected: ImageTypeJPEG,
			wantErr:  false,
		},
		{
			name:     "Too small",
			data:     []byte{0xFF, 0xD8},
			expected: ImageTypeUnknown,
			wantErr:  true,
		},
		{
			name:     "Empty reader",
			data:     []byte{},
			expected: ImageTypeUnknown,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			result, err := DetectImageTypeFromReader(reader)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectImageTypeFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result != tt.expected {
				t.Errorf("DetectImageTypeFromReader() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectImageType(t *testing.T) {
	// Test with real files from testdata
	testFiles := []struct {
		path     string
		expected ImageType
	}{
		{"testdata/test_original.jpg", ImageTypeJPEG},
		{"testdata/test_original.png", ImageTypePNG},
		{"testdata/test_original.gif", ImageTypeGIF},
		{"testdata/jpeg/quality_80.jpg", ImageTypeJPEG},
		{"testdata/png/colortype_rgba.png", ImageTypePNG},
		{"testdata/gif/frames_medium.gif", ImageTypeGIF},
	}

	for _, tf := range testFiles {
		t.Run(tf.path, func(t *testing.T) {
			// Skip if file doesn't exist
			if _, err := os.Stat(tf.path); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", tf.path)
				return
			}

			result, err := DetectImageType(tf.path)
			if err != nil {
				t.Fatalf("DetectImageType() error = %v", err)
			}

			if result != tf.expected {
				t.Errorf("DetectImageType() = %v, want %v", result, tf.expected)
			}
		})
	}

	// Test with non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := DetectImageType("non_existent_file.jpg")
		if err == nil {
			t.Error("DetectImageType() expected error for non-existent file")
		}
	})
}

func TestImageTypeString(t *testing.T) {
	tests := []struct {
		imageType ImageType
		expected  string
	}{
		{ImageTypeJPEG, "jpeg"},
		{ImageTypePNG, "png"},
		{ImageTypeGIF, "gif"},
		{ImageTypeWebP, "webp"},
		{ImageTypeAVIF, "avif"},
		{ImageTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.imageType.String(); got != tt.expected {
				t.Errorf("ImageType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestImageTypeIsSupported(t *testing.T) {
	tests := []struct {
		imageType ImageType
		expected  bool
	}{
		{ImageTypeJPEG, true},
		{ImageTypePNG, true},
		{ImageTypeGIF, true},
		{ImageTypeWebP, false},
		{ImageTypeAVIF, false},
		{ImageTypeUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.imageType.String(), func(t *testing.T) {
			if got := tt.imageType.IsSupported(); got != tt.expected {
				t.Errorf("ImageType.IsSupported() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test with WebP and AVIF files if they exist in testdata
func TestDetectImageTypeWebPAVIF(t *testing.T) {
	// Create test WebP data (minimal valid WebP)
	webpData := []byte{
		0x52, 0x49, 0x46, 0x46, // RIFF
		0x24, 0x00, 0x00, 0x00, // file size (36 bytes)
		0x57, 0x45, 0x42, 0x50, // WEBP
		0x56, 0x50, 0x38, 0x20, // VP8 chunk
		0x18, 0x00, 0x00, 0x00, // chunk size
		0x30, 0x01, 0x00, 0x9d, // frame tag
		0x01, 0x2a, 0x01, 0x00, // width/height
		0x01, 0x00, 0x00, 0x00, // data
	}

	t.Run("WebP from bytes", func(t *testing.T) {
		result := DetectImageTypeFromBytes(webpData)
		if result != ImageTypeWebP {
			t.Errorf("Expected WebP, got %v", result)
		}
	})

	// Test WebP files from testdata if they exist
	webpFiles := []string{
		"testdata/webp/lossy/quality_80.webp",
		"testdata/webp/lossless/original.webp",
	}

	for _, path := range webpFiles {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("WebP test file %s not found", path)
				return
			}

			result, err := DetectImageType(path)
			if err != nil {
				t.Fatalf("DetectImageType() error = %v", err)
			}

			if result != ImageTypeWebP {
				t.Errorf("DetectImageType() = %v, want %v", result, ImageTypeWebP)
			}
		})
	}

	// Test AVIF files from testdata if they exist
	avifFiles := []string{
		"testdata/avif/lossy/quality_80.avif",
		"testdata/avif/lossless/original.avif",
	}

	for _, path := range avifFiles {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("AVIF test file %s not found", path)
				return
			}

			result, err := DetectImageType(path)
			if err != nil {
				t.Fatalf("DetectImageType() error = %v", err)
			}

			if result != ImageTypeAVIF {
				t.Errorf("DetectImageType() = %v, want %v", result, ImageTypeAVIF)
			}
		})
	}
}

// Benchmark image type detection
func BenchmarkDetectImageTypeFromBytes(b *testing.B) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectImageTypeFromBytes(jpegData)
	}
}

func BenchmarkDetectImageTypeFromReader(b *testing.B) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(jpegData)
		_, _ = DetectImageTypeFromReader(reader)
	}
}

// TestReaderReuse tests that we can detect type without consuming the entire reader
func TestReaderReuse(t *testing.T) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	
	// Create a reader that tracks how much was read
	reader := &limitedReader{
		Reader: bytes.NewReader(jpegData),
		limit:  32, // Our detection reads up to 32 bytes
	}
	
	imgType, err := DetectImageTypeFromReader(reader)
	if err != nil {
		t.Fatalf("DetectImageTypeFromReader() error = %v", err)
	}
	
	if imgType != ImageTypeJPEG {
		t.Errorf("DetectImageTypeFromReader() = %v, want %v", imgType, ImageTypeJPEG)
	}
	
	if reader.bytesRead > 32 {
		t.Errorf("Read %d bytes, expected <= 32", reader.bytesRead)
	}
}

type limitedReader struct {
	io.Reader
	limit     int
	bytesRead int
}

func (r *limitedReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.bytesRead += n
	return n, err
}