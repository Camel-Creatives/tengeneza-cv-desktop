package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/config"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/store"
)

func TestDesktopWorkspaceNeedsNoLoginAndHasUnlimitedExports(t *testing.T) {
	temp := t.TempDir()
	localStore, err := store.OpenLocalStore(filepath.Join(temp, "workspace.json"))
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app, err := NewWithRepository(config.Config{
		Environment:      "desktop",
		AllowedOrigins:   []string{"http://wails.localhost"},
		PublicAPIBaseURL: "http://127.0.0.1:41234",
		LocalExportDir:   filepath.Join(temp, "exports"),
		LocalUploadDir:   filepath.Join(temp, "uploads"),
		ExamplesDir:      filepath.Join(temp, "examples"),
	}, logger, localStore)
	require.NoError(t, err)

	walletResponse, err := app.Test(httptest.NewRequest(http.MethodGet, "/me/wallet", nil))
	require.NoError(t, err)
	defer walletResponse.Body.Close()
	require.Equal(t, http.StatusOK, walletResponse.StatusCode)
	var wallet struct {
		Success bool `json:"success"`
		Data    struct {
			DailyExportLimit      int `json:"daily_export_limit"`
			DailyExportsRemaining int `json:"daily_exports_remaining"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(walletResponse.Body).Decode(&wallet))
	require.True(t, wallet.Success)
	require.Equal(t, 0, wallet.Data.DailyExportLimit)
	require.Equal(t, -1, wallet.Data.DailyExportsRemaining)

	request := httptest.NewRequest(http.MethodPost, "/cvs", bytes.NewBufferString(`{"title":"My CV","template_id":"ats_modern","content":{"personal":{"full_name":"Test User"},"summary":"Builder"}}`))
	request.Header.Set("Content-Type", "application/json")
	created, err := app.Test(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, created.StatusCode, "local users create CVs without bearer tokens")
	var createdCV struct {
		Data struct {
			ID string `json:"ID"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(created.Body).Decode(&createdCV))
	created.Body.Close()
	require.NotEmpty(t, createdCV.Data.ID)

	exportRequest := httptest.NewRequest(http.MethodPost, "/cvs/"+createdCV.Data.ID+"/export", bytes.NewBufferString(`{"format":"csv"}`))
	exportRequest.Header.Set("Content-Type", "application/json")
	exported, err := app.Test(exportRequest)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, exported.StatusCode)
	var exportResult struct {
		Data struct {
			Filename string `json:"filename"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(exported.Body).Decode(&exportResult))
	exported.Body.Close()
	require.NotEmpty(t, exportResult.Data.Filename)

	download, err := app.Test(httptest.NewRequest(http.MethodGet, "/exports/"+exportResult.Data.Filename, nil))
	require.NoError(t, err)
	defer download.Body.Close()
	require.Equal(t, http.StatusOK, download.StatusCode, "local export artifacts remain downloadable after render temp cleanup")
	require.Contains(t, download.Header.Get("Content-Type"), "text/csv")
	content, err := io.ReadAll(download.Body)
	require.NoError(t, err)
	require.Contains(t, string(content), "Test User")
	require.FileExists(t, filepath.Join(temp, "exports", exportResult.Data.Filename))
	renderTemps, err := os.ReadDir(filepath.Join(temp, "exports", ".render"))
	require.NoError(t, err)
	require.Empty(t, renderTemps, "temporary renderer files are removed separately")
}
