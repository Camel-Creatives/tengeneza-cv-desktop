# Tengeneza CV API Documentation

This directory contains the maintained documentation set for the Tengeneza CV Platform API. The root [`README.md`](../README.md) is the primary onboarding document; this page explains which supporting document to use for a particular implementation or operations task.

## Documentation map

| Document | Primary audience | Questions it answers |
|---|---|---|
| [`../README.md`](../README.md) | Everyone | What the service does, how to run it, how to configure it, and how the major workflows fit together. |
| [`openapi.yaml`](./openapi.yaml) | Frontend, mobile, and SDK developers | Concise route overview; the full embedded contract is maintained in `../internal/httpapi/openapi.yaml`. |
| [`../internal/httpapi/README.md`](../internal/httpapi/README.md) | API integrators | How bearer auth, response envelopes, endpoint groups, profile photos, checkout retries, and signed webhooks behave. |
| [`../internal/render/README.md`](../internal/render/README.md) | Template and renderer contributors | How canonical CV JSON becomes PDF, DOCX, or CSV and how to validate template output. |
| [`../examples/README.md`](../examples/README.md) | Frontend and QA teams | How fictional fixtures, portrait assets, and output samples are structured. |
| [`operations.md`](./operations.md) | Operators | Deployment checklist, service boundaries, and launch controls. |
| [`template_library_delivery.md`](./template_library_delivery.md) | Product, design, and QA | 24-template inventory and current verification record. |
| [`render_verification.md`](./render_verification.md) | QA | Prior rendering checks and review artifacts. |

## Source-of-truth policy

The following hierarchy prevents drift between frontend assumptions and backend behavior:

1. The running API and its route implementation are the behavioral source of truth.
2. [`../internal/httpapi/openapi.yaml`](../internal/httpapi/openapi.yaml) is embedded and served by the API. [`openapi.yaml`](./openapi.yaml) is the concise repository overview; update both when a route changes.
3. The root README and subsystem READMEs explain how to use that contract.
4. Fixtures under [`../examples/sample-cvs`](../examples/sample-cvs) are representative fictional content, not production test data or user records.

When changing a route, request field, response field, status code, configuration variable, or template identifier, update the OpenAPI file and the affected README in the same commit. The route assembly itself is in [`../internal/httpapi/router.go`](../internal/httpapi/router.go). [1]

## Recommended reading paths

### Frontend integration

Start with the root README’s authentication and canonical CV-data sections. Then use `openapi.yaml` to generate or validate a client, and read the HTTP API README before implementing checkout, export polling, profile photos, or webhook-adjacent status displays.

### Template and export work

Read the rendering README, then inspect the fixtures and the template-library delivery manifest. Run the sample renderer after any HTML, DOCX, document-normalization, or fixture change. A visual review is required when a layout system changes.

### Production operations

Read the root README configuration section, `operations.md`, and `third_party_contracts.md` together. The router requires PostgreSQL, Redis, and S3-compatible storage in production, applies embedded migrations at startup, and exposes dependency-backed `/readyz`. Read the launch verification status in [`operations.md`](./operations.md) before treating live database and object-storage integration as verified.

## Documentation quality checklist

| Change type | Documentation that must be updated |
|---|---|
| New or changed endpoint | `openapi.yaml`, root README route table if relevant, `internal/httpapi/README.md` |
| New configuration variable | `.env.example`, root README, and operations guidance |
| New template | Template catalog, fixture, `template_library_delivery.md`, and rendering README if behavior changes |
| Changed export behavior | Root README, rendering README, HTTP API README, and OpenAPI contract |
| Provider-contract change | `third_party_contracts.md`, operations guide, webhook/API README sections |

## Implementation references

[1] [`../internal/httpapi/router.go`](../internal/httpapi/router.go) — Current route assembly, middleware, and service wiring.
