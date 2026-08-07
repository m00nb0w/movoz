# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

Personal monorepo using pnpm workspaces + Turborepo. Follows **spec-driven development** — every non-trivial feature or change starts with a spec in `wiki/` before implementation begins.

### Directory Structure

```
movoz/
├── apps/           # Frontend applications (web, mobile)
├── packages/       # Shared frontend packages (config, theme, UI)
├── backend/        # Backend services (APIs, CLIs)
├── infra/          # Infrastructure (Terraform, Docker, Nginx, CI/CD)
├── wiki/           # Knowledge base — specs, technical docs, templates (single source of truth)
└── resources/      # Static assets and resources
```

### apps/ — Frontend Applications
- **apps/personal-site/** — Next.js 14 portfolio website (default zone at `/`)
- **apps/travel-quokka/** — Travel map, marathon tracker, and trip journal (planned)
- **apps/octopus/** — Second brain — personal knowledge management (planned)

### packages/ — Shared Frontend Packages
- **packages/tailwind-config/** — Shared Tailwind CSS preset (`@movoz/tailwind-config`)
- **packages/theme/** — ThemeProvider, ThemeToggle, CSS variables (`@movoz/theme`)
- **packages/tsconfig/** — Shared TypeScript base config (`@movoz/tsconfig`)

### backend/ — Backend Services
- **backend/drunken-dolphin/** — Rust personal assistant — finance (expenses, investments) and research
- **backend/hustle-turtle/** — Go REST API for habit tracking (push ups, reading, running)
- **backend/oncarinho/** — Go REST API for football team stats tracking (players, matchdays, leaderboards)
- **backend/peaky-bergers/** — AI software team — agents that help build stuff (planned)

### infra/ — Infrastructure
- **infra/terraform/** — Terraform (AWS) infrastructure for database deployment
- **infra/docker-compose.yml** + **infra/nginx/** — Nginx reverse proxy + Docker Compose for multi-zone deployment
- **infra/lonely-moley/** — Personal homelab — self-hosted K8s, CI/CD, databases (planned)
- **.github/workflows/** — CI/CD pipelines (GitHub Actions)

### wiki/ — Knowledge Base (Single Source of Truth)
- **wiki/technical/** — Architecture docs, design tokens, component specs, technical design templates
- **wiki/specs/** — Product specs, feature specs
- Templates: 1:3:1 decision writeups, bug fix analyses, product specs, technical designs
- Spec Kit (`/speckit-*`) writes the product spec to `wiki/specs/<feature>/spec.md` and all engineering artifacts (`plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/`, `tasks.md`) to `wiki/technical/<feature>/`.

## Spec-Driven Development

All non-trivial work follows this flow:

1. **Spec first** — Write or update a spec in `wiki/` before writing code. Specs define the what and why.
2. **Review spec** — Validate the spec covers requirements, edge cases, and acceptance criteria.
3. **Implement** — Build against the spec. The spec is the contract.
4. **Update wiki** — After implementation, update relevant wiki docs (architecture, component specs, etc.) to reflect the current state.

When starting new work, always check `wiki/` for existing specs and context.

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

### oncarinho (Go)
```bash
cd backend/oncarinho
go build -o bin/oncarinho ./cmd/server       # Build
go test ./...                                # Run all tests
go run ./cmd/server                          # Run server (port 8081)
go run ./cmd/server -auto-migrate            # Run with auto-migrations
go run ./cmd/server -migrate=up              # Apply migrations only
go run ./cmd/server -migrate=down            # Rollback migrations
go run ./cmd/server -version                 # Check migration version
```
Requires PostgreSQL, `ADMIN_PASSWORD`, and `SESSION_SECRET` env vars (the server refuses to start without the latter two). Default DATABASE_URL: `postgres://localhost/oncarinho?sslmode=disable`

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

### oncarinho — Standard Go Layout
- `cmd/server/` — Entrypoint + full route wiring (public vs. admin route groups)
- `internal/` — Private code: `config/`, `database/`, `models/`, `store/` (SQL access), `handlers/`, `auth/` (session tokens)
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

Workflows in `.github/workflows/`:
- **infrastructure-deploy.yml** — Terraform plan/apply on push to master (infra/terraform/ changes)
- **infrastructure-cost-analysis.yml** — Infracost analysis on PRs affecting infrastructure
- **infrastructure-cleanup.yml** — Weekly auto-destroy of dev resources (Sundays 2 AM UTC)

## Workflow Preferences

- Always commit and push changes at the end of a plan execution.
- Spec before code — check `wiki/` for existing context before starting any feature work.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
