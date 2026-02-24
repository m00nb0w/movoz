# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a multi-project monorepo containing four independent projects with no shared build tooling:

- **drunken-dolphin/** — Rust CLI for personal fitness tracking (push-ups, sit-ups, pull-ups)
- **hustle-turtle/** — Go REST API with Gin framework and PostgreSQL
- **personal-site/** — Next.js 14 portfolio website with Tailwind CSS
- **movoz-infra/** — Terraform (AWS) infrastructure for database deployment

## Build & Development Commands

### drunken-dolphin (Rust)
```bash
cd drunken-dolphin
cargo build --release          # Build
cargo test                     # Run all tests
cargo run -- fitness today 25 50 15  # Run locally
cargo install --path .         # Install globally
```

### hustle-turtle (Go)
```bash
cd hustle-turtle
go build -o bin/hustle-turtle ./cmd/server   # Build
go test ./...                                # Run all tests
go run ./cmd/server                          # Run server (port 8080)
go run ./cmd/server -auto-migrate            # Run with auto-migrations
go run ./cmd/server -migrate=up              # Apply migrations only
go run ./cmd/server -migrate=down            # Rollback migrations
go run ./cmd/server -version                 # Check migration version
```
Requires PostgreSQL. Default DATABASE_URL: `postgres://localhost/hustle_turtle?sslmode=disable`

### personal-site (Next.js)
```bash
cd personal-site
npm install        # Install dependencies
npm run dev        # Dev server
npm run build      # Production build
npm run lint       # ESLint (next lint)
```

### movoz-infra (Terraform)
```bash
cd movoz-infra
terraform init
terraform fmt -check -recursive   # Format check
terraform validate                # Validate config
terraform plan                    # Preview changes
```
Uses workspace-based environments (dev/prod). Region: us-west-2.

## Architecture Notes

### Project Independence
Each project is fully self-contained with its own build system, dependencies, and tests. There is no root-level package manager or monorepo orchestration tool.

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

### personal-site — Next.js App Router
- `src/app/` — App Router pages
- `src/components/` — React components (Hero, Projects, About, Contact, Navigation, Footer, ThemeToggle)
- Standalone output mode for Docker deployment
- Tailwind CSS with custom theme variables and dark mode

### movoz-infra — AWS VPC + RDS
- VPC with public/private subnets; PostgreSQL RDS in private subnets
- Dev: t3.micro/20GB, Prod: t3.small/100GB with enhanced monitoring
- `scripts/deploy.sh` for manual deployment

## CI/CD (GitHub Actions)

Workflows in `.github/workflows/` focus on infrastructure:
- **infrastructure-deploy.yml** — Terraform plan/apply on push to master (movoz-infra/ changes)
- **infrastructure-cost-analysis.yml** — Infracost analysis on PRs affecting infrastructure
- **infrastructure-cleanup.yml** — Weekly auto-destroy of dev resources (Sundays 2 AM UTC)
