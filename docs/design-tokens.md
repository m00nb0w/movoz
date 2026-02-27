# Design Tokens Reference

Package: `@movoz/tokens` | Source: `packages/tokens/src/`

All tokens are exported as `as const` TypeScript objects with full type inference. Zero runtime dependencies.

## Colors

### Palette

| Token | Value | Usage |
|---|---|---|
| `accent.DEFAULT` | `#d4775c` | Primary accent (warm rust/orange) |
| `accent.light` | `#e08a70` | Lighter accent variant |
| `accent.dark` | `#c46448` | Darker accent variant |

### Semantic Colors

Theme-dependent colors used via CSS variables (`--zen-*`).

| Token | Light | Dark | CSS Variable |
|---|---|---|---|
| `bg` | `#fdf6e3` | `#1a1a1a` | `--zen-bg` |
| `text` | `#3d3520` | `#e8e4dc` | `--zen-text` |
| `muted` | `#857a5c` | `#a09080` | `--zen-muted` |
| `subtle` | `#f5edd8` | `#252525` | `--zen-subtle` |
| `border` | `#e8dfc5` | `#3a3a3a` | `--zen-border` |
| `paper` | `#fffaed` | `#222222` | `--zen-paper` |

Light mode is warm cream/parchment. Dark mode is Kindle-inspired.

## Typography

### Font Families

| Token | Stack | Usage |
|---|---|---|
| `sans` | Rubik, system-ui, sans-serif | Body text, headings |
| `serif` | Libre Baskerville, Georgia, serif | Display headings |
| `ui` | Inter, system-ui, sans-serif | UI elements, small text |

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
| `sm` | 0.25rem | 4 |
| `DEFAULT` / `md` | 0.5rem | 8 |
| `lg` | 0.75rem | 12 |
| `xl` | 1rem | 16 |
| `2xl` | 1.5rem | 24 |
| `full` | 9999px | pill |

## Shadows

| Token | Value |
|---|---|
| `none` | none |
| `sm` | `0 1px 2px rgba(0,0,0,0.05)` |
| `DEFAULT` | `0 1px 3px rgba(0,0,0,0.05), 0 4px 12px rgba(0,0,0,0.04)` |
| `md` | `0 4px 6px -1px rgba(0,0,0,0.07), 0 2px 4px -2px rgba(0,0,0,0.05)` |
| `lg` | `0 10px 15px -3px rgba(0,0,0,0.08), 0 4px 6px -4px rgba(0,0,0,0.04)` |
| `xl` | `0 20px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.04)` |
| `2xl` | `0 20px 40px -12px rgba(0,0,0,0.15)` |

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
