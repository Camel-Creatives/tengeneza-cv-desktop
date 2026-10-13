# Tengeneza CV — Design Direction

## Three possible approaches

### Theme Name: Nairobi Blue Ledger
**Very Brief Intro:** A high-trust, editorial workspace inspired by precision publishing and modern financial dashboards. Bright paper surfaces, deep navy typography, and measured cobalt accents make CV work feel deliberate rather than administrative.
**Probability:** 0.07

### Theme Name: Kanga Signal
**Very Brief Intro:** A warm bilingual identity system pairing cobalt structure with subtle woven-pattern geometry and humane Swahili-forward microcopy. It is expressive but still suitable for serious professional documents.
**Probability:** 0.04

### Theme Name: Studio Terminal
**Very Brief Intro:** A sparse, operational layout with vertical navigation, fine rules, and a document-first editing canvas. It evokes a professional creative studio rather than a generic SaaS dashboard.
**Probability:** 0.09

## Chosen direction: Nairobi Blue Ledger

### Design Movement
**Swiss editorial systems** adapted for a contemporary East African professional-product workspace. The interface treats the CV as a living document—clear typographic hierarchy, disciplined columns, and confident blue annotation rather than decorative dashboard chrome.

### Core Principles
1. **Document-first confidence:** The CV preview and its professional outcome take visual precedence over application controls.
2. **Calm operational clarity:** Essential actions have strong hierarchy; secondary actions remain quiet but discoverable.
3. **Measured asymmetry:** A fixed rail, offset content blocks, and a purposeful preview pane avoid a generic central-card layout.
4. **Language with equal dignity:** English and Swahili labels share the same information hierarchy, with compact switching and no translated-text overflow.

### Color Philosophy
White and faint mineral-blue surfaces keep the interface light enough for long-form writing. **Tengeneza Cobalt (#155EEF)** is reserved for decisions, progress, and direction; **Ledger Navy (#102A56)** carries authority in headings and navigation; a restrained cyan tint gives selected and active states a sense of air rather than visual noise. The resulting system communicates precision, trust, and forward movement.

### Layout Paradigm
A **working desk layout**: a narrow fixed navigation rail establishes product context, an editorial content column supports the current task, and a movable-looking document panel gives the CV permanent visual presence on wider screens. On compact screens, the rail becomes a temporary sheet and the document context moves below the active controls.

### Signature Elements
1. A cobalt **career trajectory line**—a precise horizontal rule with a circular locator—marks active workflow states and section progress.
2. **Ledger labels** use condensed uppercase type with generous tracking to organize utility information.
3. A miniature **document folio** treatment appears in CV previews: crop marks, page numbering, and a soft blue shadow—not a fake paper texture.

### Interaction Philosophy
Interactions should feel like focused desk work. Buttons depress crisply; item selection slides a blue locator into position; save and scoring states give calm, explicit feedback. Destructive and paid actions are always clearly named and confirmed before commitment.

### Animation
Use transform and opacity only. The rail and panels should enter with a 180–240ms `cubic-bezier(0.23, 1, 0.32, 1)` transition. Section cards can stagger at 45ms intervals. Button press feedback scales to 0.97 in 140ms. No looping decorative movement; all nonessential motion respects `prefers-reduced-motion`.

### Typography System
**Manrope** is the body and utility face for its clear, compact rendering across English and Swahili. **DM Serif Display** provides selective editorial authority for key product moments and document headings. Body text remains 14–16px; ledger labels are 10–11px uppercase with expanded letterspacing; section titles use Manrope 700; only major page titles use DM Serif Display.

### Brand Essence
**Tengeneza CV is the disciplined bilingual CV workspace for East African professionals who want their experience to travel further.**

Personality adjectives: **precise, encouraging, credible**.

### Brand Voice
Headlines speak with direct professional confidence; CTAs describe the outcome; helper text anticipates the next career decision without sounding generic.

Example lines:

> "Build the document your next opportunity deserves."

> "Hifadhi mabadiliko yako—then export when the score is ready."

### Wordmark & Logo
The mark is an abstract **N-path**: three offset cobalt strokes create a forward-moving route that also resembles a folded document corner. The wordmark uses a custom-spaced Manrope 800 with a deliberate gap after `TENGENEZA`, never a default unmodified brand name.

### Signature Brand Color
**Tengeneza Cobalt — #155EEF.**

## Style Decisions

- The product will remain light, blue-and-white, document-led, and professional; dark or neon treatments are excluded.
- CV data displayed in UI states must come from the API or be visibly presented as a no-data/connection state—no fabricated user records or testimonial-style content.
- Every route will retain a clear navigation route back to the workspace and a responsive mobile escape path.
- Every major route carries a distinct document-folio cue: crop marks, page numbering, a ruled document panel, a cobalt annotation, or the career-trajectory locator.
- The cobalt trajectory system uses **#155EEF** with a circular locator whenever the product communicates current position or progress.
- East African professional desk work and document preparation are the preferred imagery context; decorative imagery must be grounded by document annotations.
