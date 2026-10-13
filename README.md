# Tengeneza CV Desktop

A Wails desktop application that combines the Tengeneza CV React client and Go API in one monorepo. The existing browser API remains available as a separate service; the desktop app starts its own local API and never requires sign-in.

## What is included

- **No account or authentication flow** in the desktop UI. Each OS user gets one local workspace.
- **Local persistence** for profiles, CVs, versions, and export records. Workspace data is stored under the current user's configuration directory in `Tengeneza CV/workspace.json`; the file and directories are private to that OS user.
- **All 24 templates and sample CVs** are available locally.
- **Unlimited local CV creation, versions, ATS checks, and PDF/DOCX/CSV exports.** The hosted API's existing daily export limit remains unchanged.
- A loopback API binds to `127.0.0.1` on a random port and is stopped when the desktop window closes. It is not exposed as a network service.
- No subscription checkout or payment flow is included.

AI drafting uses the existing Ngamia AI integration. It is optional and requires a configured `NGAMIA_AI_API_KEY`/service endpoint; without that provider configuration, the editor and local CV creation/export features still work. No AI provider key is bundled in the desktop app.

## Repository layout

- `main.go`, `internal/desktop`: Wails host and local service lifecycle.
- `internal/`: the Go API, rendering pipeline, and stores.
- `frontend/`: the existing React/Vite client adapted for the Wails bridge.
- `examples/`, `uploads/`, `migrations/`: local fixtures, assets, and hosted-service schema.
- `docs/source/api-README.md`: the original API README retained from its source repository.

The two source checkouts remain separate and unchanged: `/workspace/Tengeneza_CV-API` and `/workspace/Tengeneza_CV-UI`.

## Requirements

- Go 1.26 or newer, Node.js 22+, pnpm, and Wails v2.16.
- Linux: GTK 3 and WebKit2GTK 4.0 development libraries required by Wails. Ubuntu 24.04 in this workspace exposes WebKitGTK 4.1 instead; its native build was verified with a temporary local pkg-config compatibility alias, not a repository change. On other Linux distributions, install the WebKitGTK version expected by the installed Wails CLI.
- Chromium, Google Chrome, or Microsoft Edge for PDF rendering. The app searches common install locations and also accepts `CHROMIUM_EXECUTABLE`. DOCX and CSV export do not need a browser executable.

## Run in development

```bash
pnpm --dir frontend install
pnpm --dir frontend vite:build
go run .
```

To use Wails' live development window, install the Wails CLI and run `wails dev` from this directory. Build a packaged application with `wails build`.

## Tests and checks

```bash
gofmt -w main.go internal
go test ./...
pnpm --dir frontend check
pnpm --dir frontend vite:build
```

## Local data and backups

To back up the workspace, close the app and copy `workspace.json` from the OS configuration directory's `Tengeneza CV` folder. The JSON file contains personal CV data; protect backups accordingly. Deleting that file resets the local workspace. Sample fixtures and profile images are copied into the same per-user app-data folder on first launch.

## Source provenance

This project combines the cloned repositories in `/workspace/Tengeneza_CV-API` and `/workspace/Tengeneza_CV-UI`. Their source checkout contents were copied into this new repository; the API and UI directories were not modified during bootstrap.
