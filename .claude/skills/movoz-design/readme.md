# Movoz Design System

A lo-fi, **annotated-wireframe** design language for Movoz — a developer-identity
company (OAuth 2.1 / OpenID Connect / token security). The aesthetic is "designed
to look like a thoughtful sketch": warm paper surfaces, a marker handwriting
voice, felt-tip hairline borders, pencil-gray placeholders, blueprint grid
backgrounds, and a single terracotta accent used for hand-written spec
annotations.

> The system was derived from a reference wireframe of the Movoz blog (header,
> hero, filter row, 3-column article grid, floating tool dock). The reference
> used the placeholder brand "Tomcat"; everything here is rebranded **Movoz**.

## Sources
- `uploads/Screenshot 2026-06-22 at 3.54.30 PM.png` — the reference wireframe
  (style + font cue). The user noted they "like the font and the style."
- No codebase or Figma was provided. Components are an original recreation of the
  visual language in the screenshot, not a port of existing code.

---

## Content fundamentals

How Movoz writes copy:

- **Two registers.** UI chrome and titles are in the **marker font** (warm,
  human, slightly playful). Body copy and metadata are in a **neutral grotesque**
  (Space Grotesk) for clarity. This split is the core of the voice.
- **Titles are plain-spoken and concrete**, often first-person-plural and
  outcome-led: *"How we built token rotation that survives 50M sessions"*,
  *"Refresh-token reuse detection"*, *"PKCE by default"*. No marketing
  superlatives, no clickbait.
- **"We" for the company, "you" for the reader.** Engineering posts narrate what
  *we* built; docs/CTAs address *you*.
- **Sentence case** for titles and buttons (*"Read article"*, *"Subscribe"*).
  **UPPERCASE + wide letter-spacing** only for small meta labels
  (`FEATURED · CATEGORY · READ TIME`).
- **Terse metadata, middot-separated:** `Architecture · 9 min`, `Jun 22, 2026 · 14 min read`.
- **Annotations are hand-written asides**, lowercase, prefixed with `↳`:
  *"↳ hero — featured / latest post"*, *"↳ article grid · 3 cols"*. They label
  structure, not marketing.
- **No emoji.** The "playfulness" comes entirely from the marker typeface and the
  sketch motif, never from emoji or exclamation marks.
- **Tone:** confident, technical, unhurried. Calm engineering blog, not a startup
  landing page.

Examples:
- Button: `Read article →` · `Subscribe` · `Annotate`
- Takeaway callout: *"Rotate on every exchange. Detect reuse. Revoke the family."*
- Topic pills: `All` · `OAuth 2.1` · `OpenID Connect` · `Token Security` · `Architecture` · `Ways of Working`

---

## Visual foundations

**Colors.** Warm and paper-like. Canvas is kraft cream `--paper #EFE9DC`; recessed
bands (hero/footer) drop to `--paper-sunken #E6DECE`; cards rise to
`--paper-raised #FAF5EA`. Text is warm graphite `--ink #25231E` with a 4-step
ink ramp. Placeholders/skeletons are pencil grays (`--pencil`, `--pencil-light`).
**One chromatic accent only:** terracotta `--terracotta #C25A36`, reserved for
annotations, primary CTAs, and toggle states — never as a fill across large
areas. Status colors exist but are muted and earthy.

**Type.** Display/headings/UI = **Shantell Sans** (marker handwriting), bold,
tight tracking, often italic for the wordmark and annotations. Body + labels =
**Space Grotesk**. Meta labels are 12px uppercase with `0.16em` tracking. Body
copy is frequently *represented by gray skeleton bars* in mockups rather than
rendered as lorem — that ragged-bar texture is part of the look.

**Backgrounds.** Flat warm paper. The signature texture is a faint **blueprint
grid** (`.movoz-grid-bg`, 24px, `--grid #DFD7C6`) behind content sections — never
photographic, never gradient. No gradients anywhere. No glassmorphism.

**Borders & corners.** Everything is drawn with a **felt-tip hairline**
(`--border-width 1.5px`) in `--line`/`--ink`. Corners are soft (`--radius-md 11px`
typical); a special `--radius-sketch` gives asymmetric "drawn by hand" edges. Pills
are fully rounded.

**Elevation.** Two shadow systems: (1) soft warm paper shadows (`--shadow-sm/md/lg`)
for subtle depth; (2) the **hard "sketch lift"** `--shadow-sketch 3px 3px 0 ink`
— a marker drop-shadow used on primary buttons and emphasized cards.

