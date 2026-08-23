# Design Tokens Reference

Package: `@movoz/tokens` | Source: `packages/tokens/src/`

All tokens are exported as `as const` TypeScript objects with full type inference. Zero runtime dependencies.

## Colors

### Palette

| Token | Value | Usage |
|---|---|---|
| `accent.DEFAULT` | `#C25A36` | Primary accent (terracotta) |
| `accent.light` | `#E07A50` | Lighter accent variant |
| `accent.dark` | `#A1462A` | Darker accent variant |
| `accent.soft` | `#EFDBD0` | Tinted fill for subtle accent surfaces (e.g. `Badge` subtle variant) — documents the light-mode value; the runtime source is `--terracotta-soft` in `globals.css`, which also has a dark-mode value (`#4A3226`) |

### Semantic Colors

Theme-dependent colors used via CSS variables (`--paper`, `--ink`, `--terracotta`, etc. — see `packages/theme/src/globals.css`, the runtime source of truth).

| Token | Light | Dark | CSS Variable |
|---|---|---|---|
| `paper` | `#EFE9DC` | `#21201B` | `--paper` |
| `paperRaised` | `#FAF5EA` | `#2B2A24` | `--paper-raised` |
| `paperSunken` | `#E6DECE` | `#1A1916` | `--paper-sunken` |
| `ink` | `#25231E` | `#F2EEE4` | `--ink` |
| `inkSoft` | `#6B675F` | `#AFA99B` | `--ink-soft` |
| `line` | `#D4D0C6` | `#3B382F` | `--line` |
| `pencil` | `#C7C3B9` | `#4B473D` | `--pencil` |
| `pencilLight` | `#DEDAD0` | `#38362F` | `--pencil-light` |
| `dock` | `#2A2823` | `#F2EEE4` | `--dock` |
| `dockLine` | `#45413A` | `#D8D3C6` | `--dock-line` |

Light mode is a warm paper/parchment lo-fi palette. Dark mode inverts paper/ink while keeping the terracotta accent (brightened for contrast).

## Typography

### Font Families

| Token | Stack | Usage |
|---|---|---|
| `marker` | Shantell Sans, Comic Sans MS, cursive | Hand-drawn/marker accents (headings, callouts) |
| `body` | Space Grotesk, ui-sans-serif, system-ui, sans-serif | Body text, UI elements (also mapped to Tailwind's `sans` default) |

### Font Sizes

| Token | Value | px |
|---|---|---|
| `xs` | 0.75rem | 12 |
| `sm` | 0.875rem | 14 |
| `base` | 1rem | 16 |
| `lg` | 1.125rem | 18 |
| `xl` | 1.25rem | 20 |
| `2xl` | 1.5rem | 24 |
| `3xl` | 1.875rem | 30 |
| `4xl` | 2.25rem | 36 |
| `5xl` | 3rem | 48 |
| `6xl` | 3.75rem | 60 |
| `7xl` | 4.5rem | 72 |

### Font Weights

| Token | Value |
|---|---|
| `light` | 300 |
| `normal` | 400 |
| `medium` | 500 |
| `semibold` | 600 |
| `bold` | 700 |
| `extrabold` | 800 |

### Line Heights

| Token | Value |
|---|---|
| `none` | 1 |
| `tight` | 1.1 |
| `snug` | 1.3 |
| `normal` | 1.5 |
| `relaxed` | 1.6 |
| `loose` | 1.7 |

### Letter Spacings

| Token | Value |
|---|---|
| `tighter` | -0.02em |
| `tight` | -0.01em |
| `normal` | 0 |
| `wide` | 0.01em |

## Spacing

4px base scale, matching Tailwind conventions.

| Token | Value | px |
|---|---|---|
| `0` | 0 | 0 |
| `0.5` | 0.125rem | 2 |
| `1` | 0.25rem | 4 |
| `1.5` | 0.375rem | 6 |
| `2` | 0.5rem | 8 |
| `2.5` | 0.625rem | 10 |
| `3` | 0.75rem | 12 |
| `4` | 1rem | 16 |
| `5` | 1.25rem | 20 |
| `6` | 1.5rem | 24 |
| `8` | 2rem | 32 |
| `10` | 2.5rem | 40 |
| `12` | 3rem | 48 |
| `16` | 4rem | 64 |
| `20` | 5rem | 80 |
| `24` | 6rem | 96 |
| `32` | 8rem | 128 |

## Border Radii

| Token | Value | px |
|---|---|---|
| `none` | 0 | 0 |
| `sm` | 7px | 7 |
| `DEFAULT` / `md` | 11px | 11 |
| `lg` | 16px | 16 |
| `xl` | 22px | 22 |
| `2xl` | 30px | 30 |
| `full` | 9999px | pill |
| `sketch` | `14px 11px 13px 12px` | hand-drawn asymmetric corners |

## Shadows

| Token | Value |
|---|---|
| `none` | none |
| `sm` | `0 1px 2px rgba(37,35,30,0.05)` |
| `DEFAULT` | `0 1px 3px rgba(37,35,30,0.05), 0 4px 12px rgba(37,35,30,0.04)` |
| `md` | `0 4px 6px -1px rgba(37,35,30,0.07), 0 2px 4px -2px rgba(37,35,30,0.05)` |
| `lg` | `0 10px 15px -3px rgba(37,35,30,0.08), 0 4px 6px -4px rgba(37,35,30,0.04)` |
| `xl` | `0 20px 25px -5px rgba(37,35,30,0.1), 0 8px 10px -6px rgba(37,35,30,0.04)` |
| `2xl` | `0 20px 40px -12px rgba(37,35,30,0.15)` |
| `sketch` | `3px 3px 0 var(--ink)` — flat hand-drawn "sticker" shadow |
| `sketchAccent` | `3px 3px 0 var(--terracotta)` — accent-colored sketch shadow |

Dark mode variants (`darkShadows`): `DEFAULT` and `2xl` with higher opacity.

## Breakpoints

| Token | Value |
|---|---|
| `sm` | 640px |
| `md` | 768px |
| `lg` | 1024px |
| `xl` | 1280px |
| `2xl` | 1536px |

## Animation

### Durations

| Token | Value |
|---|---|
| `fast` | 150ms |
| `normal` | 200ms |
| `slow` | 300ms |
| `slower` | 600ms |

### Easings

| Token | Value |
|---|---|
| `ease` | ease |
| `easeIn` | ease-in |
| `easeOut` | ease-out |
| `easeInOut` | ease-in-out |
| `cubic` | cubic-bezier(0.4, 0, 0.2, 1) |

### Keyframe Animations

| Name | Description | Tailwind Class |
|---|---|---|
| `fadeIn` | Opacity 0 → 1 | `animate-fade-in` |
| `slideUp` | Translate Y +20px → 0, opacity 0 → 1 | `animate-slide-up` |
| `slideInRight` | Translate X -20px → 0, opacity 0 → 1 | `animate-slide-in-right` |

All animations: 0.6s ease-out forwards.
