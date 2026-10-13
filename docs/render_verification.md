# Render Verification Notes

## ATS-safe PDF inspection

The generated Amina Juma and Baraka Mponzi PDFs were visually inspected after Chromium rendering. Each output is a one-page, single-column document with a readable contact line, standard ATS section headings, reverse-chronological work experience, bullet lists, and plain-text links. The deep forest green divider and heading treatment is restrained and does not introduce tables, icons, sidebars, or multi-column ordering risks.

The title, summary, experience dates, education, skills, certifications, and languages all rendered within the A4 page boundary. This confirms that the semantic HTML renderer and headless Chromium PDF stage produce reviewable, searchable document output for the ATS template variants.

## Local-template and DOCX inspection

The Neema Selemani local-template PDF also rendered to one page with the alternate green header treatment and legible content hierarchy. Its fixture photo URL deliberately did not resolve to a real external image; Chromium retained the accessible decorative-image fallback text without affecting the reading order of the CV content.

The Amina Juma DOCX was opened and re-rendered through LibreOffice. The output preserved the title, section hierarchy, experience bullets, contact text, and all core body content. The DOCX archive contains native `word/document.xml`, `word/styles.xml`, and `word/numbering.xml` parts, confirming that it is a real editable Word document rather than a rasterized PDF conversion.

## Refined hierarchy review

The refined Amina output confirms that the increased section rhythm, post-header separation, experience-entry spacing, and heading rules remain balanced within one A4 page. During the Neema local-template review, the fictional profile portrait did not embed into the generated PDF and Chromium displayed the image fallback text instead. The renderer will therefore downscale local profile images before base64 embedding to keep the inlined document payload within Chromium's reliable navigation limits.

After normalizing the fictional portrait and passing it as a validated inline image URL, the final Neema local-template PDF rendered as a single A4 page. The portrait is now cropped to a compact circular treatment in the header, while the summary, experience, education, skills, certifications, and languages retain clear section separation and readable hierarchy.

## Eight-design verification set

All eight available template IDs are represented exactly once by the fictional PDF set and every output is a single readable A4 page with extractable text. The Executive ATS example confirms a professional leadership-oriented hierarchy without a profile image, while the Hospitality Local example confirms the local visual family remains readable in a single-column ATS-friendly document flow. The full set is generated from source-controlled fixtures and can be regenerated with `cmd/render-samples`.

## Differentiated-template correction

The initial template set did not propagate fixture template IDs through the JSON decoding command, so the output read as a single shared design. After correcting the mapping and rebuilding the renderer, visual inspection confirms material differences: ATS Modern uses a full-width blue identity band, constrained content column, summary panel, and timeline accent; Executive ATS uses a serif leadership profile, narrow expertise rail, and two-column information architecture. The other template systems now likewise use distinct composition rules rather than color-only overrides.
