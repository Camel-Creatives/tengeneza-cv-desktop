# Document Rendering System

The rendering subsystem creates visually distinct, professional CV outputs from the canonical CV JSON document. It produces searchable PDF, editable DOCX, and spreadsheet-friendly CSV artifacts from the same normalized content. [1]

## Pipeline

```text
domain.CV content
      │
      ▼
model.Normalize
      │ typed document model
      ▼
optional local-photo embedding
      │
      ├─────────────► html.Renderer → Chromium → PDF artifact
      ├─────────────► docx.Renderer → DOCX artifact
      └─────────────► encoding/csv → CSV artifact
```

The orchestrator is [`service.go`](./service.go). It selects `pdf`, `docx`, or `csv`, renders to a temporary/local working path, and returns path, filename, MIME type, and byte count. The HTTP handler uploads the resulting bytes to the configured object store and records metadata in PostgreSQL; local filesystem storage is a development fallback. [1]

## Responsibilities by package

| Package | Responsibility |
|---|---|
| [`model`](./model) | Normalize loosely typed CV JSON into a typed rendering document and create safe file stems. |
| [`html`](./html) | Render semantic HTML and CSS for each template ID. |
| [`pdf`](./pdf) | Print HTML to a searchable PDF through Chromium. |
| [`docx`](./docx) | Generate native editable DOCX output. |
| `renderCSV` | Generate RFC-compliant, spreadsheet-friendly CSV output including structured CV sections. |
| [`lint`](./lint) | Score deterministic ATS-safe checks against canonical CV data. |
| [`service.go`](./service.go) | Coordinate template lookup, photo embedding, rendering choice, and local artifact persistence. |

## Template design rules

The template library contains 24 active designs. Every design must remain a distinct document system: different composition, information hierarchy, typography, and visual grammar rather than a color-only reskin. The design inventory and fictional samples live in [`../../docs/template_library_delivery.md`](../../docs/template_library_delivery.md). [2]

When adding a template:

1. Add a catalog entry in `internal/store/memory.go` with a stable ID, name, family, photo support, active state, and accent color.
2. Add a matching dispatch case and a separate semantic document system in `html/renderer.go`.
3. Add one fictional JSON fixture in `examples/sample-cvs` with quantified bullets and an appropriate target region.
4. Register photo-capable IDs in `cmd/render-samples/main.go` and in the sample-fixture lint test.
5. Render the full fixture set, confirm the new PDF has a clean reading order and no collisions, and record visual review findings.

## Profile-photo handling

The document model accepts a photo URL only when `settings.show_photo` is enabled. Before HTML rendering, the service recognizes `file://` paths, `/uploads/profiles/` paths, and absolute local paths. It loads the file only when it is no larger than 5 MB and its detected MIME type is JPEG, PNG, or WebP; eligible files are inlined as a data URI so Chromium can render them reliably. [1]

This behavior is intended for local rendering and trusted, application-owned upload paths. Do not introduce arbitrary remote-URL fetching into the renderer.

## Local verification

Generate the full sample library with:

```bash
go run ./cmd/render-samples \
  -input examples/sample-cvs \
  -output /tmp/tengenezacv-rendered-samples
```

The command should write one PDF and one DOCX for every fixture. Review output metadata and inspect representative PDFs after a layout change:

```bash
pdfinfo /tmp/tengenezacv-rendered-samples/Amina_Juma_CV.pdf
```

Run code-quality checks before committing:

```bash
go test ./...
go vet ./...
go build ./cmd/api ./cmd/worker ./cmd/render-samples
```

The fictional templates were validated as single-page A4 samples during the current 24-template delivery pass. The QA record is retained in [`../../docs/template_library_visual_qa.md`](../../docs/template_library_visual_qa.md). [3]

## Production considerations

Chromium must exist at `CHROMIUM_EXECUTABLE`; the API Docker image installs Chromium and fonts. HTTP exports currently render synchronously in the API request path (there is no worker queue yet), then persist bytes to S3-compatible storage. Downloads use an authenticated owner-scoped API route; the bucket should remain private. `LOCAL_EXPORT_DIR` is only the working directory/development fallback. Configure object retention according to the product retention policy.

## Implementation references

[1] [`service.go`](./service.go) — Render orchestration, formats, artifacts, and profile-photo embedding.  
[2] [`../../docs/template_library_delivery.md`](../../docs/template_library_delivery.md) — 24-template catalog and verification record.  
[3] [`../../docs/template_library_visual_qa.md`](../../docs/template_library_visual_qa.md) — Visual review findings and remediations.