**Cards** = paper-raised + 1.5px hairline + md radius. Interactive cards nudge
up-left 2px and darken their border to ink on hover (revealing the sketch lift).

**Motion.** Restrained and quick (120–320ms, `--ease-out`). The one expressive
move is **press-to-depress**: lifted buttons translate `+3px,+3px` and drop their
shadow on `:active`, like pressing a stamp. No bounces on content, no infinite
loops, no parallax.

**Hover / press states.** Hover = darker border / slight raise (cards), subtle
fill (pills/tabs). Press = the depress-into-shadow move. Disabled = ~45% opacity.

**Imagery.** Real imagery is the exception; the default is the crossed
**PlaceholderBox** ("cover image", "code sample / token exchange"). When photos
are used they should read warm and muted to sit on paper. Diagrams are line art,
not filled illustration.

**Transparency / blur.** Essentially none — this is an opaque-paper system.

**Dark mode.** Set `data-theme="dark"` on `<html>` or `<body>`. Every semantic
token (`--paper`, `--ink`, `--terracotta`, `--line`, status colors, `--dock`)
flips via `tokens/colors-dark.css` — components read the same variables, so
nothing else changes. Surfaces become charcoal "night paper" (`--paper #21201B`),
ink becomes warm cream, and terracotta shifts lighter for contrast. Try it live
in the Movoz Blog kit's tool dock (**Dark** toggle) or the "Dark theme" swatch
card.

---

## Iconography

- **Approach: minimal, line-drawn, monoline.** Icons are sparse — the marker
  typography and the `↳` annotation hook carry most of the personality.
- **Unicode glyphs as icons** where natural: `→` (read more / next), `←` (back),
  `↳` (annotation hook), `·` (middot separators). These are intentional and
  on-brand.
- **Inline SVG** for the few functional icons (search), drawn at **2px stroke,
  round caps, no fill** to match the felt-tip line weight. Keep any added icons to
  that spec.
- **No icon font, no emoji, no filled/duotone icon sets.** If a broader icon set
  is ever needed, use a thin monoline set (e.g. Lucide at 2px) and recolor to
  `--ink` / `--ink-soft`. *(No icon library is bundled — substitution would need
  to be flagged.)*
- The **logo** is type-only: a boxed italic marker initial + italic wordmark
  (`Logo` component). No separate symbol/glyph mark.

Assets: this system ships no raster logo/illustration files — the logo is rendered
from the `Logo` component, placeholders from `PlaceholderBox`, and the grid from
CSS. `assets/` is reserved for any real imagery added later.

---

## Index / manifest

**Root**
- `styles.css` — single entry point; `@import`s everything below.
- `tokens/` — `fonts.css`, `colors.css`, `typography.css`, `layout.css`, `base.css`.
- `readme.md` — this guide.
- `SKILL.md` — Agent-Skills-compatible entry.

**Components** (`components/core/`, namespace `window.MovozDesignSystem_*`)
- `Button` — felt-tip outlined control (primary / secondary / accent / ghost; `lift`).
- `Pill` — rounded filter chip (`active`).
- `Card` — paper sheet surface (`lift`, `interactive`).
- `Badge` — uppercase status / category marker (muted tones).
- `Input` — outlined text field with optional leading icon (`pill`).
- `Avatar` — circular outlined author/user bubble.
- `Logo` — boxed marker initial + italic wordmark (`onDark`).
- `Annotation` — terracotta `↳` hand-written spec note.
- `PlaceholderBox` — crossed image/content placeholder frame.
- `Skeleton` — pencil-gray placeholder bars.
- `Tabs` — segmented control / the dark floating tool dock (`tone="dock"`).
- `core.card.html` — Design System tab specimen for all of the above.

**Foundation cards** (`guidelines/`) — Colors (Surfaces, Ink, Pencil & Lines,
Terracotta, Status), Type (Marker display, Body & labels, Annotation voice),
Spacing (Scale, Radius), Brand (Elevation, Logo lockup, Placeholders & paper).

**UI kits**
- `ui_kits/movoz-blog/` — the Movoz engineering blog: interactive Home ⇄ Article
  with Annotations / Grid overlays driven by a floating tool dock. See its
  `README.md`.

## How to consume
Link the one stylesheet, load the bundle, read components off the namespace:
```html
<link rel="stylesheet" href="styles.css" />
<script src="_ds_bundle.js"></script>
<script>const { Button, Card } = window.MovozDesignSystem_c2bc1d;</script>
```
