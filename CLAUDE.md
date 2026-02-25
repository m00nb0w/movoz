# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

Monorepo using pnpm workspaces + Turborepo for frontend apps, with backend and infra projects organized by section.

**Frontend apps** (in `apps/`):
- **apps/personal-site/** — Next.js 14 portfolio website (default zone at `/`)

**Shared packages** (in `packages/`):
- **packages/tailwind-config/** — Shared Tailwind CSS preset (`@movoz/tailwind-config`)
- **packages/theme/** — ThemeProvider, ThemeToggle, CSS variables (`@movoz/theme`)
- **packages/tsconfig/** — Shared TypeScript base config (`@movoz/tsconfig`)

**Backend** (in `backend/`):
- **backend/drunken-dolphin/** — Rust CLI for personal fitness tracking
- **backend/hustle-turtle/** — Go REST API with Gin framework and PostgreSQL

**Infrastructure** (in `infra/`):
- **infra/terraform/** — Terraform (AWS) infrastructure for database deployment
- **infra/docker-compose.yml** + **infra/nginx/** — Nginx reverse proxy + Docker Compose for multi-zone deployment

## Build & Development Commands

### drunken-dolphin (Rust)
```bash
cd backend/drunken-dolphin
cargo build --release          # Build
cargo test                     # Run all tests
cargo run -- fitness today 25 50 15  # Run locally
cargo install --path .         # Install globally
```

### hustle-turtle (Go)
```bash
cd backend/hustle-turtle
go build -o bin/hustle-turtle ./cmd/server   # Build
go test ./...                                # Run all tests
go run ./cmd/server                          # Run server (port 8080)
go run ./cmd/server -auto-migrate            # Run with auto-migrations
go run ./cmd/server -migrate=up              # Apply migrations only
go run ./cmd/server -migrate=down            # Rollback migrations
go run ./cmd/server -version                 # Check migration version
```
Requires PostgreSQL. Default DATABASE_URL: `postgres://localhost/hustle_turtle?sslmode=disable`

### Frontend (pnpm + Turborepo)
```bash
pnpm install                           # Install all workspace dependencies
pnpm dev                               # Start all apps via Turborepo
pnpm build                             # Build all apps
pnpm --filter personal-site dev        # Start only personal-site
pnpm --filter personal-site build      # Build only personal-site
pnpm --filter personal-site lint       # Lint only personal-site
```

### Docker (multi-zone deployment)
```bash
docker compose -f infra/docker-compose.yml up --build
```

### Terraform
```bash
cd infra/terraform
terraform init
terraform fmt -check -recursive   # Format check
terraform validate                # Validate config
terraform plan                    # Preview changes
```
Uses workspace-based environments (dev/prod). Region: us-west-2.

## Architecture Notes

### Micro Frontend (Multi-Zones)
Frontend uses Next.js Multi-Zones architecture. Each app in `apps/` is an independent Next.js zone served under a path prefix on the same domain. Nginx routes in production; Next.js rewrites route in dev. Theme state syncs across zones via shared localStorage. To add a new zone: create app in `apps/`, add `basePath`/`assetPrefix` to its `next.config.mjs`, add Nginx route in `infra/nginx/nginx.conf`.

### Shared Packages
All frontend apps should use `@movoz/tailwind-config` as a Tailwind preset and `@movoz/theme` for the theme system. The `@movoz/tsconfig/nextjs.json` provides a shared TypeScript base.

### hustle-turtle — Standard Go Layout
- `cmd/server/` — Entrypoint with CLI flags for migration control
- `internal/` — Private code: `config/`, `database/`, `handlers/`, `models/`
- `pkg/utils/` — Public utility code
- `migrations/` — SQL migration files (golang-migrate)

### drunken-dolphin — Modular CLI
- `src/main.rs` — CLI definition (clap)
- `src/commands/` — One module per exercise type (pushups.rs, situps.rs, pullups.rs)
- `src/personal.rs` — Core data management; persists to JSON files
- Adding new commands: create module in `src/commands/`, register in `mod.rs` and `main.rs`

### personal-site — Next.js App Router (default zone)
- `src/app/` — App Router pages
- `src/components/` — Page-specific components (Hero, Projects, About, Contact, Navigation, Footer)
- Theme (ThemeProvider, ThemeToggle) lives in `@movoz/theme` shared package
- Standalone output mode for Docker deployment

### Terraform — AWS VPC + RDS
- VPC with public/private subnets; PostgreSQL RDS in private subnets
- Dev: t3.micro/20GB, Prod: t3.small/100GB with enhanced monitoring
- `scripts/deploy.sh` for manual deployment

## CI/CD (GitHub Actions)

Workflows in `.github/workflows/` focus on infrastructure:
- **infrastructure-deploy.yml** — Terraform plan/apply on push to master (infra/terraform/ changes)
- **infrastructure-cost-analysis.yml** — Infracost analysis on PRs affecting infrastructure
- **infrastructure-cleanup.yml** — Weekly auto-destroy of dev resources (Sundays 2 AM UTC)

## Workflow Preferences

- Always commit and push changes at the end of a plan execution.
