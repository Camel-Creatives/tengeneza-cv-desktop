# Tengeneza CV UI

Tengeneza CV is a Tanzania-focused career workspace built with Next.js 16, React 19, and Tailwind CSS 4. The signed-in workspace uses CamelAccounts OAuth, creates and edits CVs through the Tengeneza CV API, shows the signed-in user’s name, email, and phone in new CV content, and supports template selection, profile-photo management, version history, ATS feedback, AI-assisted drafting, and PDF, DOCX, and CSV downloads.

## Features

- **Personal workspace:** CVs, versions, and exports are requested from the API with the signed-in user’s bearer token. The API is responsible for owner scoping and durable storage.
- **Account details:** Workspace and new-CV defaults use name, email, and phone returned by the API’s `/me` profile, populated from verified CamelAccounts claims.
- **Exports:** The editor requests PDF, DOCX, or CSV from the API and downloads the result through the authenticated API URL. The API allows five new artifacts per Tanzania-local calendar day; completed matching exports may be downloaded again without another slot.
- **English and Kiswahili:** The app navigation rail offers EN/SW controls. The selection persists in browser local storage and the profile’s preferred language is hydrated when profile data loads. Translation coverage is currently focused on navigation and shared shell labels; several page-specific labels and messages remain English.
- **Responsive layout:** Workspace, profile, templates, examples, and CV editing use the same responsive application shell.

## Configuration

Set these variables in the UI deployment environment:

| Variable | Purpose |
|---|---|
| `NEXT_PUBLIC_TENGENEZACV_API_BASE_URL` | Browser-visible API base URL. `VITE_TENGENEZACV_API_BASE_URL` remains supported for the Vite client path. |
| `CAMEL_ACCOUNTS_URL` | Server-side OAuth issuer/base URL used by the callback. |
| `CAMEL_ACCOUNTS_CLIENT_ID` | OAuth client ID used by the callback. |
| `NEXT_PUBLIC_CAMEL_ACCOUNTS_URL` | Browser OAuth authorization URL. |
| `NEXT_PUBLIC_CAMEL_ACCOUNTS_CLIENT_ID` | Browser OAuth client ID. |

The registered public client ID is `tengenezacv-web`, with redirect URI `https://app.tengenezacv.online/auth/callback`. If you use another client, set both UI client ID variables to it and set the API's `CAMEL_ACCOUNTS_AUDIENCE` to the same value. Email/password and social authentication are handled by Camel Accounts; the CV API only validates the resulting access token and syncs its verified identity claims.

Configure Camel Accounts `CORS_ALLOWED_ORIGINS` and the CV API `ALLOWED_ORIGINS` to include the deployed UI origin. The credential form uses a first-party browser POST to Camel Accounts, while bearer tokens go to the CV API. Do not put API signing keys, provider credentials, or S3 credentials in `NEXT_PUBLIC_*` variables.

## Local development

Install dependencies with the repository’s pinned pnpm version, then run:

```bash
pnpm dev
```

Useful checks:

```bash
pnpm check
pnpm build
```

The Next.js app routes are under `src/app`; shared page components, API client, authentication helpers, and language context are under `client/src`. `pnpm vite:dev` and `pnpm vite:build` are retained for the separate Vite/server entrypoints.

For end-to-end local use, start the API and its PostgreSQL, Redis, and MinIO services using the API repository’s [`README`](../API/README.md). The UI’s callback route and API must share the same configured CamelAccounts issuer/client registration and allowed redirect origin.

## Integration and verification notes

- The API is authoritative for workspace isolation, export quota enforcement, and export ownership. Hiding or disabling an export button in the UI is not a security boundary.
- The UI currently supports downloading exports through an authenticated API route. Local file paths are legacy development inputs and should not be relied upon in a production browser.
- Language switching is available from the signed-in side navigation. Do not describe the entire application as fully translated until page copy has been localized and reviewed in both English and Kiswahili.
- API-backed integration should be checked with separate CamelAccounts test users, including cross-user CV access and the five-export limit.
