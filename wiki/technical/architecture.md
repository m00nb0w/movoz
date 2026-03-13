# Architecture Overview

## System Diagram

```
movoz/
├── apps/                          # Frontend applications (pnpm workspace)
│   └── personal-site/             # Next.js 14 — default zone at /
│
├── packages/                      # Shared packages (pnpm workspace)
│   ├── tokens/                    # @movoz/tokens — platform-agnostic design tokens
│   ├── ui-web/                    # @movoz/ui-web — React + Tailwind component library
│   ├── tailwind-config/           # @movoz/tailwind-config — shared Tailwind preset
│   ├── theme/                     # @movoz/theme — ThemeProvider, ThemeToggle, CSS vars
│   └── tsconfig/                  # @movoz/tsconfig — shared TypeScript configs
│
├── backend/                       # Backend services (outside pnpm workspace)
│   ├── drunken-dolphin/           # Rust CLI — personal fitness tracking
│   └── hustle-turtle/             # Go REST API — Gin + PostgreSQL
│
├── infra/                         # Infrastructure (outside pnpm workspace)
│   ├── terraform/                 # AWS VPC + RDS (workspace-based dev/prod)
│   ├── docker-compose.yml         # Multi-zone deployment
│   └── nginx/                     # Reverse proxy config
│
└── docs/                          # Knowledge base (this folder)
```

## Package Dependency Graph

```
@movoz/tokens          (zero deps, platform-agnostic)
    │
    ├──> @movoz/tailwind-config    (consumes tokens for fonts, colors, animations)
    │        │
    │        ├──> @movoz/ui-web    (uses tailwind-config as preset, devDep)
    │        │
    │        └──> apps/*           (uses tailwind-config as preset, devDep)
    │
    └──> @movoz/ui-web             (imports tokens directly)
              │
              └──> apps/*          (runtime dependency)

@movoz/theme           (standalone: ThemeProvider, CSS variables)
    │
    └──> apps/*                    (runtime dependency)

@movoz/tsconfig        (standalone: TypeScript base configs)
    │
    └──> all packages + apps       (devDep)
```

## Key Architectural Decisions

### Multi-Zone Frontend

Apps use Next.js Multi-Zones. Each app in `apps/` is an independent Next.js zone served under a path prefix on the same domain. In production, Nginx routes requests to the correct zone. In development, Next.js rewrites handle routing.

To add a new zone:
1. Create a new app in `apps/`
2. Set `basePath` and `assetPrefix` in its `next.config.mjs`
3. Add an Nginx route in `infra/nginx/nginx.conf`
4. Add `@movoz/ui-web` and `@movoz/theme` as dependencies
5. Add ui-web and theme to `transpilePackages` in next config
6. Add ui-web and theme content paths to `tailwind.config.ts`

### Design Token Architecture

Tokens are the single source of truth for all visual design values. The flow:

```
@movoz/tokens (TS objects)
    │
    ├──> @movoz/tailwind-config    Fonts, palette, keyframes, animations
    │                               Semantic colors stay as CSS var references
    │
    ├──> @movoz/theme/globals.css  Hex values for CSS variables (--zen-*)
    │                               Must stay in sync manually (documented)
    │
    └──> future: @movoz/ui-mobile  Same tokens, native styling
```

Semantic zen colors (`bg`, `text`, `muted`, `subtle`, `border`, `paper`) use CSS variables in Tailwind so they respond to light/dark theme. The actual hex values live in both `@movoz/tokens` (source of truth) and `@movoz/theme/globals.css` (runtime CSS variables).

### Theme System

- `ThemeProvider` wraps the app at the root layout level
- Theme state persists in `localStorage` (key: `"theme"`)
- Three modes: `light`, `dark`, `system`
- CSS variables on `:root` (light) or `.dark` class (dark) control all colors
- Components use semantic Tailwind classes (`text-zen-text`, `bg-zen-bg`) — never import from `@movoz/theme` directly

### Component Library Strategy

Two separate packages for web and future mobile, sharing tokens:

| | Web (`@movoz/ui-web`) | Mobile (future) |
|---|---|---|
| **Framework** | React + Tailwind CSS | TBD (React Native / Expo / Flutter) |
| **Styling** | Tailwind utility classes via `cn()` | Native styling using token values |
| **Tokens** | Via Tailwind config (CSS vars) | Direct import from `@movoz/tokens` |
| **Icons** | `lucide-react` (peer dep) | TBD |

### Backend Architecture

Backend services are intentionally outside the pnpm workspace — they use different languages and build systems:

- **drunken-dolphin** (Rust): CLI tool, data persisted to local JSON files
- **hustle-turtle** (Go): REST API with Gin framework, PostgreSQL via golang-migrate

### Infrastructure

- AWS VPC with public/private subnets
- PostgreSQL RDS in private subnets
- Dev: `t3.micro` / 20GB, Prod: `t3.small` / 100GB
- Terraform workspace-based environments (`dev` / `prod`)
- Weekly auto-destroy of dev resources (Sundays 2 AM UTC)

## Build Pipeline

```
pnpm build  →  Turborepo orchestrates:

1. @movoz/tokens       tsup → dist/ (ESM + CJS + .d.ts)
2. @movoz/ui-web       tsup → dist/ (ESM + CJS + .d.ts)
3. personal-site       next build → .next/ (standalone)
```

Turborepo handles dependency ordering via `dependsOn: ["^build"]`. Packages without a `build` script (`tailwind-config`, `theme`, `tsconfig`) are consumed directly from source via `transpilePackages`.
