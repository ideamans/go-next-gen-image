package nextgenimage

import (
	"os"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestMain(m *testing.M) {
	// Initialize vips once for all tests
	vips.Startup(nil)

	// Run tests
	code := m.Run()

	// Clean up
	vips.Shutdown()

	os.Exit(code)
}
