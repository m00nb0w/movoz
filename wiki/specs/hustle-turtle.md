# Product Spec: Hustle Turtle

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Project**: hustle-turtle
> **Location**: `backend/hustle-turtle/`
> **Stack**: Go (Gin + PostgreSQL)

## Overview

Hustle Turtle is a habit tracking service. Log daily habits like push ups, reading, and running — then track streaks, progress, and consistency over time.

## Problem Statement

Building good habits requires consistent tracking and visibility into progress. Generic habit apps are bloated with features and lack the simplicity of tracking a few core habits well. A focused, personal habit tracker that integrates into the movoz ecosystem is more useful than a third-party app.

## Goals

- Track daily habits with minimal friction
- Support different habit types (count-based, duration-based, boolean)
- Visualize streaks and progress over time
- REST API backend that can serve a frontend or CLI client

## Non-Goals

- Not a social/community habit platform
- Not a gamification engine — keep it simple
- Not a notification/reminder system (for now)

## Habit Types

| Type | Example | How Logged |
|------|---------|------------|
| Count | Push ups, pull ups, sit ups | Number per session |
| Duration | Reading, meditation | Minutes per session |
| Distance | Running, walking, cycling | Distance per session (km) |
| Boolean | Cold shower, journaling | Done / not done |

## Requirements

### Core

| # | Requirement | Priority |
|---|-------------|----------|
| H1 | Create and manage habits (name, type, unit, target) | Must have |
| H2 | Log habit entries (value, date, optional notes) | Must have |
| H3 | View habit history (daily, weekly, monthly) | Must have |
| H4 | Track current and longest streaks per habit | Must have |
| H5 | REST API for all operations | Must have |

### Analytics

| # | Requirement | Priority |
|---|-------------|----------|
| A1 | Weekly/monthly summaries per habit | Should have |
| A2 | Progress toward daily/weekly targets | Should have |
| A3 | Trends over time (improving, declining, stable) | Nice to have |
| A4 | Export data as CSV/JSON | Nice to have |

## API Design

```
POST   /habits              — Create a habit
GET    /habits              — List all habits
GET    /habits/:id          — Get habit details + stats
PUT    /habits/:id          — Update habit config
DELETE /habits/:id          — Delete habit

POST   /habits/:id/entries  — Log an entry
GET    /habits/:id/entries  — List entries (with date filters)

GET    /habits/:id/streaks  — Get streak info
GET    /habits/:id/summary  — Get weekly/monthly summary
```

## Current State

Hustle Turtle currently exists as a Go REST API with Gin and PostgreSQL, with migrations infrastructure in place. The existing fitness-related code in drunken-dolphin (push ups, sit ups, pull ups) should migrate here as habits, since drunken-dolphin is being repurposed for finance and research.

## Data Model

```
habits
  id, name, type (count|duration|distance|boolean), unit, daily_target, created_at

habit_entries
  id, habit_id, value, date, notes, created_at
```

## Open Questions

- [ ] Migrate existing fitness data from drunken-dolphin?
- [ ] Should there be a frontend dashboard in apps/, or CLI-only for now?
- [ ] Weekly targets in addition to daily targets?
- [ ] Support for habit pausing (vacation mode)?
