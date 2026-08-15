# Movoz — Project Overview

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13

## What is Movoz?

Movoz is a personal monorepo — a collection of small, independent projects built for learning, experimentation, and personal use.

## Projects

<!-- List your projects below. For each, describe what it is, why it exists, and its current status. -->

| Project | Location | Stack | Status | Description |
|---------|----------|-------|--------|-------------|
| Drunken Dolphin | `backend/drunken-dolphin/` | OpenClaw | Active | Personal assistant — finance management (expenses, investments) and research |
| Peaky Bergers | `backend/peaky-bergers/` | TBD | Planned | AI software team — agents that help build stuff |
| Hustle Turtle | `backend/hustle-turtle/` | Go | Active | Habit tracker — push ups, reading, running, and more |
| Lonely Moley | `infra/lonely-moley/` | K8s, Terraform, Ansible | Planned | Personal homelab — self-hosted cloud infrastructure |
| Travel Quokka | `apps/travel-quokka/` | Next.js | Planned | Travel map — countries visited, Vietnam districts, marathons, trip memories |
| Óctopus | `apps/octopus/` | Next.js | Planned | Second brain — personal knowledge management and note-taking |
| Oncarinho | `apps/oncarinho/`, `backend/oncarinho/` | Next.js + Go | Planned | Football team stats tracker — goals, assists, cards, leaderboards |
| Scout | `apps/scout/`, `backend/scout/` | Next.js + Go | Planned | Private engineering team tracker — FIFA-style attribute cards, GitHub/Jira metrics, AI-assisted ranking |

## Project Details

### Drunken Dolphin
- **Purpose**: Personal assistant for finance and research
- **Target users**: Personal use (To Ngoc Long)
- **Key features**: Expense tracking, investment management, research automation
- **Spec**: [wiki/specs/drunken-dolphin.md](./drunken-dolphin.md)

### Peaky Bergers
- **Purpose**: AI-powered software team to help build projects
- **Target users**: Personal use (To Ngoc Long)
- **Key features**: AI agents that assist with software development tasks
- **Spec**: [wiki/specs/peaky-bergers.md](./peaky-bergers.md)

### Hustle Turtle
- **Purpose**: Track daily habits and build consistency
- **Target users**: Personal use (To Ngoc Long)
- **Key features**: Log habits (push ups, reading, running), streaks, progress tracking
- **Spec**: [wiki/specs/hustle-turtle.md](./hustle-turtle.md)

### Lonely Moley
- **Purpose**: Self-hosted cloud infrastructure — own the stack end to end
- **Target users**: Personal use (To Ngoc Long)
- **Key features**: Kubernetes cluster, CI/CD pipelines, self-hosted services
- **Spec**: [wiki/specs/lonely-moley.md](./lonely-moley.md)

### Travel Quokka
- **Purpose**: Visual travel journal and map tracker
- **Target users**: Personal use + public showcase
- **Key features**: Interactive world/Vietnam maps, marathon tracker, trip photo galleries
- **Spec**: [wiki/specs/travel-quokka.md](./travel-quokka.md)

### Óctopus
- **Purpose**: Second brain — capture, connect, and retrieve personal knowledge
- **Target users**: Personal use (To Ngoc Long)
- **Key features**: Note-taking, knowledge graph, bidirectional linking, search
- **Spec**: [wiki/specs/octopus.md](./octopus.md)

### Oncarinho
- **Purpose**: Football team stats tracker
- **Target users**: Team admin (data entry) + team/public (viewing)
- **Key features**: Matchday stat entry (goals, assists, cards), year and all-time leaderboards, player profiles
- **Spec**: [wiki/specs/oncarinho.md](./oncarinho.md)

### Scout
- **Purpose**: Private tool for tracking engineering team performance
- **Target users**: Personal use (To Ngoc Long, as engineering manager)
- **Key features**: FIFA-style attribute cards (6 main attributes, growable sub-attributes) scored via biweekly forced ranking, GitHub/Jira metrics sync, AI-assisted conversational ranking
- **Spec**: [wiki/specs/scout.md](./scout.md)

## Shared Infrastructure

- **Monorepo tooling**: pnpm workspaces + Turborepo
- **Frontend packages**: Shared Tailwind config, theme system, TypeScript config
- **Infra**: Terraform (AWS), Docker Compose, Nginx reverse proxy
- **CI/CD**: GitHub Actions for infrastructure deploy, cost analysis, cleanup

## Principles

- Spec-driven development — spec before code
- Small, focused projects over monolithic apps
- Shared packages to avoid duplication across frontend apps
