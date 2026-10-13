package pdf

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderPDFWithConfiguredChromium(t *testing.T) {
	executable := os.Getenv("CHROMIUM_EXECUTABLE")
	if executable == "" {
		t.Skip("CHROMIUM_EXECUTABLE is not configured")
	}
	if _, err := os.Stat(executable); err != nil {
		t.Skipf("configured Chromium is unavailable: %v", err)
	}
	raw, err := (Renderer{Executable: executable}).Render(context.Background(), "<!doctype html><html><body><h1>Export smoke test</h1></body></html>")
	require.NoError(t, err)
	require.Greater(t, len(raw), 100)
	require.Equal(t, "%PDF-", string(raw[:5]))
}
