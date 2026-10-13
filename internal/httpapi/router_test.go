package httpapi

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/config"
	"io"
	"log/slog"
)

func TestHealthzUsesCanonicalEnvelope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := New(config.Config{AllowedOrigins: []string{"http://localhost:3000"}}, logger)
	response, err := app.Test(httptest.NewRequest("GET", "/healthz", nil))
	require.NoError(t, err)
	require.Equal(t, 200, response.StatusCode)
}

func TestExportsRequireAuthentication(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := New(config.Config{AllowedOrigins: []string{"http://localhost:3000"}}, logger)
	response, err := app.Test(httptest.NewRequest("GET", "/exports/private_cv-user.csv", nil))
	require.NoError(t, err)
	require.Equal(t, 404, response.StatusCode, "anonymous/demo identity cannot retrieve a private artifact")
}
