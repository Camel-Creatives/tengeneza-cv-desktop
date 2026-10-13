# HTTP API Integration Guide

This guide is for web, mobile, and SDK contributors integrating with the Tengeneza CV Platform API. The route implementation is assembled in [`router.go`](./router.go); the machine-readable contract is [`openapi.yaml`](./openapi.yaml). [1]

## Base URL and interactive contract

Use the environment-specific API base URL. Local development defaults to `http://localhost:8080`; production configuration uses `PUBLIC_API_BASE_URL`. Swagger UI is available at `/docs` and OpenAPI YAML at `/docs/openapi.yaml`. [1] [2]

```text
GET  {API_BASE_URL}/healthz
GET  {API_BASE_URL}/docs
GET  {API_BASE_URL}/docs/openapi.yaml
```

## Request conventions

| Header | When to send it | Notes |
|---|---|---|
| `Authorization: Bearer <access_token>` | All protected routes | Obtain an access token from CamelAccounts OAuth; the API validates its issuer, audience, and signature. |
| `Content-Type: application/json` | JSON routes | Required for standard request bodies. |
| `Idempotency-Key: <stable-uuid>` | Checkout creation | Reuse the key on a retry of the same client action. |
| `X-Request-ID: <client-correlation-id>` | Recommended | The API echoes or generates a request ID in the response metadata. |

For profile photos, use `multipart/form-data` and attach the binary file under the `photo` field rather than sending JSON.

## Response envelopes

All successful requests return the following shape:

```json
{
  "success": true,
  "data": {},
  "meta": {
    "request_id": "uuid",
    "timestamp": "RFC3339 UTC timestamp"
  }
}
```

