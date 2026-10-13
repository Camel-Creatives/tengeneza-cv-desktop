package desktop

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/config"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/httpapi"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/store"
)

// DesktopApp owns the local-only services behind the Wails window.
type DesktopApp struct {
	assets   fs.FS
	samples  fs.FS
	logger   *slog.Logger
	mu       sync.RWMutex
	api      *fiber.App
	listener net.Listener
	store    *store.LocalStore
	apiURL   string
	startErr error
}

func New(assets, samples fs.FS, logger *slog.Logger) *DesktopApp {
	if logger == nil {
		logger = slog.Default()
	}
	return &DesktopApp{assets: assets, samples: samples, logger: logger}
}

// Startup creates a private, per-user workspace and starts the existing API
// only on loopback. The API's development auth bypass is therefore restricted
// to this local desktop process; no external account or auth provider is used.
func (a *DesktopApp) Startup(context.Context) {
	if err := a.start(); err != nil {
		a.mu.Lock()
		a.startErr = err
		a.mu.Unlock()
		a.logger.Error("desktop services failed to start", "error", err)
	}
}

func (a *DesktopApp) start() error {
	base, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find user data directory: %w", err)
		}
		base = filepath.Join(base, ".config")
	}
	appDir := filepath.Join(base, "Tengeneza CV")
	paths := map[string]os.FileMode{
		appDir:                           0700,
		filepath.Join(appDir, "exports"): 0700,
		filepath.Join(appDir, "uploads", "profiles"): 0700,
		filepath.Join(appDir, "examples"):            0700,
	}
	for path, mode := range paths {
		if err := os.MkdirAll(path, mode); err != nil {
			return fmt.Errorf("create local app directory: %w", err)
		}
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("secure local app directory: %w", err)
		}
	}

	examplesDir := filepath.Join(appDir, "examples", "sample-cvs")
	if err := copyEmbeddedTree(a.samples, ".", examplesDir); err != nil {
		return fmt.Errorf("install sample CVs: %w", err)
	}
	profileAssets := filepath.Join(appDir, "uploads", "profiles")
	if err := copyEmbeddedTree(a.assets, "uploads/profiles", profileAssets); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("install sample profile images: %w", err)
	}

	localStore, err := store.OpenLocalStore(filepath.Join(appDir, "workspace.json"))
	if err != nil {
		return fmt.Errorf("open local workspace: %w", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("reserve private API port: %w", err)
	}
	apiURL := "http://" + listener.Addr().String()
	cfg := config.Config{
		Environment:      "desktop",
		Port:             strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		PublicAPIBaseURL: apiURL,
		AllowedOrigins: []string{
			"http://wails.localhost", "https://wails.localhost",
			"http://localhost:34115", "http://localhost:3000", "http://127.0.0.1:3000",
		},
		LocalExportDir:     filepath.Join(appDir, "exports"),
		LocalUploadDir:     filepath.Join(appDir, "uploads"),
		ExamplesDir:        examplesDir,
		ChromiumExecutable: findChromium(),
		NgamiaAIBaseURL:    "https://api.ngamia.cc/v1",
		MailAfricaBaseURL:  "https://api.mailafrica.online",
	}
	api, err := httpapi.NewWithRepository(cfg, a.logger, localStore)
	if err != nil {
		_ = listener.Close()
		return fmt.Errorf("configure local API: %w", err)
	}

	a.mu.Lock()
	a.api = api
	a.listener = listener
	a.store = localStore
	a.apiURL = apiURL
	a.mu.Unlock()
	go func() {
		if err := api.Listener(listener); err != nil && !errors.Is(err, net.ErrClosed) {
			a.logger.Error("local API stopped", "error", err)
		}
	}()
	if err := waitForAPI(apiURL, 3*time.Second); err != nil {
		return fmt.Errorf("local API did not become ready: %w", err)
	}
	a.logger.Info("local desktop workspace ready", "data_directory", appDir, "api_address", listener.Addr().String(), "chromium_available", cfg.ChromiumExecutable != "")
	return nil
}

func (a *DesktopApp) Shutdown(context.Context) {
	a.mu.Lock()
	api, listener := a.api, a.listener
	a.apiURL = ""
	a.mu.Unlock()
	if api != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := api.ShutdownWithContext(ctx); err != nil {
			a.logger.Warn("local API shutdown was not clean", "error", err)
		}
	}
	if listener != nil {
		_ = listener.Close()
	}
}

func (a *DesktopApp) APIBaseURL() (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.startErr != nil {
		return "", a.startErr
	}
	if a.apiURL == "" {
		return "", errors.New("the local workspace is not ready")
	}
	return a.apiURL, nil
}

func waitForAPI(base string, timeout time.Duration) error {
	client := &http.Client{Timeout: 300 * time.Millisecond}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		response, err := client.Get(base + "/healthz")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	return errors.New("timed out waiting for local service")
}

func copyEmbeddedTree(source fs.FS, sourceDir, targetDir string) error {
	return fs.WalkDir(source, sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		target := targetDir
		if rel != "." {
			target = filepath.Join(targetDir, rel)
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		if _, err := os.Stat(target); err == nil {
			return nil
		}
		raw, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0600)
	})
}

func findChromium() string {
	if configured := strings.TrimSpace(os.Getenv("CHROMIUM_EXECUTABLE")); configured != "" {
		return configured
	}
	for _, candidate := range []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable", "msedge", "chrome"} {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"}
	case "windows":
		for _, root := range []string{os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)"), os.Getenv("LOCALAPPDATA")} {
			if root != "" {
				candidates = append(candidates, filepath.Join(root, "Microsoft", "Edge", "Application", "msedge.exe"), filepath.Join(root, "Google", "Chrome", "Application", "chrome.exe"))
			}
		}
	default:
		candidates = []string{"/usr/bin/chromium", "/usr/bin/chromium-browser", "/usr/bin/google-chrome"}
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}
