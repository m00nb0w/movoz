# Movoz Lo-fi Design System Rebrand Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebrand the shared `@movoz/tokens` → `@movoz/tailwind-config` → `@movoz/theme` → `@movoz/ui-web` package chain from the old "zen" warm-cream palette to the new "Movoz" lo-fi annotated-wireframe kit (warm paper surfaces, Shantell Sans marker headings, Space Grotesk body, terracotta accent, felt-tip hairline borders), mechanically propagate the rename into `apps/personal-site` and `apps/oncarinho`, add four new brand components to `@movoz/ui-web`, and install the kit as a Claude skill.

**Architecture:** No new layers — the existing 4-package pipeline (`tokens` → `tailwind-config` → `theme` → `ui-web`) is rebranded in place, package by package, downstream-first (tokens must rebuild before tailwind-config/ui-web pick up new values). App-level work is a mechanical find/replace against a fixed rename table, not a redesign.

**Tech Stack:** TypeScript, Tailwind CSS, Next.js (App Router), pnpm workspaces, tsup (for `@movoz/tokens` builds).

**Spec:** `wiki/technical/design-system-rebrand.md`

**Kit source** (for verifying exact values against the original if anything looks off): `/private/tmp/claude-502/-Users-lto-repos-personal-movoz/106e165e-26c9-4e0b-bb04-0fa3045c5a74/scratchpad/movoz-design-system/` — this is a session-scoped scratch path and may not exist in a fresh session; the zip is also archived at `/Users/lto/Downloads/Movoz Design System.zip` and, after Task 18, permanently at `.claude/skills/movoz-design/`.

## Global Constraints

