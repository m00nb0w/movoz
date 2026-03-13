# Product Spec: Óctopus

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Project**: octopus
> **Location**: `apps/octopus/`
> **Stack**: Next.js

## Overview

Óctopus is a second brain — a personal knowledge management system for capturing thoughts, notes, learnings, and ideas, then connecting them into a web of knowledge that grows over time. Like an octopus with tentacles reaching into every domain of life, it links knowledge across topics.

## Problem Statement

Knowledge is scattered across Notion, bookmarks, random text files, Slack messages, and memory. Without a system to capture, connect, and resurface information, valuable insights are lost. Existing tools (Notion, Obsidian, Roam) are either too heavy, too opinionated, or don't integrate with a personal ecosystem.

## Goals

- Frictionless capture of notes, ideas, and learnings
- Bidirectional linking to connect knowledge across topics
- Powerful search and retrieval — find anything fast
- Visual knowledge graph to see how ideas connect
- Own the data — local-first, part of the movoz ecosystem

## Non-Goals

- Not a task manager (that's separate tooling)
- Not a document editor / Google Docs replacement
- Not collaborative — single-user system

## Features

### 1. Notes

| # | Requirement | Priority |
|---|-------------|----------|
| N1 | Create and edit notes in Markdown | Must have |
| N2 | Quick capture — minimal friction to jot something down | Must have |
| N3 | Tags for categorization | Must have |
| N4 | Full-text search across all notes | Must have |
| N5 | Daily notes / journal entries | Should have |

### 2. Connections

| # | Requirement | Priority |
|---|-------------|----------|
| C1 | Bidirectional links between notes (`[[note-title]]` syntax) | Must have |
| C2 | Backlinks panel — see all notes that link to current note | Must have |
| C3 | Visual knowledge graph (nodes = notes, edges = links) | Should have |
| C4 | Unlinked mentions — surface related notes automatically | Nice to have |

### 3. Knowledge Domains

Organize knowledge into tentacles (domains):

| Domain | Examples |
|--------|----------|
| Engineering | Architecture decisions, TILs, debugging notes, code patterns |
| Finance | Investment research, market notes (feeds into Drunken Dolphin) |
| Health | Workout learnings, nutrition notes (complements Hustle Turtle) |
| Travel | Trip planning notes, destination research (feeds into Travel Quokka) |
| Reading | Book notes, article highlights, summaries |
| Ideas | Project ideas, shower thoughts, things to explore |

### 4. Retrieval

| # | Requirement | Priority |
|---|-------------|----------|
| R1 | Full-text search with instant results | Must have |
| R2 | Filter by tag, domain, date range | Must have |
| R3 | Recent notes and recently edited | Must have |
| R4 | Random note resurfacing (spaced repetition / serendipity) | Nice to have |
| R5 | AI-powered semantic search | Nice to have |

## Pages

```
/octopus                — Dashboard (recent notes, quick capture, graph preview)
/octopus/notes          — All notes list with search and filters
/octopus/notes/:id      — Note editor/viewer
/octopus/graph          — Full knowledge graph visualization
/octopus/daily          — Daily notes / journal
/octopus/domains/:name  — Notes filtered by domain
```

## Data

Notes stored as Markdown files or in a database:

```
# Option A: File-based (like Obsidian)
content/
  engineering/
    nextjs-multi-zones.md
    go-migration-patterns.md
  reading/
    atomic-habits.md
  daily/
    2026-03-13.md

# Option B: Database-backed (via hustle-turtle or dedicated API)
notes table: id, title, content (markdown), domain, tags[], created_at, updated_at
links table: source_note_id, target_note_id
```

## Integration with Movoz

- Deployed as a Next.js multi-zone under `/octopus` path prefix
- Uses `@movoz/tailwind-config` and `@movoz/theme`
- Cross-references with other projects:
  - Finance notes feed context to **Drunken Dolphin**
  - Health/fitness notes complement **Hustle Turtle**
  - Travel notes connect to **Travel Quokka** trips

## Open Questions

- [ ] File-based (MDX/Markdown on disk) vs database-backed?
- [ ] Editor choice — custom Markdown editor, Tiptap, or MDX?
- [ ] Knowledge graph library — D3.js, Cytoscape, or Sigma.js?
- [ ] How to handle quick capture from mobile / outside the browser?
- [ ] AI features — semantic search, auto-linking, summarization?
