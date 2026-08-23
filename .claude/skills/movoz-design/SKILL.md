---
name: movoz-design
description: Use this skill to generate well-branded interfaces and assets for Movoz, either for production or throwaway prototypes/mocks/etc. Contains essential design guidelines, colors, type, fonts, assets, and UI kit components for prototyping.
user-invocable: true
---

Read the README.md file within this skill, and explore the other available files.
If creating visual artifacts (slides, mocks, throwaway prototypes, etc), copy assets out and create static HTML files for the user to view. If working on production code, you can copy assets and read the rules here to become an expert in designing with this brand.
If the user invokes this skill without any other guidance, ask them what they want to build or design, ask some questions, and act as an expert designer who outputs HTML artifacts _or_ production code, depending on the need.

## Quick map
- `readme.md` — the design guide: voice, color, type, motion, iconography, full manifest.
- `styles.css` — single CSS entry point; `@import`s `tokens/*.css` (fonts, colors, typography, layout, base + the `.movoz-grid-bg` / `.movoz-label` / `.movoz-annotation` helpers).
- `components/core/` — React primitives (`Button`, `Pill`, `Card`, `Badge`, `Input`, `Avatar`, `Logo`, `Annotation`, `PlaceholderBox`, `Skeleton`, `Tabs`). Each has a `.prompt.md` with usage.
- `guidelines/` — foundation specimen cards (colors, type, spacing, brand).
- `ui_kits/movoz-blog/` — full interactive product recreation (home + article).

## The look in one line
Lo-fi annotated wireframe: warm cream paper, Shantell Sans marker headings + Space Grotesk body, felt-tip 1.5px hairline borders, pencil-gray placeholders, faint blueprint grid, and a single terracotta accent for `↳` annotations and primary CTAs. No gradients, no emoji, no glass.

## Using components in an HTML artifact
```html
<link rel="stylesheet" href="styles.css" />
<script src="_ds_bundle.js"></script>
<script type="text/babel">
  const { Button, Card, Pill } = window.MovozDesignSystem_c2bc1d; // confirm namespace via check_design_system
</script>
```
For static mocks without the bundle, copy the token CSS and rebuild markup using the same classes/tokens.

## Notes
- Fonts load from Google Fonts (Shantell Sans is a substitute for the reference marker face — swap `tokens/fonts.css` if a licensed font is provided).
- No bundled icon set; use Unicode arrows (`→ ← ↳`) and 2px-stroke monoline SVG only.