- Replace the "zen" system wholesale in the shared packages — no coexisting second theme (spec Decisions #1).
- Scope is the shared packages plus the *mechanical* call-site rename in `apps/personal-site`/`apps/oncarinho` — no page-level redesign, no layout changes, no touching `apps/drunken-dolphin` (spec Decisions #2).
- Token names are renamed to match new semantics, not just re-valued (spec Decisions #3). Full rename table:
  - `--zen-bg` / `bg-zen-bg` → `--paper` / `bg-paper`
  - `--zen-paper` / `bg-paper` (old) → `--paper-raised` / `bg-paper-raised`
  - `--zen-subtle` / `bg-zen-subtle` → `--paper-sunken` / `bg-paper-sunken`
  - `--zen-text` / `text-zen-text` → `--ink` / `text-ink`
  - `--zen-muted` / `text-zen-muted` → `--ink-soft` / `text-ink-soft`
  - `--zen-border` / `border-zen-border` → `--line` / `border-line`
  - `font-serif` / `font="serif"` → `font-marker` / `font="marker"`
  - `font-sans`, `font-ui` / `font="sans"`, `font="ui"` → `font-body` / `font="body"`
  - `accent` keeps its name; only its value changes (terracotta).
- Four new components ship in `@movoz/ui-web`: `Pill`, `Skeleton`, `Tabs`, `PlaceholderBox`. `Annotation` and `Logo` are explicitly deferred (spec Decisions #4).
- The kit installs as a Claude skill at `.claude/skills/movoz-design/` (spec Decisions #5).
- No unit test framework exists for these packages/apps (matches the rest of the repo) — verification is `build`/`lint` + a grep sweep for leftover old class names + a manual visual check, not automated tests.
- `@movoz/tokens` publishes compiled `dist/` output (`tsup`) — it must be rebuilt (`pnpm --filter @movoz/tokens build`) after every edit to its `src/`, or downstream packages keep reading stale values.
- `@movoz/theme` and `@movoz/ui-web` are consumed as TypeScript/CSS source directly (no build step for theme; `ui-web`'s `lint` script is `tsc --noEmit`, its `build` is `tsup`) — both apps `transpilePackages` them, so edits are picked up on the next `next build`/`dev` without a separate publish step.
- `@movoz/theme`'s `package.json` has no `lint` script — do not run `pnpm --filter @movoz/theme lint`, it will fail with "Missing script". Verify theme changes via the app builds instead.
- `packages/tokens/src/radii.ts` and `shadows.ts` are **not** wired into `@movoz/tailwind-config`'s Tailwind theme (no `borderRadius`/`boxShadow` extension exists — this is a pre-existing gap, not something this plan introduces or is expected to fix). Components historically hardcode literal arbitrary-value classes (e.g. `shadow-[0_1px_3px_rgba(...)]`) instead of a `shadow-md` token class. Task 1 updates these two files for documentation parity with the design-tokens wiki page, but the values that actually render come from the CSS custom properties added directly to `packages/theme/src/globals.css` in Task 3 (`--radius-sm/md/lg`, `--shadow-sketch`) and referenced via Tailwind arbitrary-value syntax (`rounded-[var(--radius-md)]`, `shadow-[var(--shadow-sketch)]`) in Tasks 4–15. Don't expect Task 1 alone to change anything visible.

---

### Task 1: Rebrand `@movoz/tokens` — colors, fonts, radii, shadows

**Files:**
- Modify: `packages/tokens/src/colors.ts`
- Modify: `packages/tokens/src/typography.ts`
- Modify: `packages/tokens/src/radii.ts`
- Modify: `packages/tokens/src/shadows.ts`

**Interfaces:**
- Consumes: nothing (leaf package)
- Produces: `palette.accent.{DEFAULT,light,dark}`, `semantic.{light,dark}.{paper,paperRaised,paperSunken,ink,inkSoft,line}`, `fontFamilies.{marker,body}` (replaces the old `sans`/`serif`/`ui` keys), `radii.{none,sm,DEFAULT,md,lg,xl,2xl,full,sketch}`, `shadows.{none,sm,DEFAULT,md,lg,xl,2xl,sketch,sketchAccent}`, `darkShadows.{DEFAULT,2xl}` (unchanged) — every later task in this plan imports from this package's built output.

- [ ] **Step 1: Replace `packages/tokens/src/colors.ts`**

```ts
export const palette = {
  accent: {
    DEFAULT: "#C25A36",
    light: "#E07A50",
    dark: "#A1462A",
  },
} as const;

export const semantic = {
  light: {
    paper: "#EFE9DC",
    paperRaised: "#FAF5EA",
    paperSunken: "#E6DECE",
    ink: "#25231E",
    inkSoft: "#6B675F",
    line: "#D4D0C6",
  },
  dark: {
    paper: "#21201B",
    paperRaised: "#2B2A24",
    paperSunken: "#1A1916",
    ink: "#F2EEE4",
    inkSoft: "#AFA99B",
    line: "#3B382F",
  },
} as const;

export type SemanticColorKey = keyof (typeof semantic)["light"];
export type ColorMode = keyof typeof semantic;
```

- [ ] **Step 2: Replace `fontFamilies` in `packages/tokens/src/typography.ts`**

Replace only the `fontFamilies` export (leave `fontSizes`, `fontWeights`, `lineHeights`, `letterSpacings` untouched):

```ts
export const fontFamilies = {
  marker: ["Shantell Sans", "Comic Sans MS", "cursive"],
  body: ["Space Grotesk", "ui-sans-serif", "system-ui", "sans-serif"],
} as const;
```

- [ ] **Step 3: Replace `packages/tokens/src/radii.ts`**

```ts
export const radii = {
  none: "0",
  sm: "7px",
  DEFAULT: "11px",
  md: "11px",
  lg: "16px",
  xl: "22px",
  "2xl": "30px",
  full: "9999px",
  sketch: "14px 11px 13px 12px",
} as const;
```

Note: the kit's own scale stops at 22px before jumping to the 999px pill; `2xl` (30px) is extrapolated continuing the kit's ~1.4x growth pattern, since no kit value exists for it (documented in the spec's Section A).

- [ ] **Step 4: Replace `packages/tokens/src/shadows.ts`**

```ts
export const shadows = {
  none: "none",
  sm: "0 1px 2px rgba(37, 35, 30, 0.05)",
  DEFAULT: "0 1px 3px rgba(37, 35, 30, 0.05), 0 4px 12px rgba(37, 35, 30, 0.04)",
  md: "0 4px 6px -1px rgba(37, 35, 30, 0.07), 0 2px 4px -2px rgba(37, 35, 30, 0.05)",
  lg: "0 10px 15px -3px rgba(37, 35, 30, 0.08), 0 4px 6px -4px rgba(37, 35, 30, 0.04)",
  xl: "0 20px 25px -5px rgba(37, 35, 30, 0.1), 0 8px 10px -6px rgba(37, 35, 30, 0.04)",
  "2xl": "0 20px 40px -12px rgba(37, 35, 30, 0.15)",
  sketch: "3px 3px 0 var(--ink)",
  sketchAccent: "3px 3px 0 var(--terracotta)",
} as const;

export const darkShadows = {
  DEFAULT: "0 1px 3px rgba(0, 0, 0, 0.2), 0 4px 12px rgba(0, 0, 0, 0.15)",
  "2xl": "0 20px 40px -12px rgba(0, 0, 0, 0.4)",
} as const;
```

Only the rgba color component changed (neutral black → ink-tinted `rgba(37, 35, 30, …)`) — all blur/spread/offset numbers are untouched. `darkShadows` is unchanged (dark-mode shadows correctly stay black-tinted against a dark background). Two new tokens added: `sketch` (hard stamp shadow) and `sketchAccent`.

- [ ] **Step 5: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/tokens build
```

Expected: clean build, no TypeScript errors. Confirm `packages/tokens/dist/index.js` was regenerated (check its mtime or `git diff --stat packages/tokens/dist/`).

- [ ] **Step 6: Commit**

```bash
git add packages/tokens/src packages/tokens/dist
git commit -m "design-system: rebrand @movoz/tokens colors, fonts, radii, and shadows"
```

---

### Task 2: Rebrand `@movoz/tailwind-config`

**Files:**
- Modify: `packages/tailwind-config/index.ts`

**Interfaces:**
- Consumes: `fontFamilies`, `keyframes`, `animations` from `@movoz/tokens` (Task 1's built output)
- Produces: Tailwind classes `font-marker`, `font-body`, `bg-accent`/`text-accent`/`border-accent` (+ `accent-light`/`accent-dark`), `bg-paper`, `bg-paper-raised`, `bg-paper-sunken`, `text-ink`, `text-ink-soft`, `border-line`, `bg-pencil`, `bg-pencil-light`, `bg-dock`, `border-dock-line` — every ui-web component task and every app-rename task depends on these classes existing.

- [ ] **Step 1: Replace `packages/tailwind-config/index.ts`**

```ts
import type { Config } from "tailwindcss";
import { fontFamilies, keyframes, animations } from "@movoz/tokens";

const config: Partial<Config> = {
  darkMode: "class",
  theme: {
    extend: {
      fontFamily: {
        marker: [...fontFamilies.marker],
        body: [...fontFamilies.body],
      },
      colors: {
        accent: {
          DEFAULT: "var(--terracotta)",
          light: "var(--terracotta-light)",
          dark: "var(--terracotta-ink)",
        },
        paper: "var(--paper)",
        "paper-raised": "var(--paper-raised)",
        "paper-sunken": "var(--paper-sunken)",
        ink: "var(--ink)",
        "ink-soft": "var(--ink-soft)",
        line: "var(--line)",
        pencil: "var(--pencil)",
        "pencil-light": "var(--pencil-light)",
        dock: "var(--dock)",
        "dock-line": "var(--dock-line)",
      },
      animation: { ...animations },
      keyframes: { ...keyframes },
    },
  },
};

export default config;
```

Note: `accent` switches from a static hex value (baked in at build time, never dark-mode-aware) to CSS-variable references, matching how `paper`/`ink` already worked — this is required because the kit's terracotta is genuinely a different shade in dark mode (`#C25A36` light vs `#E07A50` dark), which a static value can't express. `palette` is no longer imported here (it stays in `@movoz/tokens` purely as a documented reference, matching how `semantic` already worked before this change — never imported into Tailwind, kept in sync by hand).

`pencil`, `pencil-light`, `dock`, `dock-line` are additions beyond the Section A rename table — they have no old "zen" equivalent, they're net-new tokens needed by the `Skeleton`/`PlaceholderBox` (pencil) and `Tabs` dock tone (dock) components added in Tasks 12–15. Their CSS variables are defined in Task 3.

- [ ] **Step 2: Verify it compiles**

There's no standalone build for `@movoz/tailwind-config` (it's consumed as TS source via each app's `tailwind.config.ts`). Verification happens in Task 4 once `@movoz/ui-web` — the first consumer — is rebuilt. For now, just confirm there's no syntax error:

```bash
cd /Users/lto/repos/personal/movoz
npx tsc --noEmit packages/tailwind-config/index.ts --esModuleInterop --skipLibCheck
```

Expected: no errors (ignore any "cannot find module" for `tailwindcss`/`@movoz/tokens` if `skipLibCheck` doesn't fully resolve workspace packages in this ad hoc invocation — the real check is Task 4's `ui-web` build).

- [ ] **Step 3: Commit**

```bash
git add packages/tailwind-config/index.ts
git commit -m "design-system: rebrand @movoz/tailwind-config colors and font families"
```

---

### Task 3: Rebrand `@movoz/theme` — CSS variables, dark mode, font loading

**Files:**
- Modify: `packages/theme/src/globals.css`
- Modify: `packages/theme/src/ThemeToggle.tsx`

**Interfaces:**
- Consumes: nothing new (still just CSS custom properties + the existing `.dark` class toggled by `ThemeProvider.tsx`, which is unchanged)
- Produces: `--paper`, `--paper-raised`, `--paper-sunken`, `--ink`, `--ink-soft`, `--line`, `--pencil`, `--pencil-light`, `--terracotta`, `--terracotta-ink`, `--terracotta-light`, `--dock`, `--dock-line`, `--radius-sm`, `--radius-md`, `--radius-lg`, `--border-width`, `--shadow-sketch` CSS variables, both `:root` (light) and `.dark` (dark) — every ui-web component task that uses `rounded-[var(--radius-*)]`, `border-[length:var(--border-width)]`, or `shadow-[var(--shadow-sketch)]` arbitrary Tailwind values depends on these existing.

- [ ] **Step 1: Replace `packages/theme/src/globals.css`**

```css
/* Color values sourced from @movoz/tokens — keep in sync */
@import url('https://fonts.googleapis.com/css2?family=Shantell+Sans:ital,wght@0,300..800;1,300..700&family=Space+Grotesk:wght@400;500;600;700&display=swap');

:root {
  /* Warm paper/ink/terracotta palette */
  --paper: #EFE9DC;
  --paper-raised: #FAF5EA;
  --paper-sunken: #E6DECE;
  --ink: #25231E;
  --ink-soft: #6B675F;
  --line: #D4D0C6;
  --pencil: #C7C3B9;
  --pencil-light: #DEDAD0;
  --terracotta: #C25A36;
  --terracotta-ink: #A1462A;
  --terracotta-light: #E07A50;
  --dock: #2A2823;
  --dock-line: #45413A;

  --radius-sm: 7px;
  --radius-md: 11px;
  --radius-lg: 16px;
  --border-width: 1.5px;
  --shadow-sketch: 3px 3px 0 var(--ink);
}

.dark {
  --paper: #21201B;
  --paper-raised: #2B2A24;
  --paper-sunken: #1A1916;
  --ink: #F2EEE4;
  --ink-soft: #AFA99B;
  --line: #3B382F;
  --pencil: #4B473D;
  --pencil-light: #38362F;
  --terracotta: #E07A50;
  --terracotta-ink: #F2946C;
  --dock: #F2EEE4;
  --dock-line: #D8D3C6;
}
```

Note: `--terracotta-light` is deliberately *not* redefined inside `.dark` — it keeps inheriting the `:root` value (`#E07A50`), which happens to equal dark mode's own base `--terracotta`. This is an accepted, documented inference (spec Section A) since the kit has no explicit "light-mode-only accent, lighter still" swatch; leaving it undefined in `.dark` is the simplest safe choice, not a bug.

The Google Fonts `@import` moves here from being personal-site's own responsibility — every app that imports `@movoz/theme/globals.css` now gets the real fonts for free (spec Section C).

- [ ] **Step 2: Rename classes in `packages/theme/src/ThemeToggle.tsx`**

Read the file first, then replace every `zen-subtle`/`zen-text` occurrence per the rename table (this file wasn't explicitly called out in the spec's Section B component list, but it directly references the old class names, which stop existing after Task 2 — it must be updated or the theme toggle button silently loses its styling):

```bash
cd /Users/lto/repos/personal/movoz
perl -pi -e 's/zen-subtle/paper-sunken/g; s/zen-text/ink/g' packages/theme/src/ThemeToggle.tsx
```

- [ ] **Step 3: Verify the renames landed correctly**

```bash
grep -n "zen-" packages/theme/src/ThemeToggle.tsx
```

Expected: no output (no `zen-` references remain in this file).

- [ ] **Step 4: Commit**

```bash
git add packages/theme/src/globals.css packages/theme/src/ThemeToggle.tsx
git commit -m "design-system: rebrand @movoz/theme CSS variables, dark mode, and font loading"
```

---

### Task 4: Restyle `Button`

**Files:**
- Modify: `packages/ui-web/src/primitives/Button/Button.tsx`

**Interfaces:**
- Consumes: `cn` from `../../utils/cn` (unchanged), Tailwind classes from Task 2/3 (`font-marker`, `bg-ink`, `text-paper-raised`, `bg-paper-raised`, `bg-paper-sunken`, `border-line`... wait, Button doesn't use `border-line`, see below)
- Produces: `Button` component — **props/API unchanged** (`variant`, `size`, `loading`, `icon`, `iconRight` keep the same names, types, and defaults). Only internal classes change. Every consumer in `apps/personal-site` and `apps/oncarinho` keeps working with zero call-site changes to `<Button>` usages themselves (only the app-level class-name renames in Tasks 16–17 touch adjacent code, not `Button`'s own props).

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/Button/Button.tsx`**

```tsx
"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface ButtonProps extends ComponentPropsWithoutRef<"button"> {
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  loading?: boolean;
  icon?: ReactNode;
  iconRight?: ReactNode;
}

const variantStyles = {
  primary:
    "bg-ink text-paper-raised border border-ink hover:opacity-90 shadow-[var(--shadow-sketch)] active:translate-x-[3px] active:translate-y-[3px] active:shadow-none",
  secondary:
    "bg-paper-raised text-ink border border-ink hover:bg-paper-sunken",
  ghost:
    "bg-transparent text-ink border border-transparent hover:bg-paper-sunken",
  danger:
    "bg-red-500 text-white border border-red-500 hover:bg-red-600",
};

const sizeStyles = {
  sm: "px-3 py-1.5 text-sm rounded-[var(--radius-sm)] gap-1.5",
  md: "px-5 py-2.5 text-base rounded-[var(--radius-md)] gap-2",
  lg: "px-6 py-3.5 text-base rounded-[var(--radius-md)] gap-2",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = "primary",
      size = "md",
      loading = false,
      icon,
      iconRight,
      disabled,
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(
          "inline-flex items-center justify-center font-marker font-semibold transition-all duration-200",
          "disabled:opacity-45 disabled:pointer-events-none",
          variantStyles[variant],
          sizeStyles[size],
          className
        )}
        {...props}
      >
        {loading ? (
          <svg
            className="animate-spin h-4 w-4"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            />
          </svg>
        ) : (
          icon
        )}
        {children}
        {iconRight}
      </button>
    );
  }
);

Button.displayName = "Button";
```

Design note: the kit's `Button` has an opt-in `lift` boolean prop gating the sketch shadow. Adding a new prop would be an API change (ruled out by the spec), so instead the sketch-lift + press-in interaction (`active:translate-x-[3px] active:translate-y-[3px] active:shadow-none`, replacing the kit's imperative `onMouseDown`/`onMouseUp` handlers with pure CSS `:active`) is baked permanently into the `primary` variant only — consistent with the kit's own guidance that this treatment is "reserved for... primary CTAs." `secondary`/`ghost`/`danger` get no shadow.

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: both clean. This is the first real compile of the new Tailwind color/font names from Tasks 2–3 — if `font-marker`, `bg-ink`, etc. don't resolve, this is where it surfaces.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/Button/Button.tsx
git commit -m "design-system: restyle Button for the paper/ink/terracotta brand"
```

---

### Task 5: Restyle `Text`

**Files:**
- Modify: `packages/ui-web/src/primitives/Text/Text.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2 (`font-marker`, `font-body`, `text-ink`, `text-ink-soft`)
- Produces: `Text` component — **the `font` prop's allowed values change** from `"sans" | "serif" | "ui"` to `"marker" | "body"` (this is the one prop-surface change in this whole restyle pass, and it's required by the approved font-token rename in Global Constraints — every other prop is untouched). `apps/personal-site`'s `font="serif"` usages get updated to `font="marker"` in Task 16.

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/Text/Text.tsx`**

```tsx
"use client";

import { type ElementType, type ReactNode, createElement } from "react";
import { cn } from "../../utils/cn";

export interface TextProps<E extends ElementType = "p"> {
  as?: E;
  size?: "xs" | "sm" | "base" | "lg" | "xl" | "2xl" | "3xl" | "4xl" | "5xl";
  weight?: "light" | "normal" | "medium" | "semibold" | "bold";
  color?: "default" | "muted" | "accent";
  font?: "marker" | "body";
  truncate?: boolean;
  align?: "left" | "center" | "right";
  className?: string;
  children: ReactNode;
}

const sizeStyles = {
  xs: "text-xs",
  sm: "text-sm",
  base: "text-base",
  lg: "text-lg",
  xl: "text-xl",
  "2xl": "text-2xl",
  "3xl": "text-3xl",
  "4xl": "text-4xl",
  "5xl": "text-5xl",
};

const weightStyles = {
  light: "font-light",
  normal: "font-normal",
  medium: "font-medium",
  semibold: "font-semibold",
  bold: "font-bold",
};

const colorStyles = {
  default: "text-ink",
  muted: "text-ink-soft",
  accent: "text-accent",
};

const fontStyles = {
  marker: "font-marker",
  body: "font-body",
};

const alignStyles = {
  left: "text-left",
  center: "text-center",
  right: "text-right",
};

export function Text<E extends ElementType = "p">({
  as,
  size = "base",
  weight = "normal",
  color = "default",
  font,
  truncate = false,
  align,
  className,
  children,
  ...props
}: TextProps<E> & Omit<React.ComponentPropsWithoutRef<E>, keyof TextProps<E>>) {
  const Component = as || "p";
  return createElement(
    Component,
    {
      className: cn(
        sizeStyles[size],
        weightStyles[weight],
        colorStyles[color],
        font && fontStyles[font],
        align && alignStyles[align],
        truncate && "truncate",
        className
      ),
      ...props,
    },
    children
  );
}
```

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/Text/Text.tsx
git commit -m "design-system: restyle Text, rename font prop values to marker/body"
```

---

### Task 6: Restyle `Card`

**Files:**
- Modify: `packages/ui-web/src/primitives/Card/Card.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2/3 (`bg-paper-raised`, `bg-paper-sunken`, `border-line`, `rounded-[var(--radius-md)]`)
- Produces: `Card` component — props/API unchanged (`variant`, `padding`, `header`, `footer`).

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/Card/Card.tsx`**

```tsx
"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface CardProps extends ComponentPropsWithoutRef<"div"> {
  variant?: "elevated" | "outlined" | "filled";
  padding?: "none" | "sm" | "md" | "lg";
  header?: ReactNode;
  footer?: ReactNode;
}

const variantStyles = {
  elevated:
    "bg-paper-raised border border-line shadow-[0_1px_3px_rgba(37,35,30,0.05),0_4px_12px_rgba(37,35,30,0.04)] dark:shadow-[0_1px_3px_rgba(0,0,0,0.2),0_4px_12px_rgba(0,0,0,0.15)]",
  outlined: "bg-paper-raised border border-line",
  filled: "bg-paper-sunken",
};

const paddingStyles = {
  none: "",
  sm: "p-4",
  md: "p-6",
  lg: "p-8",
};

export const Card = forwardRef<HTMLDivElement, CardProps>(
  (
    {
      variant = "elevated",
      padding = "md",
      header,
      footer,
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          "rounded-[var(--radius-md)] overflow-hidden",
          variantStyles[variant],
          !header && !footer && paddingStyles[padding],
          className
        )}
        {...props}
      >
        {header && (
          <div className={cn("border-b border-line", paddingStyles[padding])}>
            {header}
          </div>
        )}
        {header || footer ? (
          <div className={paddingStyles[padding]}>{children}</div>
        ) : (
          children
        )}
        {footer && (
          <div className={cn("border-t border-line", paddingStyles[padding])}>
            {footer}
          </div>
        )}
      </div>
    );
  }
);

Card.displayName = "Card";
```

Design note: `elevated` previously had no border at all (`bg-paper shadow-[...]`, nothing else) — it now gets `border border-line` unconditionally, because the kit's Card *always* carries a felt-tip hairline border regardless of elevation ("Everything is drawn with a felt-tip hairline" — spec/readme). This is a deliberate brand-required visual change, not scope creep. No hover-lift/interactive transform is added: the kit's Card has an `interactive` prop gating a hover-raise animation, but our `CardProps` has no such prop, and adding one would be an API change (ruled out). Applying a hover-raise unconditionally to every `Card` — including static, non-clickable ones — would be a real UX regression, so it's deliberately left out. `Button`'s primary variant is the only place the press/lift interaction landed (Task 4), since a button's click affordance is unambiguous.

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/Card/Card.tsx
git commit -m "design-system: restyle Card for the paper/ink/terracotta brand"
```

---

### Task 7: Restyle `Badge`

**Files:**
- Modify: `packages/ui-web/src/primitives/Badge/Badge.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2 (`font-body`, `bg-ink`, `text-paper-raised`, `bg-paper-sunken`, `text-ink`, `border-line`)
- Produces: `Badge` component — props/API unchanged (`variant`, `color`, `size`). `success`/`warning`/`danger` colors keep their existing raw Tailwind green/amber/red values — those were never part of the "zen" system and are out of scope for this rebrand (the spec only covers the paper/ink/terracotta system, not a semantic-status-color overhaul).

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/Badge/Badge.tsx`**

```tsx
"use client";

import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface BadgeProps extends ComponentPropsWithoutRef<"span"> {
  variant?: "solid" | "subtle" | "outline";
  color?: "default" | "accent" | "success" | "warning" | "danger";
  size?: "sm" | "md";
}

const baseStyles =
  "inline-flex items-center font-body font-semibold uppercase tracking-wide rounded-[var(--radius-sm)]";

const sizeStyles = {
  sm: "px-2 py-0.5 text-xs",
  md: "px-2.5 py-1 text-sm",
};

const variantColorStyles = {
  solid: {
    default: "bg-ink text-paper-raised",
    accent: "bg-accent text-white",
    success: "bg-green-600 text-white",
    warning: "bg-amber-500 text-white",
    danger: "bg-red-500 text-white",
  },
  subtle: {
    default: "bg-paper-sunken text-ink",
    accent: "bg-accent/10 text-accent",
    success: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
    warning: "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400",
    danger: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
  },
  outline: {
    default: "border border-line text-ink",
    accent: "border border-accent text-accent",
    success: "border border-green-500 text-green-600 dark:text-green-400",
    warning: "border border-amber-500 text-amber-600 dark:text-amber-400",
    danger: "border border-red-500 text-red-600 dark:text-red-400",
  },
};

export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  (
    {
      variant = "subtle",
      color = "default",
      size = "sm",
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <span
        ref={ref}
        className={cn(
          baseStyles,
          sizeStyles[size],
          variantColorStyles[variant][color],
          className
        )}
        {...props}
      >
        {children}
      </span>
    );
  }
);

Badge.displayName = "Badge";
```

Design note: `uppercase tracking-wide` is new — the previous Badge had no text-transform at all. The kit's Badge spec explicitly calls it "an uppercase status/category marker," so this is a deliberate brand-fidelity addition, using Tailwind's stock `tracking-wide` (0.025em) rather than introducing another custom CSS variable for the kit's exact 0.04em, since it's visually indistinguishable at badge sizes.

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/Badge/Badge.tsx
git commit -m "design-system: restyle Badge for the paper/ink/terracotta brand"
```

---

### Task 8: Restyle `Input`

**Files:**
- Modify: `packages/ui-web/src/primitives/Input/Input.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2/3 (`bg-paper-raised`, `border-line`, `text-ink`, `text-ink-soft`, `rounded-[var(--radius-md)]`, `border-[length:var(--border-width)]`)
- Produces: `Input` component — props/API unchanged (`label`, `error`, `helperText`, `icon`, `size`).

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/Input/Input.tsx`**

```tsx
"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface InputProps
  extends Omit<ComponentPropsWithoutRef<"input">, "size"> {
  label?: string;
  error?: string;
  helperText?: string;
  icon?: ReactNode;
  size?: "sm" | "md" | "lg";
}

const sizeStyles = {
  sm: "px-3 py-2 text-sm",
  md: "px-4 py-3 text-base",
  lg: "px-5 py-4 text-base",
};

export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      label,
      error,
      helperText,
      icon,
      size = "md",
      className,
      id,
      ...props
    },
    ref
  ) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, "-");

    return (
      <div className="w-full">
        {label && (
          <label
            htmlFor={inputId}
            className="block font-semibold text-ink text-base mb-2"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {icon && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-soft">
              {icon}
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            className={cn(
              "w-full bg-paper-raised border-[length:var(--border-width)] rounded-[var(--radius-md)] transition-all",
              "focus:outline-none focus:ring-2 focus:ring-ink/20 focus:border-ink",
              "text-ink placeholder-ink-soft",
              error
                ? "border-red-500 focus:ring-red-500/20 focus:border-red-500"
                : "border-line",
              sizeStyles[size],
              icon && "pl-10",
              className
            )}
            {...props}
          />
        </div>
        {error && (
          <p className="mt-1.5 text-sm text-red-500">{error}</p>
        )}
        {helperText && !error && (
          <p className="mt-1.5 text-sm text-ink-soft">{helperText}</p>
        )}
      </div>
    );
  }
);

Input.displayName = "Input";
```

Design note: background changed from the old `bg-zen-subtle` (recessed fill) to `bg-paper-raised`, matching the kit's own Input spec exactly (`background: var(--paper-raised)`), not a straight 1:1 token substitution — this is the one Input field where the kit picks a different surface than what "subtle" would map to.

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/Input/Input.tsx
git commit -m "design-system: restyle Input for the paper/ink/terracotta brand"
```

---

### Task 9: Restyle `Avatar`

**Files:**
- Modify: `packages/ui-web/src/primitives/Avatar/Avatar.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2/3 (`bg-paper-raised`, `text-ink-soft`, `border-line`, `font-marker`, `rounded-[var(--radius-md)]`)
- Produces: `Avatar` component — props/API unchanged (`src`, `alt`, `fallback`, `size`, `shape`).

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/Avatar/Avatar.tsx`**

```tsx
"use client";

import { forwardRef, useState, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface AvatarProps extends ComponentPropsWithoutRef<"div"> {
  src?: string;
  alt?: string;
  fallback?: string;
  size?: "sm" | "md" | "lg" | "xl";
  shape?: "circle" | "square";
}

const sizeStyles = {
  sm: "w-8 h-8 text-xs",
  md: "w-10 h-10 text-sm",
  lg: "w-12 h-12 text-base",
  xl: "w-16 h-16 text-lg",
};

const shapeStyles = {
  circle: "rounded-full",
  square: "rounded-[var(--radius-md)]",
};

export const Avatar = forwardRef<HTMLDivElement, AvatarProps>(
  (
    {
      src,
      alt = "",
      fallback,
      size = "md",
      shape = "circle",
      className,
      ...props
    },
    ref
  ) => {
    const [imgError, setImgError] = useState(false);
    const showFallback = !src || imgError;

    const initials = fallback
      ?.split(" ")
      .map((w) => w[0])
      .join("")
      .slice(0, 2)
      .toUpperCase();

    return (
      <div
        ref={ref}
        className={cn(
          "relative inline-flex items-center justify-center overflow-hidden",
          "bg-paper-raised text-ink-soft font-marker font-semibold border border-line",
          sizeStyles[size],
          shapeStyles[shape],
          className
        )}
        {...props}
      >
        {showFallback ? (
          <span>{initials || "?"}</span>
        ) : (
          <img
            src={src}
            alt={alt}
            onError={() => setImgError(true)}
            className="w-full h-full object-cover"
          />
        )}
      </div>
    );
  }
);

Avatar.displayName = "Avatar";
```

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/Avatar/Avatar.tsx
git commit -m "design-system: restyle Avatar for the paper/ink/terracotta brand"
```

---

### Task 10: Restyle `IconButton`

**Files:**
- Modify: `packages/ui-web/src/primitives/IconButton/IconButton.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2 (`bg-ink`, `text-paper-raised`, `bg-paper-sunken`, `text-ink`, `border-line`)
- Produces: `IconButton` component — props/API unchanged (`icon`, `variant`, `size`, `label`).

- [ ] **Step 1: Replace `packages/ui-web/src/primitives/IconButton/IconButton.tsx`**

```tsx
"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface IconButtonProps
  extends Omit<ComponentPropsWithoutRef<"button">, "children"> {
  icon: ReactNode;
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  label: string;
}

const variantStyles = {
  primary: "bg-ink text-paper-raised hover:opacity-90",
  secondary: "bg-paper-sunken text-ink hover:bg-line",
  ghost: "bg-transparent text-ink hover:bg-paper-sunken",
  danger: "bg-red-500 text-white hover:bg-red-600",
};

const sizeStyles = {
  sm: "p-1.5 rounded-[var(--radius-sm)]",
  md: "p-2.5 rounded-[var(--radius-sm)]",
  lg: "p-3 rounded-[var(--radius-md)]",
};

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  (
    {
      icon,
      variant = "ghost",
      size = "md",
      label,
      disabled,
      className,
      ...props
    },
    ref
  ) => {
    return (
      <button
        ref={ref}
        aria-label={label}
        disabled={disabled}
        className={cn(
          "inline-flex items-center justify-center transition-all duration-200",
          "disabled:opacity-45 disabled:pointer-events-none",
          variantStyles[variant],
          sizeStyles[size],
          className
        )}
        {...props}
      >
        {icon}
      </button>
    );
  }
);

IconButton.displayName = "IconButton";
```

- [ ] **Step 2: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add packages/ui-web/src/primitives/IconButton/IconButton.tsx
git commit -m "design-system: restyle IconButton for the paper/ink/terracotta brand"
```

---

### Task 11: Mechanical class rename in `Dropdown`, `Modal`, `Toast`, `Divider`

**Files:**
- Modify: `packages/ui-web/src/complex/Dropdown/Dropdown.tsx`
- Modify: `packages/ui-web/src/complex/Modal/Modal.tsx`
- Modify: `packages/ui-web/src/complex/Toast/Toast.tsx`
- Modify: `packages/ui-web/src/layout/Divider/Divider.tsx`

**Interfaces:**
- Consumes: Tailwind classes from Task 2 (`bg-paper-raised`, `border-line`, `text-ink`, `text-ink-soft`, `bg-paper-sunken`, `bg-ink`, `text-paper-raised`)
- Produces: no interface change — these four files reference the old `zen-*`/`bg-paper` (old) class names directly (confirmed by grep; the design spec's Section B undersold this — it said layout/complex components "don't hardcode brand-specific styling," which is true for `Stack`/`Container`/`Spacer`/`ToastProvider` but not these four). This is a straight rename, not a restyle — same visual intent, new class names, no new interactions or design decisions.

- [ ] **Step 1: Rename in `Dropdown.tsx`**

Read the file, then verify these two exact lines exist before editing (line numbers as of this plan's writing — re-check if they've drifted):
- Line 70: `"bg-paper border border-zen-border rounded-xl shadow-lg",`
- Line 89: `: "text-zen-text hover:bg-zen-subtle"`

Change to:
```tsx
"bg-paper-raised border border-line rounded-xl shadow-lg",
```
and
```tsx
: "text-ink hover:bg-paper-sunken"
```

- [ ] **Step 2: Rename in `Modal.tsx`**

Verify these lines exist, then update:
- Line 70: `"relative w-full bg-paper rounded-2xl shadow-xl",` → `"relative w-full bg-paper-raised rounded-2xl shadow-xl",`
- Line 77: `<div className="flex items-center justify-between px-6 py-4 border-b border-zen-border">` → `<div className="flex items-center justify-between px-6 py-4 border-b border-line">`
- Line 78: `<h2 className="text-lg font-semibold text-zen-text">{title}</h2>` → `<h2 className="text-lg font-semibold text-ink">{title}</h2>`
- Line 81: `className="p-1.5 rounded-lg text-zen-muted hover:bg-zen-subtle hover:text-zen-text transition-colors"` → `className="p-1.5 rounded-lg text-ink-soft hover:bg-paper-sunken hover:text-ink transition-colors"`

- [ ] **Step 3: Rename in `Toast.tsx`**

Verify this line exists, then update:
- Line 26: `"bg-zen-text text-zen-bg",` → `"bg-ink text-paper",`

- [ ] **Step 4: Rename in `Divider.tsx`**

Verify this line exists, then update:
- Line 21: `const colorStyle = color === "subtle" ? "border-zen-subtle" : "border-zen-border";` → `const colorStyle = color === "subtle" ? "border-paper-sunken" : "border-line";`

- [ ] **Step 5: Verify no leftovers**

```bash
cd /Users/lto/repos/personal/movoz
grep -rnE "zen-(bg|text|muted|subtle|border|paper)|\bbg-paper\b" packages/ui-web/src
```

Expected: no output.

- [ ] **Step 6: Build and verify**

```bash
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 7: Commit**

```bash
git add packages/ui-web/src/complex/Dropdown/Dropdown.tsx packages/ui-web/src/complex/Modal/Modal.tsx packages/ui-web/src/complex/Toast/Toast.tsx packages/ui-web/src/layout/Divider/Divider.tsx
git commit -m "design-system: rename zen-* classes in Dropdown, Modal, Toast, Divider"
```

---

### Task 12: Add `Skeleton` primitive

**Files:**
- Create: `packages/ui-web/src/primitives/Skeleton/Skeleton.tsx`
- Create: `packages/ui-web/src/primitives/Skeleton/index.ts`
- Modify: `packages/ui-web/src/index.ts`

**Interfaces:**
- Consumes: `cn` from `../../utils/cn`, `bg-pencil` Tailwind class from Task 2
- Produces: `Skeleton({ lines, width, height, className, ...props })` — pencil-gray loading placeholder bars. Exported from `@movoz/ui-web`'s root barrel for the first time in this plan (later tasks add three more).

- [ ] **Step 1: Create `packages/ui-web/src/primitives/Skeleton/Skeleton.tsx`**

```tsx
import { type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface SkeletonProps extends ComponentPropsWithoutRef<"div"> {
  lines?: number;
  width?: string | number;
  height?: string | number;
}

export function Skeleton({
  lines = 1,
  width = "100%",
  height = 12,
  className,
  style,
  ...props
}: SkeletonProps) {
  const barHeight = typeof height === "number" ? `${height}px` : height;
  const widths = Array.from({ length: lines }, (_, i) =>
    lines > 1 && i === lines - 1
      ? "62%"
      : typeof width === "number"
        ? `${width}px`
        : width
  );

  return (
    <div className={cn("flex flex-col gap-2.5", className)} style={style} {...props}>
      {widths.map((w, i) => (
        <div key={i} className="rounded-full bg-pencil" style={{ width: w, height: barHeight }} />
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Create the barrel `packages/ui-web/src/primitives/Skeleton/index.ts`**

```ts
export { Skeleton } from "./Skeleton";
export type { SkeletonProps } from "./Skeleton";
```

- [ ] **Step 3: Add the export to `packages/ui-web/src/index.ts`**

Add after the `IconButton` export block:

```ts
export { Skeleton } from "./primitives/Skeleton";
export type { SkeletonProps } from "./primitives/Skeleton";
```

- [ ] **Step 4: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add packages/ui-web/src/primitives/Skeleton packages/ui-web/src/index.ts
git commit -m "design-system: add Skeleton primitive to @movoz/ui-web"
```

---

### Task 13: Add `Pill` primitive

**Files:**
- Create: `packages/ui-web/src/primitives/Pill/Pill.tsx`
- Create: `packages/ui-web/src/primitives/Pill/index.ts`
- Modify: `packages/ui-web/src/index.ts`

**Interfaces:**
- Consumes: `cn` from `../../utils/cn`, `bg-paper-raised`/`bg-paper-sunken`/`border-line`/`text-ink`/`font-marker` Tailwind classes from Task 2/3
- Produces: `Pill({ active, className, children, ...props })` (extends `ComponentPropsWithoutRef<"button">`) — rounded filter/toggle chip.

- [ ] **Step 1: Create `packages/ui-web/src/primitives/Pill/Pill.tsx`**

```tsx
"use client";

import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface PillProps extends ComponentPropsWithoutRef<"button"> {
  active?: boolean;
}

export const Pill = forwardRef<HTMLButtonElement, PillProps>(
  ({ active = false, className, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        type="button"
        aria-pressed={active}
        className={cn(
          "inline-flex items-center gap-1.5 font-marker font-medium text-base leading-none",
          "px-[18px] py-2 rounded-full border transition-colors duration-150",
          active
            ? "bg-paper-sunken border-line text-ink"
            : "bg-paper-raised border-line text-ink hover:bg-paper-sunken",
          className
        )}
        {...props}
      >
        {children}
      </button>
    );
  }
);

Pill.displayName = "Pill";
```

- [ ] **Step 2: Create the barrel `packages/ui-web/src/primitives/Pill/index.ts`**

```ts
export { Pill } from "./Pill";
export type { PillProps } from "./Pill";
```

- [ ] **Step 3: Add the export to `packages/ui-web/src/index.ts`**

Add after the `Skeleton` export block (from Task 12):

```ts
export { Pill } from "./primitives/Pill";
export type { PillProps } from "./primitives/Pill";
```

- [ ] **Step 4: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add packages/ui-web/src/primitives/Pill packages/ui-web/src/index.ts
git commit -m "design-system: add Pill primitive to @movoz/ui-web"
```

---

### Task 14: Add `Tabs` primitive

**Files:**
- Create: `packages/ui-web/src/primitives/Tabs/Tabs.tsx`
- Create: `packages/ui-web/src/primitives/Tabs/index.ts`
- Modify: `packages/ui-web/src/index.ts`

**Interfaces:**
- Consumes: `cn` from `../../utils/cn`, `bg-dock`/`border-dock-line`/`bg-paper`/`bg-pencil-light`/`text-pencil`/`bg-accent` Tailwind classes from Task 2/3
- Produces: `Tabs({ items, value, onChange, tone, className })`, `TabItem { value, label, accent? }` type — segmented control with a `"light"` (in-page) and `"dock"` (dark floating bar) tone.

- [ ] **Step 1: Create `packages/ui-web/src/primitives/Tabs/Tabs.tsx`**

```tsx
"use client";

import { cn } from "../../utils/cn";

export interface TabItem {
  value: string;
  label: string;
  accent?: boolean;
}

export interface TabsProps {
  items: (string | TabItem)[];
  value: string;
  onChange: (value: string) => void;
  tone?: "light" | "dock";
  className?: string;
}

export function Tabs({ items, value, onChange, tone = "light", className }: TabsProps) {
  const dark = tone === "dock";

  return (
    <div
      role="tablist"
      className={cn(
        "inline-flex items-center gap-1.5 p-1.5 rounded-[var(--radius-lg)] border",
        dark ? "bg-dock border-dock-line" : "bg-paper-raised border-line",
        className
      )}
    >
      {items.map((item) => {
        const itemValue = typeof item === "string" ? item : item.value;
        const label = typeof item === "string" ? item : item.label;
        const accent = typeof item === "object" && item.accent;
        const active = itemValue === value;

        return (
          <button
            key={itemValue}
            type="button"
            role="tab"
            aria-selected={active}
            onClick={() => onChange(itemValue)}
            className={cn(
              "px-4 py-2 rounded-[var(--radius-md)] font-marker font-semibold text-base leading-none transition-colors duration-150",
              active
                ? accent
                  ? "bg-accent text-white"
                  : dark
                    ? "bg-paper text-ink"
                    : "bg-pencil-light text-ink"
                : accent
                  ? "text-accent border border-accent"
                  : dark
                    ? "text-pencil border border-transparent"
                    : "text-ink-soft border border-transparent"
            )}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 2: Create the barrel `packages/ui-web/src/primitives/Tabs/index.ts`**

```ts
export { Tabs } from "./Tabs";
export type { TabsProps, TabItem } from "./Tabs";
```

- [ ] **Step 3: Add the export to `packages/ui-web/src/index.ts`**

Add after the `Pill` export block (from Task 13):

```ts
export { Tabs } from "./primitives/Tabs";
export type { TabsProps, TabItem } from "./primitives/Tabs";
```

- [ ] **Step 4: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add packages/ui-web/src/primitives/Tabs packages/ui-web/src/index.ts
git commit -m "design-system: add Tabs primitive to @movoz/ui-web"
```

---

### Task 15: Add `PlaceholderBox` primitive

**Files:**
- Create: `packages/ui-web/src/primitives/PlaceholderBox/PlaceholderBox.tsx`
- Create: `packages/ui-web/src/primitives/PlaceholderBox/index.ts`
- Modify: `packages/ui-web/src/index.ts`

**Interfaces:**
- Consumes: `cn` from `../../utils/cn`, `bg-paper-raised`/`border-line`/`text-ink-soft`/`text-pencil`/`font-marker` Tailwind classes from Task 2/3
- Produces: `PlaceholderBox({ label, ratio, cross, dashed, height, className })` — crossed placeholder frame for missing images/content.

- [ ] **Step 1: Create `packages/ui-web/src/primitives/PlaceholderBox/PlaceholderBox.tsx`**

```tsx
import { cn } from "../../utils/cn";

export interface PlaceholderBoxProps {
  label?: string;
  ratio?: string;
  cross?: boolean;
  dashed?: boolean;
  height?: string | number;
  className?: string;
}

export function PlaceholderBox({
  label = "",
  ratio = "16 / 9",
  cross = true,
  dashed = false,
  height,
  className,
}: PlaceholderBoxProps) {
  return (
    <div
      className={cn(
        "relative flex items-center justify-center overflow-hidden w-full",
        "bg-paper-raised border-line rounded-[var(--radius-sm)]",
        "border-[length:var(--border-width)]",
        dashed ? "border-dashed" : "border-solid",
        "text-ink-soft font-marker italic",
        className
      )}
      style={{ aspectRatio: height ? undefined : ratio, height: height ?? undefined }}
    >
      {cross && (
        <svg
          className="absolute inset-0 text-pencil"
          preserveAspectRatio="none"
          viewBox="0 0 100 100"
        >
          <line x1="0" y1="0" x2="100" y2="100" stroke="currentColor" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
          <line x1="100" y1="0" x2="0" y2="100" stroke="currentColor" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
        </svg>
      )}
      {label && <span className="relative bg-paper-raised px-2">{label}</span>}
    </div>
  );
}
```

- [ ] **Step 2: Create the barrel `packages/ui-web/src/primitives/PlaceholderBox/index.ts`**

```ts
export { PlaceholderBox } from "./PlaceholderBox";
export type { PlaceholderBoxProps } from "./PlaceholderBox";
```

- [ ] **Step 3: Add the export to `packages/ui-web/src/index.ts`**

Add after the `Tabs` export block (from Task 14):

```ts
export { PlaceholderBox } from "./primitives/PlaceholderBox";
export type { PlaceholderBoxProps } from "./primitives/PlaceholderBox";
```

- [ ] **Step 4: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter @movoz/ui-web lint
pnpm --filter @movoz/ui-web build
```

Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add packages/ui-web/src/primitives/PlaceholderBox packages/ui-web/src/index.ts
git commit -m "design-system: add PlaceholderBox primitive to @movoz/ui-web"
```

---

### Task 16: Mechanical rename across `apps/personal-site`

**Files:**
- Modify: `apps/personal-site/src/app/globals.css`
- Modify: `apps/personal-site/src/components/Hero.tsx`
- Modify: `apps/personal-site/src/components/About.tsx`
- Modify: `apps/personal-site/src/components/Projects.tsx`
- Modify: `apps/personal-site/src/components/Contact.tsx`
- Modify: `apps/personal-site/src/components/Navigation.tsx`

**Interfaces:**
- Consumes: every renamed class/CSS var from Tasks 1–3 (`bg-paper`, `text-ink`, `text-ink-soft`, `border-line`, `bg-paper-sunken`, `--paper`, `--ink`, `font-marker`) and the restyled `Text`/`Button` components from Tasks 4–5 (`font="marker"` prop value)
- Produces: no new interface — personal-site keeps building and rendering, just with the new brand's class/token names in place of the old ones. Confirmed by grep (see setup below) that personal-site never uses bare `bg-paper` (the pre-rename ui-web-internal collision case) or `font-sans`/`font-ui`, so this is a safe, order-independent global substitution.

- [ ] **Step 1: Apply the mechanical rename**

```bash
cd /Users/lto/repos/personal/movoz
perl -pi -e '
  s/zen-border/line/g;
  s/zen-subtle/paper-sunken/g;
  s/zen-paper/paper-raised/g;
  s/zen-muted/ink-soft/g;
  s/zen-text/ink/g;
  s/zen-bg/paper/g;
  s/font-serif/font-marker/g;
  s/font="serif"/font="marker"/g;
' apps/personal-site/src/app/globals.css \
  apps/personal-site/src/components/Hero.tsx \
  apps/personal-site/src/components/About.tsx \
  apps/personal-site/src/components/Projects.tsx \
  apps/personal-site/src/components/Contact.tsx \
  apps/personal-site/src/components/Navigation.tsx
```

- [ ] **Step 2: Verify no leftovers**

```bash
grep -rnE "zen-(bg|text|muted|subtle|border|paper)|font-serif|font=\"serif\"" apps/personal-site/src
```

Expected: no output. (personal-site's own bespoke `.font-ui`, `.glass`, `.paper-card`, `.gradient-text` classes in `globals.css` are untouched — they don't match this rename table and are explicitly out of scope per the spec.)

- [ ] **Step 3: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter personal-site lint
pnpm --filter personal-site build
```

Expected: both clean.

- [ ] **Step 4: Commit**

```bash
git add apps/personal-site/src
git commit -m "design-system: rename zen-*/font-serif class references in personal-site"
```

---

### Task 17: Mechanical rename across `apps/oncarinho`

**Files:**
- Modify: `apps/oncarinho/src/app/globals.css`
- Modify: `apps/oncarinho/src/app/page.tsx`
- Modify: `apps/oncarinho/src/app/leaderboard/page.tsx`
- Modify: `apps/oncarinho/src/app/players/[id]/page.tsx`
- Modify: `apps/oncarinho/src/app/players/[id]/not-found.tsx`
- Modify: `apps/oncarinho/src/app/admin/page.tsx`
- Modify: `apps/oncarinho/src/app/admin/players/page.tsx`
- Modify: `apps/oncarinho/src/app/admin/matchdays/page.tsx`
- Modify: `apps/oncarinho/src/app/admin/matchdays/[id]/page.tsx`
- Modify: `apps/oncarinho/src/components/nav/Nav.tsx`
- Modify: `apps/oncarinho/src/components/nav/LanguageToggle.tsx`
- Modify: `apps/oncarinho/src/components/StatTile.tsx`
- Modify: `apps/oncarinho/src/components/StatTypeTabs.tsx`
- Modify: `apps/oncarinho/src/components/SeasonSelector.tsx`

**Interfaces:**
- Consumes: same renamed classes/tokens as Task 16
- Produces: no new interface — same mechanical guarantee as Task 16, confirmed by grep that oncarinho never uses bare `bg-paper` or `font-sans`/`font-ui` either.

- [ ] **Step 1: Apply the mechanical rename**

```bash
cd /Users/lto/repos/personal/movoz
perl -pi -e '
  s/zen-border/line/g;
  s/zen-subtle/paper-sunken/g;
  s/zen-paper/paper-raised/g;
  s/zen-muted/ink-soft/g;
  s/zen-text/ink/g;
  s/zen-bg/paper/g;
  s/font-serif/font-marker/g;
  s/font="serif"/font="marker"/g;
' apps/oncarinho/src/app/globals.css \
  apps/oncarinho/src/app/page.tsx \
  apps/oncarinho/src/app/leaderboard/page.tsx \
  "apps/oncarinho/src/app/players/[id]/page.tsx" \
  "apps/oncarinho/src/app/players/[id]/not-found.tsx" \
  apps/oncarinho/src/app/admin/page.tsx \
  apps/oncarinho/src/app/admin/players/page.tsx \
  apps/oncarinho/src/app/admin/matchdays/page.tsx \
  "apps/oncarinho/src/app/admin/matchdays/[id]/page.tsx" \
  apps/oncarinho/src/components/nav/Nav.tsx \
  apps/oncarinho/src/components/nav/LanguageToggle.tsx \
  apps/oncarinho/src/components/StatTile.tsx \
  apps/oncarinho/src/components/StatTypeTabs.tsx \
  apps/oncarinho/src/components/SeasonSelector.tsx
```

- [ ] **Step 2: Verify no leftovers**

```bash
grep -rnE "zen-(bg|text|muted|subtle|border|paper)|font-serif|font=\"serif\"" apps/oncarinho/src
```

Expected: no output.

- [ ] **Step 3: Build and verify**

```bash
cd /Users/lto/repos/personal/movoz
pnpm --filter oncarinho lint
pnpm --filter oncarinho build
```

Expected: both clean.

- [ ] **Step 4: Commit**

```bash
git add apps/oncarinho/src
git commit -m "design-system: rename zen-*/font-serif class references in oncarinho"
```

---

### Task 18: Install the `movoz-design` Claude skill

**Files:**
- Create: `.claude/skills/movoz-design/` (copy of the extracted kit)

**Interfaces:**
- Consumes: the extracted kit contents (session-scoped scratch path noted at the top of this plan, or re-extract from `/Users/lto/Downloads/Movoz Design System.zip` if that path no longer exists)
- Produces: an installed, user-invocable Claude skill named `movoz-design`, independent of the package rebrand — no other task depends on this one, and it doesn't depend on Tasks 1–17 either. It can run at any point in this plan.

- [ ] **Step 1: Locate or re-extract the kit**

```bash
KIT_DIR="/private/tmp/claude-502/-Users-lto-repos-personal-movoz/106e165e-26c9-4e0b-bb04-0fa3045c5a74/scratchpad/movoz-design-system"
if [ ! -d "$KIT_DIR" ]; then
  KIT_DIR="/tmp/movoz-design-system-reextract"
  mkdir -p "$KIT_DIR"
  unzip -o "/Users/lto/Downloads/Movoz Design System.zip" -d "$KIT_DIR"
fi
echo "$KIT_DIR"
```

- [ ] **Step 2: Copy it into the skill directory**

```bash
cd /Users/lto/repos/personal/movoz
mkdir -p .claude/skills/movoz-design
cp -R "$KIT_DIR"/. .claude/skills/movoz-design/
```

- [ ] **Step 3: Verify the skill is well-formed**

```bash
test -f .claude/skills/movoz-design/SKILL.md && echo "SKILL.md present"
head -6 .claude/skills/movoz-design/SKILL.md
```

Expected: `SKILL.md present`, followed by the frontmatter showing `name: movoz-design`, `user-invocable: true`.

- [ ] **Step 4: Commit**

```bash
git add .claude/skills/movoz-design
git commit -m "design-system: install the movoz-design kit as a Claude skill"
```

---

### Task 19: Final verification pass

**Files:** none (verification only)

**Interfaces:** none — this task confirms Tasks 1–18 are consistent as a whole.

- [ ] **Step 1: Full workspace build and lint**

```bash
cd /Users/lto/repos/personal/movoz
pnpm build
pnpm lint
```

Expected: both clean. `turbo build`'s `dependsOn: ["^build"]` means `@movoz/tokens` builds before anything that imports it, so this also re-confirms Task 1's build artifact is current. Packages without a `lint` script (`@movoz/theme`, `@movoz/tokens`, `@movoz/tailwind-config`) are silently skipped by turbo, not an error.

- [ ] **Step 2: Repo-wide grep sweep for leftover old references**

```bash
grep -rnE "zen-(bg|text|muted|subtle|border|paper)" apps/personal-site/src apps/oncarinho/src packages/theme/src packages/ui-web/src
grep -rnE "\bfont-serif\b|font=\"serif\"|\bfont-ui\b|font=\"ui\"" apps/personal-site/src apps/oncarinho/src packages/ui-web/src
grep -rnE "\bbg-paper\b" packages/ui-web/src
```

Expected: no output from any of the three commands. (`apps/drunken-dolphin` is intentionally excluded — it never used these tokens.)

- [ ] **Step 3: Manual visual check**

```bash
pnpm dev
```

Visit personal-site and oncarinho (check the ports each app's `package.json` dev script binds — personal-site is the default zone, oncarinho runs on port 3100 per its `dev` script). For each app: toggle light/dark mode and confirm the paper/ink palette and terracotta accent render, confirm Shantell Sans renders on headings (`font-marker`/`font="marker"` usages) and Space Grotesk on body text, confirm felt-tip hairline borders are visible on cards/buttons/inputs. Accept whatever the pages look like compositionally — page-level polish is explicitly out of scope (spec Decisions #2); the only failure condition here is a token/class that didn't resolve (e.g. visibly unstyled/default-browser-font text, which would mean a rename was missed).

- [ ] **Step 4: Report results**

Summarize: build/lint status for all 4 packages + 2 apps, grep sweep results, and what was visually confirmed in light/dark mode for each app. Flag anything that didn't render as expected — do not silently patch further without noting it, since this is the final gate before the rebrand is considered done.
