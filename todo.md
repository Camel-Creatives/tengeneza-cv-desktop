# Tengeneza CV Platform API Delivery Checklist

## Sample CV examples

- [x] Create representative, fictional structured CV examples for three Tanzania- and diaspora-relevant career paths.
- [x] Verify each example against the ATS-safe lint rules.
- [x] Commit and push the sample CV collection to `main`.

## Document rendering worker

- [x] Implement semantic HTML templates and deterministic PDF generation for ATS-safe and local CV variants.
- [x] Implement native editable DOCX generation with real headings, paragraphs, and bullet lists.
- [x] Integrate render jobs with the export endpoint and local filesystem/S3-compatible storage interface.
- [x] Generate, extract text from, and visually inspect all three fictional sample PDF/DOCX outputs.
- [x] Commit and push the verified rendering-worker phase to `main`.

## Profile photo and template refinement

- [x] Add validated image-upload support and photo metadata for profile-capable local templates.
- [x] Improve typographic hierarchy and section spacing in PDF/DOCX renderers.
- [x] Expand the template catalog with distinct ATS-safe and local visual variants.
- [x] Generate clearly fictional portrait assets and render refreshed professional sample documents.
- [x] Inspect, verify, commit, and push the visual refinement phase to `main`.

## Eight-design PDF verification set

- [x] Create one clearly fictional structured CV fixture for each of the eight available designs.
- [x] Render all eight fixtures to real searchable PDF outputs.
- [x] Inspect representative outputs and confirm every design is represented exactly once.
- [x] Commit and push the eight-design PDF verification set to `main`.

## Template differentiation correction

- [x] Define eight visibly distinct layout, typography, color, and information-hierarchy systems.
- [x] Implement template-specific PDF and DOCX rendering rather than color-only overrides.
- [x] Regenerate and visually compare all eight variants for meaningful differentiation.
- [x] Commit and push the corrected document-design system to `main`.

## Sixteen-template professional library expansion

- [x] Define quality rules and compose one creative, technology, finance, and legal template at a time.
- [x] Define and compose one healthcare, education, sales, and operations template at a time.
- [x] Define and compose one consulting, engineering, public-sector, and nonprofit template at a time.
- [x] Define and compose one early-career, skilled-trades, remote, and creative-producer template at a time.
- [x] Add professional fictional profile-photo fixtures for all 16 new designs.
- [x] Render, inspect, and validate every new PDF individually for hierarchy, spacing, and legibility.
- [x] Commit and push each completed design phase and the final 24-template library to `main`.

## Repository documentation refresh

- [x] Audit the API implementation, environment configuration, contracts, and current documentation coverage.
- [x] Write a comprehensive root README covering setup, architecture, API usage, rendering, billing, security, and operations.
- [x] Add focused READMEs for the API surface, rendering system, and development environment.
- [x] Validate all documented commands and links, then commit and push the documentation refresh to `main`.

- [x] Review all remaining specification sections, dependencies, and production constraints.
- [x] Create the private GitHub repository and initialize a clean Go module on `main`.
- [x] Implement configuration, health, request tracing, structured errors, migrations, and repository scaffolding.
- [x] Implement authentication, users, CV documents, ATS linting, exports, and wallet accounting.
- [x] Implement AI gateway, MailAfrica, webhook verification, and asynchronous integration boundaries.
- [x] Author and serve a complete versioned OpenAPI specification with Swagger UI.
- [x] Add security, validation, idempotency, rate-limiting, test coverage, and operational documentation.
- [x] Verify every phase, commit it, and push it to the new repository’s `main` branch.
- [x] Run final integration checks and deliver the source repository and API documentation URLs.