Failures return an error code and user-safe message:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Provide valid CV details."
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "RFC3339 UTC timestamp"
  }
}
```

The canonical envelope is implemented in [`../apiresponse/response.go`](../apiresponse/response.go). [3]

## Authentication

### CamelAccounts OAuth

End-user authentication is owned by CamelAccounts. The UI offers CamelAccounts email/password and configured social sign-in, then uses OAuth authorization code with PKCE and sends the returned access token to the API. The API verifies the token against `CAMEL_ACCOUNTS_ISSUER`, `CAMEL_ACCOUNTS_AUDIENCE`, and `CAMEL_ACCOUNTS_JWKS_URL`. `CAMEL_ACCOUNTS_AUDIENCE` must match the public OAuth client ID configured in the UI (default `tengenezacv-web`). On authenticated requests, the API provisions a PostgreSQL user keyed by the external subject and syncs available name, email, and phone claims.

The API `/auth/signup`, `/auth/otp/verify`, `/auth/otp/resend`, and `/auth/login/password` routes are compatibility endpoints and return `CAMELACCOUNTS_REQUIRED`. Admin password login is separate and configured with `ADMIN_LOGIN_EMAIL` and `ADMIN_LOGIN_PASSWORD`.

Keep access tokens in the approved secure-session strategy. Never place provider keys, signing keys, raw CV content, or tokens into browser logs.

## Endpoint groups

| Group | Routes | Frontend implementation note |
|---|---|---|
| Health and docs | `GET /healthz`, `/readyz`, `/metrics`, `/docs`, `/docs/openapi.yaml` | `/healthz` is liveness; `/readyz` pings PostgreSQL, Redis, and storage, returning 503 if any check fails. `/metrics` is currently a placeholder. |
| Identity | CamelAccounts OAuth; `POST /auth/admin/login` | End-user access tokens must satisfy configured issuer/audience/JWKS checks; admin password login uses `ADMIN_LOGIN_EMAIL` and `ADMIN_LOGIN_PASSWORD`. Legacy signup/OTP/password endpoints return `CAMELACCOUNTS_REQUIRED`. |
| User | `GET/PATCH /me`, `POST /me/profile-photo`, `GET /me/wallet` | Fetch `GET /me` after app boot to hydrate account state. |
| Template catalog | `GET /templates` | Respect the `locked`, `supports_photo`, and `accent_color_hex` fields returned for the current plan. |
| CV editor | `POST/GET /cvs`, `GET/PATCH/DELETE /cvs/:id`, `POST /cvs/:id/duplicate` | Persist a `change_summary` with meaningful editor actions. |
| Version history | `GET /cvs/:id/versions`, `POST /cvs/:id/versions/:versionID/restore` | Treat restore as a destructive replacement of current content; ask the user to confirm. |
| Guidance | `POST /cvs/:id/agent/ats-score`, `/agent/autofill-section` | Render ATS reports as advice and require user review before persisting AI-authored text. |
| Exports | `POST /cvs/:id/export` | Request `pdf`, `docx`, or `csv`; the API permits at most five unique export artifacts per user per Africa/Dar_es_Salaam calendar day and responds `429 DAILY_EXPORT_LIMIT` after that. |
| Billing | `GET /billing/plans`, `POST /billing/checkout`, `GET /billing/sessions/:ref/status` | Generate one idempotency key per intentional checkout action. |
| Provider webhooks | `POST /webhooks/mailafrica` | Server-to-server only; never call these from the frontend. |

## CV editor workflow

1. Request `GET /templates` and let the user choose an available template.
2. Create a CV with `POST /cvs`, placing editor content inside the `content` object.
3. Save incremental changes with `PATCH /cvs/:id` and a concise `change_summary`.
4. Call `POST /cvs/:id/agent/ats-score` before export to show structured lint advice.
5. Upload a user profile photo once using `POST /me/profile-photo`, then call `POST /cvs/:id/use-profile-photo` when the selected template supports a photo.
6. Create a PDF, DOCX, or spreadsheet-ready CSV through `POST /cvs/:id/export`; at most five new artifacts can be exported per Tanzania-local day.

Exports are written to the configured object store and downloaded through an authenticated owner-scoped API route. PostgreSQL records the artifact metadata and atomically enforces five new exports per user per Africa/Dar_es_Salaam day; byte-identical repeat exports reuse the stored artifact. Profile photos use the same object store in production.

## CORS, rate limits, and errors

The API accepts `GET`, `POST`, `PATCH`, `DELETE`, and `OPTIONS` requests. The allowlist derives from `ALLOWED_ORIGINS`; permitted request headers include `Authorization`, `Content-Type`, `Idempotency-Key`, and `X-Request-ID`. Production uses Redis-backed distributed limits (120 requests per minute globally per IP; admin login uses a stricter shared limit). Development falls back to process memory. [1]

| HTTP status | Typical code | Frontend behavior |
|---:|---|---|
| `400` | `VALIDATION_ERROR` | Show field-level or request-level input guidance. |
| `401` | `UNAUTHORIZED` | Clear stale session state and redirect to authentication. |
| `402` | `INSUFFICIENT_CREDITS` | Explain the export-credit requirement and offer the billing flow. |
| `404` | `NOT_FOUND` | Route to a recoverable missing-resource state. |
| `409` | `VALIDATION_ERROR` | Explain duplicate email or phone conflicts during signup. |
| `500` | `RENDER_FAILED` or `INTERNAL_ERROR` | Preserve the request ID and allow a safe retry where appropriate. |

## Webhooks and payments

MailAfrica webhooks operate on raw request bodies and are verified server-side. Do not proxy them through the frontend or alter the provider payload before it reaches the verification handler.

## Implementation references

[1] [`router.go`](./router.go) — Route map, CORS, rate limiter, docs endpoints, and webhook routes.  
[2] [`../config/config.go`](../config/config.go) — Base URL and CORS environment defaults.  
[3] [`../apiresponse/response.go`](../apiresponse/response.go) — Stable success and error envelopes.  
[4] [`handlers/core.go`](./handlers/core.go) — Auth, profile photo, CV, lint, and export behavior.  
[5] [`../../docs/third_party_contracts.md`](../../docs/third_party_contracts.md) — Provider webhook and service contracts.
## Administration

Admin password sign-in uses `ADMIN_LOGIN_EMAIL` and `ADMIN_LOGIN_PASSWORD`. Set both in the API deployment secret store; passwords must contain at least 12 characters. `POST /auth/admin/login` is rate limited and issues a 7-day access token. The planned opaque refresh-token lifetime is 30 days; refresh-token persistence, revocation, and rotation must be enabled before exposing refresh tokens to clients. Regular account sign-in continues through CamelAccounts. Never commit the administrator password.
