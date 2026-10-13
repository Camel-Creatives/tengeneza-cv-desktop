# Fictional CV Fixtures and Render Samples

This directory contains **fictional demonstration data only**. The fixtures exist so frontend, API, template, and QA contributors can exercise the full CV pipeline without touching personal customer data.

## Directory layout

| Path | Contents |
|---|---|
| [`sample-cvs/`](./sample-cvs) | One JSON CV fixture for each active template design. |
| [`profile-assets/library/`](./profile-assets/library) | Clearly fictional normalized profile-photo assets used by photo-capable fixtures. |
| [`../docs/template_library_delivery.md`](../docs/template_library_delivery.md) | Template inventory, sample names, and delivery verification record. |

## Fixture contract

Each fixture contains the top-level metadata used by `cmd/render-samples` plus a canonical `content` object:

```json
{
  "title": "Fictional Candidate — Example CV",
  "document_style": "resume",
  "template_id": "tech_product",
  "target_region": "eac",
  "language": "en",
  "content": {
    "personal": {},
    "summary": "",
    "experience": [],
    "education": [],
    "certifications": [],
    "skills": {},
    "settings": {}
  }
}
```

Use stable snake_case JSON keys. Experience bullets should begin with clear action verbs and include quantified outcomes whenever naturally possible; the fixture lint test enforces core ATS-quality checks. [1]

## Photo-capable samples

All 16 new industry-specific templates have fixture-level `show_photo: true` and a local `file://` portrait path. The `cmd/render-samples` command marks photo-capable template IDs as visual templates so the renderer can inline those local assets. The original and normalized portraits are fictional and must never be represented as real customers, staff, or public figures. [2]

## Generate samples locally

```bash
go run ./cmd/render-samples \
  -input examples/sample-cvs \
  -output /tmp/tengenezacv-rendered-samples
```

The command generates a PDF and a DOCX file for every fixture. The output filenames derive from `personal.full_name`, for example `Amina_Juma_CV.pdf` and `Amina_Juma_CV.docx`.

```bash
find /tmp/tengenezacv-rendered-samples -maxdepth 1 -type f | sort
```

## Adding or revising a fixture

1. Use a fictional person, employer, school, contact address, phone number, and achievement data.
2. Choose a template ID that exists in the active template catalog.
3. Set `target_region` intentionally because photo-capable templates may be discouraged for certain regions by ATS lint guidance.
4. Keep the CV complete enough to test summary, experience, education, skills, certification, and language rendering paths.
5. Add a fictional photo only for templates that support it and use the normalized `profile-assets/library` path.
6. Run the full lint test and render command; visually inspect the affected PDF before committing.

## Privacy and safety

Do not add real applicants, actual contact details, government identifiers, customer records, or copyrighted portrait materials. Use fictitious `.example.test` email domains and explicitly fictional organization labels where appropriate.

## Implementation references

[1] [`../internal/render/lint/sample_fixtures_test.go`](../internal/render/lint/sample_fixtures_test.go) — Fixture count and ATS-quality test.  
[2] [`../cmd/render-samples/main.go`](../cmd/render-samples/main.go) — Fixture loading and photo-capable template registration.
