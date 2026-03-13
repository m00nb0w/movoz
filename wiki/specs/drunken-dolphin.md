# Product Spec: Drunken Dolphin

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Project**: drunken-dolphin
> **Location**: `backend/drunken-dolphin/`
> **Stack**: Rust

## Overview

Drunken Dolphin is a personal assistant that handles two core domains: **finance management** and **research**. Built as a Rust CLI, it serves as a single tool for managing personal expenses, tracking investments, and conducting research tasks.

## Problem Statement

Managing personal finances across multiple tools (bank apps, spreadsheets, investment platforms) is fragmented and time-consuming. Research tasks (comparing options, gathering data, summarizing findings) are ad-hoc and scattered across browser tabs. A unified personal assistant consolidates these workflows into one tool.

## Goals

- Single CLI tool for personal finance and research tasks
- Track daily expenses and categorize spending
- Monitor investments and portfolio performance
- Automate research workflows (data gathering, summarization)
- Keep data local and private

## Non-Goals

- Not a replacement for banking apps (no transactions, no payments)
- Not a team/shared tool — personal use only
- Not a GUI application (CLI-first)

## Domains

### 1. Finance — Personal Expenses

Track and categorize daily spending.

| # | Requirement | Priority |
|---|-------------|----------|
| F1 | Add expense entries (amount, category, description, date) | Must have |
| F2 | Categorize expenses (food, transport, housing, etc.) | Must have |
| F3 | View spending summaries (daily, weekly, monthly) | Must have |
| F4 | Set and track budgets per category | Should have |
| F5 | Import expenses from CSV/bank exports | Nice to have |

### 2. Finance — Investments

Track portfolio and investment performance.

| # | Requirement | Priority |
|---|-------------|----------|
| I1 | Record investment positions (asset, quantity, cost basis) | Must have |
| I2 | Track portfolio value over time | Must have |
| I3 | View gains/losses per position and overall | Must have |
| I4 | Fetch live market data for price updates | Should have |
| I5 | Compare allocation against target percentages | Nice to have |

### 3. Research

Automate and organize personal research tasks.

| # | Requirement | Priority |
|---|-------------|----------|
| R1 | Define and run research tasks | Must have |
| R2 | Store and retrieve research findings | Must have |
| R3 | Summarize research results | Should have |
| R4 | Compare options side-by-side | Should have |
| R5 | Integration with web sources for data gathering | Nice to have |

## Current State

Drunken Dolphin currently exists as a fitness tracking CLI (pushups, situps, pullups). This spec redefines its scope to become a broader personal assistant focused on finance and research.

## Data Storage

- Local-first — data persists to local files (JSON or SQLite)
- No cloud dependency for core functionality
- Export/backup capabilities

## Open Questions

- [ ] Should fitness tracking remain as a sub-command, or be removed entirely?
- [ ] SQLite vs JSON files for structured finance data?
- [ ] What research sources to integrate with? (web scraping, APIs, LLM-powered?)
- [ ] CLI-only or add a TUI (terminal UI) for dashboards?
