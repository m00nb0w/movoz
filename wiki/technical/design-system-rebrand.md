# Design System Rebrand — "Movoz" lo-fi annotated-wireframe kit

**Status**: Design approved, not yet implemented.
**Source material**: a full design-kit package (tokens, React component specimens, guideline HTML pages, a sample blog UI kit, and an installable Claude Agent Skill) delivered as `Movoz Design System.zip`, extracted for reference during this design at `/private/tmp/claude-502/-Users-lto-repos-personal-movoz/106e165e-26c9-4e0b-bb04-0fa3045c5a74/scratchpad/movoz-design-system/` (a session-scoped scratch path — the canonical copy this doc refers to going forward is the one installed at `.claude/skills/movoz-design/`, see Section D).

## Why

The user provided a new brand kit — a "lo-fi annotated-wireframe" aesthetic (warm paper surfaces, Shantell Sans marker headings, Space Grotesk body, a single terracotta accent, felt-tip hairline borders, faint blueprint grid) — and asked to update the monorepo's design system to it. This doc records the scope and technical approach agreed after walking through the existing system and asking the user to resolve the real architectural forks (replace vs. coexist, which apps, token naming, which new components, skill install).

## Current state (for reference)

A 4-layer, token-driven system already exists and is being kept — only the values (and some names) change:

`@movoz/tokens` (raw values, `packages/tokens/src/*.ts`) → `@movoz/tailwind-config` (Tailwind preset, `packages/tailwind-config/index.ts`) → `@movoz/theme` (CSS vars + light/dark toggling, `packages/theme/src/globals.css` + `ThemeProvider.tsx`) → `@movoz/ui-web` (Tailwind-based React components, `packages/ui-web/src/`).

Consumers: `apps/personal-site` and `apps/oncarinho` both consume all four layers, including referencing `zen-*` Tailwind classes and CSS vars directly in their own JSX/CSS, not only through `@movoz/ui-web` components. `apps/drunken-dolphin` has an entirely separate, unrelated token system (`wiki/design/drunken-dolphin/styles.css`-derived) and never touches `@movoz/theme`.

`ThemeProvider.tsx` toggles a `.dark` class on `document.documentElement`; `packages/theme/src/globals.css` already scopes dark values under a `.dark { ... }` selector. This mechanism is unchanged by this rebrand.

## Decisions

