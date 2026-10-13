# Production Operations Guide

The public API base URL is configured with `PUBLIC_API_BASE_URL` (currently `https://api.tengenezacv.online`). It is separate from the upstream Ngamia AI URL (`NGAMIA_AI_BASE_URL`, currently `https://api.ngamia.cc/v1`).

## Runtime dependencies

| Dependency | Production behavior | Configuration |
|---|---|---|
| PostgreSQL | Durable users, CamelAccounts subject mapping, CV workspaces, CV versions, templates, exports, and daily export accounting. Embedded SQL migrations apply at API startup under a PostgreSQL advisory lock. | `DATABASE_URL` |
| Redis | Shared API rate-limit counters. | `REDIS_URL` |
| S3-compatible object storage | Export artifacts and profile photos. Export downloads are owner-authenticated through the API. | `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, optional `S3_SECURE=true` |
| Chromium | Synchronous PDF generation in the API process. The API Docker image installs Chromium and fonts. | `CHROMIUM_EXECUTABLE` |
| CamelAccounts | End-user identity and profile claims. The API maps the verified subject to a durable internal user ID and refreshes name/email/phone claims. | `CAMEL_ACCOUNTS_ISSUER`, `CAMEL_ACCOUNTS_AUDIENCE`, `CAMEL_ACCOUNTS_JWKS_URL` |

Production startup fails when required dependencies cannot be connected or required settings are absent. `/healthz` is a liveness endpoint. `/readyz` pings PostgreSQL, Redis, and object storage and returns `503` if a dependency is unavailable. Development may use in-memory state and local filesystem storage; those fallbacks are not suitable for multi-instance traffic.

## Exports and quota

`POST /cvs/:id/export` renders `pdf`, `docx`, or `csv` synchronously. PostgreSQL atomically reserves at most five new artifact exports per user per `Africa/Dar_es_Salaam` calendar day. A completed, matching artifact can be reused without consuming another slot. A failed render or failed object upload releases its reservation. The API returns an authenticated owner-scoped download route; keep the S3 bucket private.

PDF generation runs in the API request path. Account for Chromium CPU and memory use when sizing and autoscaling API replicas. The worker binary currently only validates configuration and exits after logging bootstrap; it does not consume jobs. There is no asynchronous render queue or signed-URL workflow in the current implementation.

## Deployment checklist

1. Provision PostgreSQL, Redis, a private S3-compatible bucket, CamelAccounts client settings, and provider credentials in the deployment secret manager.
2. Set the production `APP_ENV`, `PUBLIC_API_BASE_URL`, `ALLOWED_ORIGINS`, and all required credentials. Keep JWT signing keys at least 32 characters and never commit secrets.
3. Build and deploy the API using `Dockerfile`; the image includes Chromium for PDF rendering. Confirm `/healthz` and `/readyz` after startup and after dependency changes.
4. Confirm database migrations completed in API logs. Back up PostgreSQL and configure object lifecycle/retention policies appropriate to CV data and applicable privacy obligations.
5. Test CamelAccounts login, profile claim mapping, cross-user workspace isolation, profile-photo upload, and owner-only downloads with separate test accounts.
6. Verify PDF, DOCX, and CSV files open correctly from each active template family. Test five new exports, the sixth-export `429 DAILY_EXPORT_LIMIT`, cached reuse, and the reset at Tanzania midnight.
7. Register and test MailAfrica webhook signatures and verify the production sending domain before enabling transactional email.

## Verification status

The repository includes unit and contract tests plus an opt-in PostgreSQL integration test (`TEST_DATABASE_URL`). The API test suite passes locally. At the time this guide was updated, live database integration could not be confirmed because the running local PostgreSQL container rejected the credentials declared in this checkout's Compose configuration. The database volume was not reset. Treat database migration/persistence and S3 integration as requiring deployment-environment verification before opening real multi-instance traffic.

## Security controls and operational limits

The API uses short-lived JWT access tokens, CamelAccounts token verification, request IDs, security headers, restricted CORS, Redis-backed rate limits, private object storage, and owner-scoped export retrieval. Profile images are content-checked JPEG/PNG/WebP files up to 5 MB and stored under UUID-based object keys. Avoid logging CV content, tokens, OTPs, provider keys, or raw webhook payloads.

`/metrics` is currently a placeholder endpoint, not a production metrics exporter. Configure infrastructure-level metrics, logs, alerts, database backups, and restore drills before launch. Database migrations currently run during API startup; deploy with credentials able to apply schema changes and ensure the advisory lock serializes concurrent replica starts.
