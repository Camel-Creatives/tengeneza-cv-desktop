# Tengeneza CV Platform API

**Tengeneza CV Platform API** is a Tanzania-first, structured-CV service for creating, revising, linting, exporting, and billing for professional CVs. It exposes a REST interface for authenticated CV workflows, deterministic ATS checks, profile-photo handling, PDF/DOCX rendering, AI-assisted section drafting, and mobile-money billing integration.

> **Repository scope.** This repository owns the backend API and document-rendering workflow. The API serves its embedded contract from [`internal/httpapi/openapi.yaml`](./internal/httpapi/openapi.yaml) at `/docs/openapi.yaml`; the concise repository overview is [`docs/openapi.yaml`](./docs/openapi.yaml). Keep both in sync when route behavior changes. [1]

## Contents

| Topic | Purpose |
|---|---|
| [Quick start](#quick-start) | Run the API and local service dependencies. |
| [System architecture](#system-architecture) | Understand process boundaries and the current implementation topology. |
| [HTTP API](#http-api) | Discover authentication, CV, export, AI, billing, and webhook routes. |
| [CV data model](#cv-data-model) | Integrate frontend forms with the canonical CV JSON document. |
| [Templates and rendering](#templates-and-rendering) | Work with the 24-template library and PDF/DOCX exports. |
| [Configuration](#configuration) | Configure local development and production-required environment variables. |
| [Security and operations](#security-and-operations) | Apply request, authentication, webhook, and deployment controls. |
| [Repository documentation](#repository-documentation) | Navigate focused subsystem documentation. |

## Product capabilities

| Capability | What the API provides |
|---|---|
| Authentication | CamelAccounts OAuth sign-in and verified JWT claims mapped to durable internal user records. Legacy API signup/password routes respond that CamelAccounts is required. |
| Profile management | User profile updates, preferred target region and language, and validated JPEG/PNG/WebP profile-photo upload. |
| CV lifecycle | Create, list, read, patch, archive, duplicate, inspect immutable versions, and restore an earlier version. |
| Template library | Eight baseline templates plus sixteen industry-specific visual templates, for a total of 24 active designs. |
| Document intelligence | Deterministic ATS linting and an AI gateway for section-level drafting. |
| Exports | Searchable Chromium PDFs, native editable DOCX files, and spreadsheet-friendly CSV from the canonical CV document. New artifact creation is limited to five per Tanzania-local calendar day; completed matching artifacts can be reused. |
| Billing | Plan discovery and wallet allowance configuration. |
| Notifications | MailAfrica client and signed webhook boundary for transactional-email workflows. |

The request router, supported methods, and security middleware are assembled in [`internal/httpapi/router.go`](./internal/httpapi/router.go). [2]

## Quick start

### Prerequisites

The module declares **Go 1.27.0**. Local dependency services are defined in Docker Compose, so Docker Desktop or a compatible Docker Engine is also required. Real PDF export requires a Chromium-compatible executable; the development default is `/usr/bin/chromium`. [3]

| Dependency | Used for | Local default |
|---|---|---|
| Go | API, worker, tests, and render-sample command | `go 1.27.0` |
| PostgreSQL 16 | Persistent users, workspaces, CVs, versions, and export quota | `localhost:5434` |
| Redis 7 | Distributed API rate limiting | `localhost:6380` |
| MinIO / S3-compatible storage | Persistent exports and profile photos | API `9000`, console `9001` |
| Chromium | HTML-to-PDF rendering | `/usr/bin/chromium` |

### First local run

```bash
git clone https://github.com/troubleman96/tengenezacv-api.git
cd tengenezacv-api

cp .env.example .env
docker compose up -d postgres redis minio

make tidy
make run
```

In a second terminal, confirm the service is reachable:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/readyz
open http://localhost:8080/docs
```

The default API port is `8080`. The interactive Swagger page is served at `/docs`, and the versioned OpenAPI YAML is served at `/docs/openapi.yaml`. [2]

### Everyday development commands

| Command | Result |
|---|---|
| `make run` | Start the Fiber API process. |
| `make worker` | Start the worker entrypoint. |
| `make test` | Run the Go test suite. |
| `make lint` | Run `go vet ./...`. |
| `make build` | Build API and worker commands. |
| `make docker-up` | Start PostgreSQL, Redis, and MinIO. |
| `go run ./cmd/render-samples -input examples/sample-cvs -output /tmp/tengenezacv-rendered-samples` | Generate PDF and DOCX output for every fictional template fixture. |

## System architecture

```text
Web / Mobile Frontend
        │ HTTPS + Bearer JWT
        ▼
┌──────────────────────────┐
│ Fiber API (`cmd/api`)    │
│ auth · CVs · templates   │
│ lint · exports · billing │
└───────┬──────────┬───────┘
        │          │
        │          ├──────────────► Ngamia AI gateway
        │          └──────────────► MailAfrica client + signed webhook
        │
        ▼
┌──────────────────────────┐
│ Render service           │
│ canonical CV JSON        │
│ semantic HTML → PDF      │
│ native DOCX writer       │
└──────────────────────────┘
```

The API uses PostgreSQL for accounts, workspaces, CVs, version history, templates, export records, and atomic Tanzania-local daily quota accounting. Redis backs distributed rate limits. S3-compatible storage persists exports and profile photos, with owner-authorized downloads. Development without these services explicitly falls back to process memory and local files; production startup rejects missing or unavailable PostgreSQL, Redis, or object storage. Schema upgrades are embedded and applied under a PostgreSQL advisory lock at startup. HTTP exports render synchronously in the API, including PDF rendering via the Chromium-enabled API image.

| Process | Entry point | Primary responsibility |
|---|---|---|
| API | `cmd/api` | HTTP transport, auth, validation, template/CV workflows, OpenAPI, and webhooks. |
| Worker | `cmd/worker` | Bootstrap placeholder only; it does not currently consume jobs. |
| Render samples | `cmd/render-samples` | Generate local PDF/DOCX verification outputs for every JSON fixture. |
| Render service | `internal/render` | Normalize the canonical document, inline local photos, and create PDF or DOCX artifacts. |

## HTTP API

Every success response uses a stable envelope with `success`, `data`, and `meta`; every error response uses `success`, `error`, and `meta`. The `meta` object contains the request identifier and server timestamp. Supply `X-Request-ID` to correlate a frontend action with server logs, or let the server generate one. [5]

```json
{
  "success": true,
  "data": { "status": "ok" },
  "meta": {
    "request_id": "c1e6c671-1c4e-4ee8-93ed-8aaf52fcb67a",
    "timestamp": "2026-08-21T10:30:00Z"
  }
}
```

### Public routes

| Method | Route | Use |
|---|---|---|
| `GET` | `/healthz` | Liveness probe. |
| `GET` | `/readyz` | Readiness probe that checks PostgreSQL, Redis, and object storage; returns `503` on dependency failure. |
| `GET` | `/metrics` | Text metrics placeholder. |
| `GET` | `/docs` | Swagger UI. |
| `GET` | `/docs/openapi.yaml` | OpenAPI 3.1 document. |
| `POST` | `/auth/signup` | Compatibility endpoint; returns `CAMELACCOUNTS_REQUIRED`. |
| `POST` | `/auth/otp/verify` | Compatibility endpoint; returns `CAMELACCOUNTS_REQUIRED`. |
| `POST` | `/auth/login/password` | Compatibility endpoint; returns `CAMELACCOUNTS_REQUIRED`. |
| `POST` | `/auth/admin/login` | Sign in with the configured administrator email and password; returns a superadmin bearer token. |
| `GET` | `/billing/plans` | List available plans. |
| `POST` | `/webhooks/mailafrica` | Receive signed MailAfrica events. |

### Authenticated routes

All routes below require `Authorization: Bearer <access_token>`.

| Domain | Methods and routes |
|---|---|
| Account | `GET/PATCH /me`, `POST /me/profile-photo`, `GET /me/wallet` |
| Templates | `GET /templates` |
| CV lifecycle | `POST/GET /cvs`, `GET/PATCH/DELETE /cvs/:id`, `POST /cvs/:id/duplicate` |
| Version control | `GET /cvs/:id/versions`, `POST /cvs/:id/versions/:versionID/restore` |
| Profile photo on a CV | `POST /cvs/:id/use-profile-photo` |
| Guidance and exports | `POST /cvs/:id/agent/ats-score`, `POST /cvs/:id/agent/autofill-section`, `POST /cvs/:id/export` |
| Billing | `POST /billing/checkout`, `GET /billing/sessions/:ref/status` |

For payload schemas and response codes, use the embedded OpenAPI contract. Note that the authentication endpoint descriptions there are compatibility routes: end-user sign-in is provided by CamelAccounts OAuth in the UI.

### Authentication workflow

1. End users authenticate through CamelAccounts. The UI performs the OAuth authorization-code flow with PKCE and returns to its `/auth/callback` route. The API verifies the CamelAccounts JWT using its configured issuer, audience, and JWKS endpoint. On the first authenticated API request, it creates or retrieves a durable user row keyed by the verified external subject and syncs available name, email, and phone claims. Send the access token as a bearer token for protected API calls.

The API `/auth/signup`, `/auth/otp/verify`, `/auth/otp/resend`, and `/auth/login/password` routes remain for compatibility but return `CAMELACCOUNTS_REQUIRED`; do not integrate new end-user flows with them. Configure `CAMEL_ACCOUNTS_AUDIENCE` to the registered public OAuth client ID used by the UI (default `tengenezacv-web`) so the API accepts the access tokens issued for that client.

## CV data model

The API stores a canonical JSON `content` object inside each CV. This makes editing, linting, template switching, versioning, PDF rendering, and DOCX rendering operate from the same source document.

```json
{
  "title": "Amina Juma — Finance CV",
  "template_id": "finance_audit",
  "document_style": "resume",
  "target_region": "eac",
  "language": "en",
  "content": {
    "personal": {
      "full_name": "Amina Juma",
      "headline": "Financial Analyst",
      "email": "amina@example.test",
      "phone": "+255700000000",
      "location": "Dar es Salaam, Tanzania"
    },
    "summary": "A concise outcome-oriented professional summary.",
    "experience": [],
    "education": [],
    "certifications": [],
    "skills": {
      "hard_skills": [],
      "soft_skills": [],
      "languages": []
    },
    "settings": {
      "show_photo": false,
      "show_referees": false
    }
  }
}
```

`PATCH /cvs/:id` accepts an incremental `content` object and records a version with an optional human-readable `change_summary`. `DELETE /cvs/:id` archives rather than hard-deletes the document. Production records are stored in PostgreSQL and every CV query is scoped to the authenticated internal user ID.

## Templates and rendering

The database migration seeds the active catalog with **24 differentiated designs**, including ATS-focused baseline designs and industry-led systems for creative work, product, finance, law, clinical care, education, sales, operations, consulting, engineering, public service, nonprofit programmes, graduates, trades, remote work, and creative production. Sample fixtures and generated outputs are documented in [`docs/template_library_delivery.md`](./docs/template_library_delivery.md). [7]

### Export workflow

1. The client selects `pdf`, `docx`, or `csv` and calls `POST /cvs/:id/export`.
2. The server validates the CV and selected template, then atomically reserves one of five Tanzania-local daily exports, or reuses a matching artifact.
3. The render service normalizes the canonical CV JSON into a typed document model.
4. Local profile photos are validated and embedded as base64 data URIs for reliable Chromium rendering.
5. PDF rendering uses semantic HTML and Chromium; DOCX rendering uses native OOXML generation.
6. The completed artifact is stored in S3-compatible storage and downloaded via an authenticated owner-scoped API route. Matching exports reuse the stored artifact without another daily slot. [4] [6]

For renderer behavior, custom-template development, photo constraints, and the sample command, see [`internal/render/README.md`](./internal/render/README.md).

## Configuration

Copy `.env.example` to `.env` for local development. Values in the first group are needed by the HTTP process; provider and storage values are required in production validation. [8]

| Group | Variables |
|---|---|
| Runtime | `APP_ENV`, `PORT`, `PUBLIC_API_BASE_URL`, `ALLOWED_ORIGINS`, `LOG_LEVEL` |
| Identity | `JWT_SIGNING_KEY`, `JWT_ISSUER` |
| Admin sign-in (optional; both required to enable) | `ADMIN_LOGIN_EMAIL`, `ADMIN_LOGIN_PASSWORD` (minimum 12 characters) |
| Persistent data and coordination | `DATABASE_URL`, `REDIS_URL` |
| AI | `NGAMIA_AI_BASE_URL`, `NGAMIA_AI_API_KEY` |
| MailAfrica | `MAILAFRICA_BASE_URL`, `MAILAFRICA_API_KEY`, `MAILAFRICA_WEBHOOK_SECRET`, `MAILAFRICA_DEFAULT_FROM` |
| Object storage | `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` |
| Local development paths | `LOCAL_EXPORT_DIR`, `LOCAL_UPLOAD_DIR`, `CHROMIUM_EXECUTABLE` |

In production, startup requires non-empty database, Redis, JWT, AI, provider-webhook, and S3 credentials. `JWT_SIGNING_KEY` must have at least 32 characters. [8]

## Security and operations

The router applies panic recovery, request IDs, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, restricted CORS methods/headers, and a Redis-backed shared limit of 120 requests per minute (process-memory fallback in development). [2]

| Control | Integration requirement |
|---|---|
| Passwords | Passwords are hashed with Argon2id through the auth package. |
| Auth | Protected calls must carry a valid bearer token. |
| CORS | Set `ALLOWED_ORIGINS` to explicit frontend origins; do not use a permissive wildcard for authenticated clients. |
| Webhooks | Validate provider signatures over the raw request body and preserve replay-protection behavior. |
| Idempotency | Send `Idempotency-Key` for checkout creation and retain the same key across a safe client retry. |
| Profile images | Send multipart field `photo`; accepted MIME types are JPEG, PNG, and WebP, with a 5 MB render-service guard for local embedding. |
| Logging | Do not log CV content, tokens, passwords, or provider secrets. Use request IDs for support correlation. |

The detailed production checklist is in [`docs/operations.md`](./docs/operations.md), while third-party boundaries are defined in [`docs/third_party_contracts.md`](./docs/third_party_contracts.md).

## Repository documentation

| Document | Audience | Description |
|---|---|---|
| [`docs/README.md`](./docs/README.md) | All contributors | Documentation map, source-of-truth policy, and reading paths. |
| [`internal/httpapi/README.md`](./internal/httpapi/README.md) | Frontend and API integrators | Auth, envelopes, endpoints, errors, CORS, idempotency, and webhooks. |
| [`internal/render/README.md`](./internal/render/README.md) | Rendering and template contributors | PDF/DOCX pipeline, photo handling, template conventions, and verification. |
| [`examples/README.md`](./examples/README.md) | Frontend and QA teams | Fictional fixtures, portrait assets, sample render commands, and naming rules. |
| [`docs/openapi.yaml`](./docs/openapi.yaml) | Client-generation tools | Versioned OpenAPI contract. |

## Quality checks

Run the following before opening a pull request or pushing a release candidate:

```bash
go test ./...
go vet ./...
go build ./cmd/api ./cmd/worker ./cmd/render-samples
go run ./cmd/render-samples -input examples/sample-cvs -output /tmp/tengenezacv-rendered-samples
```

The sample-render command should produce one PDF and one DOCX for every fixture. The 24-template library was visually reviewed with two layout defects corrected before the verification record was published. [7]

## Implementation references

[1] [`docs/openapi.yaml`](./docs/openapi.yaml) — Versioned OpenAPI contract.
[2] [`internal/httpapi/router.go`](./internal/httpapi/router.go) — Route map, middleware, and dependency wiring.
[3] [`go.mod`](./go.mod), [`docker-compose.yml`](./docker-compose.yml), and [`Makefile`](./Makefile) — Toolchain and local dependencies.
[4] [`internal/render/service.go`](./internal/render/service.go) — Export orchestration and local photo embedding.
[5] [`internal/apiresponse/response.go`](./internal/apiresponse/response.go) — Canonical response envelopes.
[6] [`internal/httpapi/handlers/core.go`](./internal/httpapi/handlers/core.go) — Authentication, CV, profile, and export behavior.
[7] [`docs/template_library_delivery.md`](./docs/template_library_delivery.md) — 24-template verification and delivery manifest.
[8] [`internal/config/config.go`](./internal/config/config.go) and [`.env.example`](./.env.example) — Configuration defaults and production validation.