1. **Replace wholesale**, not a coexisting second theme. The new palette becomes the one design system for the shared packages going forward.
2. **Scope: shared packages only.** `@movoz/tokens`, `@movoz/tailwind-config`, `@movoz/theme`, `@movoz/ui-web` get fully rebranded. Page-level visual QA/adjustment for `apps/personal-site` and `apps/oncarinho` (making sure each page's layout/composition still looks intentional under the new brand, not just "doesn't 404") is explicitly **out of scope** — separate follow-up work, one app at a time. `apps/drunken-dolphin` is untouched entirely.
3. **Token names are renamed** to match the new semantics (not just re-valued), e.g. `--zen-bg` → `--paper`, `--zen-text` → `--ink`. This requires a **mechanical** find/replace of class/var names at every call site in `personal-site` and `oncarinho` — renaming only, no new visual decisions, no layout changes. This is the one place where "shared packages only" (#2) necessarily reaches into app code, because a shared package can't rename its own public vocabulary without updating whoever calls it.
4. **Four new components** get added to `@movoz/ui-web`, restyled to the new brand: `Pill`, `Skeleton`, `Tabs`, `PlaceholderBox`. Two are deliberately deferred: `Annotation` (terracotta "↳" hand-written note — tied to the specific wireframe/blog narrative) and `Logo` (each app likely wants its own literal wordmark, not a shared generic one).
5. **Install the kit as a Claude skill** at `.claude/skills/movoz-design/` for future on-brand mockup/prototype work, independent of and in addition to the package rebrand.

## Section A — Token renaming

### Colors

| Old CSS var | New CSS var | Old Tailwind class | New Tailwind class |
|---|---|---|---|
| `--zen-bg` | `--paper` | `bg-zen-bg` | `bg-paper` |
| `--zen-paper` | `--paper-raised` | `bg-paper` | `bg-paper-raised` |
| `--zen-subtle` | `--paper-sunken` | `bg-zen-subtle` | `bg-paper-sunken` |
| `--zen-text` | `--ink` | `text-zen-text` | `text-ink` |
| `--zen-muted` | `--ink-soft` | `text-zen-muted` | `text-ink-soft` |
| `--zen-border` | `--line` | `border-zen-border` | `border-line` |

`accent` is unchanged as a name (already generic, not "zen"-prefixed) — only its value changes:
- `accent.DEFAULT`: `#d4775c` → `#C25A36` (terracotta)
- `accent.dark`: `#c46448` → `#A1462A` (terracotta-ink; matches the kit's `--accent-hover`)
- `accent.light`: `#e08a70` → `#E07A50` — **note**: the kit has no explicit light-mode "lighter terracotta for hover" swatch (its `terracotta-soft`/`terracotta-faint` are pale backgrounds, not a saturated light variant). This value is the kit's own *dark-mode* terracotta, repurposed here as the light-mode "light" variant since it's the closest on-brand lighter/more-vivid tone available. Flagging this as an inference, not a value taken verbatim from the source.

Light-mode values sourced directly from the kit: `--paper #EFE9DC`, `--paper-sunken #E6DECE`, `--paper-raised #FAF5EA`, `--ink #25231E`, `--ink-soft #6B675F`, `--line #D4D0C6`, `--terracotta #C25A36`.

### Fonts

| Old token | Old value | New token | New value |
|---|---|---|---|
| `serif` (display/headings) | Libre Baskerville | `marker` | Shantell Sans |
| `sans` (body) | Rubik | `body` | Space Grotesk |
| `ui` (small UI text) | Inter | *(merged into `body`)* | Space Grotesk |

Renaming `sans`/`ui` down to one `body` token matches the kit's actual "two registers" voice (marker for chrome/titles, one neutral grotesque for everything else) — keeping 3 font tokens when the brand only has 2 registers would be a false distinction. This is the largest mechanical-rename footprint of the three (headings/fonts appear on nearly every page): `font-serif` → `font-marker`, `font-sans`/`font-ui` → `font-body`, and `Text`'s `font` prop values (`"sans" | "serif" | "ui"` → `"marker" | "body"`).

### Radii & shadows

No renames — these token names were already generic scale names, not brand-specific. Values update in place:
- Radii: `sm` 4px→7px, `DEFAULT`/`md` 8px→11px, `lg` 12px→16px, `xl` 16px→22px (`full`/pill unchanged at 9999px). The kit doesn't define a `2xl` step (its scale stops at `xl`/22px before jumping to pill); `2xl` (currently 24px) is extrapolated to 30px, continuing the kit's own ~1.4x growth pattern — flagged as an inference, same as `accent.light` above.
- Shadows: `sm`/`DEFAULT`/`md`/`lg`/`xl`/`2xl` get re-tinted from neutral black rgba to ink-tinted rgba (`rgba(37, 35, 30, …)`), matching the kit's warm-paper shadow system.
- Two new tokens added: `shadow-sketch` (`3px 3px 0 var(--ink)`, the hard "stamp" shadow for primary buttons/emphasized cards) and `radius-sketch` (`14px 11px 13px 12px`, an asymmetric hand-drawn corner, used selectively, not a blanket replacement for `md`).

## Section B — Components

**Restyled in place, no prop/API changes** (only internal Tailwind classes and visual treatment change): `Button`, `Text`, `Input`, `Card`, `Badge`, `Avatar`, `IconButton`. They pick up felt-tip hairline borders (`border-line`, 1.5px), the re-tinted soft shadows, the sketch-lift press effect on `Button` and interactive `Card` (translate `+3px,+3px` and drop `shadow-sketch` on `:active`, per the kit's motion spec), and the `marker`/`body` fonts through the existing `font` prop plumbing.

**New primitives**, added and styled to match: `Pill` (rounded filter/toggle chip), `Skeleton` (pencil-gray loading bars), `Tabs` (segmented control, including a dark "dock" tone variant), `PlaceholderBox` (crossed placeholder frame for missing images/content). Built to the same conventions as existing primitives (`wiki/technical/ui-web-components.md`): `"use client"`, `className` passthrough via `cn()`, `forwardRef` where a DOM ref is meaningful.

**Deferred**: `Annotation`, `Logo` (see Decisions #4).

**Unaffected beyond automatic token inheritance**: `Stack`, `Container`, `Divider`, `Spacer` (layout), `Modal`, `ToastProvider`, `Dropdown` (complex) — none hardcode brand-specific styling today, so they pick up new colors/fonts for free with no direct edits needed.

## Section C — Theming, dark mode, fonts

**Dark mode**: the kit's dark palette (currently authored as `[data-theme="dark"]`-scoped CSS in the source kit) gets transplanted into the existing `.dark { ... }` block in `packages/theme/src/globals.css` — same class-toggling mechanism `ThemeProvider.tsx` already drives, new values only (e.g. dark `--paper: #21201B`, `--ink: #F2EEE4`, `--terracotta: #E07A50`).

**Font loading**: today, `apps/personal-site` loads its fonts via its own Google Fonts `@import` in `src/app/globals.css`; `apps/oncarinho` loads none (silently falls back to system fonts). This rebrand adds the Shantell Sans + Space Grotesk Google Fonts `@import` to `packages/theme/src/globals.css` itself, which every app already imports — consolidating font loading into the shared package apps get for free, rather than each app re-declaring it. `personal-site`'s now-unused Rubik/Inter/Libre-Baskerville `@import` line is left in place (removing it is personal-site-specific cleanup, out of scope; leaving it is harmless, just a redundant ignorable network request).

## Section D — Migration, verification, skill install

**Mechanical call-site rename**: apply the Section A rename tables verbatim across `apps/personal-site` and `apps/oncarinho` (oncarinho especially, since its whole frontend was built directly against `zen-*`/`font-serif`/`font-sans` classes per `tasks/oncarinho-frontend-plan.md`'s Global Constraints). No layout or composition changes — this is a rename, not a redesign. Sequencing note for the implementation plan: the old `bg-paper` class must be renamed to `bg-paper-raised` *before* `zen-bg` is renamed to `paper`, to avoid the two renames colliding on the string `paper`.

**Skill install**: copy the kit's contents (`SKILL.md`, `readme.md`, `styles.css`, `tokens/`, `components/`, `guidelines/`, `ui_kits/`, `_ds_bundle.js`, `_ds_manifest.json`) into `.claude/skills/movoz-design/` as-is.

**Verification** (matches this repo's existing bar — no unit test framework for these packages/apps today):
- `pnpm --filter @movoz/tokens build`, `pnpm --filter @movoz/theme lint` (tsc), `pnpm --filter @movoz/ui-web build` clean
- `pnpm --filter personal-site build` and `pnpm --filter oncarinho build` clean
- Belt-and-suspenders grep across both apps for any leftover `zen-`/old font-class references after the rename (a missed rename is silently wrong CSS, not a build failure)
- Manual visual check: `pnpm dev`, load personal-site and oncarinho in light + dark mode, confirm the new fonts/colors/borders render — accepting whatever the pages look like as a result, since page-level polish is explicitly deferred (Decisions #2)

## Out of scope (explicitly deferred, not forgotten)

- Page-level visual QA/adjustment for `personal-site` and `oncarinho` (Decisions #2) — separate follow-up, one app at a time.
- `apps/drunken-dolphin` — untouched, unrelated design system.
- `Annotation` and `Logo` components (Decisions #4).
- `personal-site`'s bespoke non-token CSS (`.glass`, `.paper-card`, `.card-hover`, hardcoded hex values like `#d4775c` in `.gradient-text`, its own now-redundant font `@import`) — these live outside the shared packages and are exactly the kind of thing the deferred page-level QA pass would address.
- oncarinho's previously-identified "What's Left" backlog (deployment, loading states, etc., from `tasks/oncarinho-frontend-plan.md`) — unrelated to this rebrand, still pending from before.
