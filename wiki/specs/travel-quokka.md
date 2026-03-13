# Product Spec: Travel Quokka

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Project**: travel-quokka
> **Location**: `apps/travel-quokka/`
> **Stack**: Next.js

## Overview

Travel Quokka is an interactive travel map and journal. It visualizes places visited at multiple levels — countries around the world, provinces and districts in Vietnam — and doubles as a marathon map tracker. Each trip can have photos, notes, and memories attached.

## Problem Statement

Travel memories are scattered across phone galleries, Google Maps timelines, and social media posts. There's no single place to see everywhere you've been, relive trip memories, and track marathon races on a map. A personal travel site brings it all together in a visual, shareable format.

## Goals

- Interactive maps showing visited places at multiple granularities
- Trip journal with photos and memories
- Marathon/race tracker on a map
- Public-facing — shareable with friends and family
- Part of the movoz multi-zone frontend (served under a path prefix)

## Non-Goals

- Not a trip planning tool
- Not a social network or review platform
- Not real-time location tracking

## Features

### 1. World Map

| # | Requirement | Priority |
|---|-------------|----------|
| W1 | Interactive world map with visited countries highlighted | Must have |
| W2 | Click country to see trip details | Must have |
| W3 | Country count / stats (X of 195 visited) | Must have |
| W4 | Color coding by visit type (lived, traveled, transited) | Should have |

### 2. Vietnam Map

| # | Requirement | Priority |
|---|-------------|----------|
| V1 | Vietnam map with provinces highlighted | Must have |
| V2 | Drill down to district level within provinces | Must have |
| V3 | District count / stats per province | Should have |
| V4 | Filter by year or trip | Nice to have |

### 3. Marathon / Race Tracker

| # | Requirement | Priority |
|---|-------------|----------|
| M1 | Map pins for each marathon/race location | Must have |
| M2 | Race details (name, date, distance, finish time) | Must have |
| M3 | Timeline view of all races | Should have |
| M4 | Stats (total distance, average pace, PBs) | Should have |
| M5 | Route/GPX overlay on map | Nice to have |

### 4. Trip Journal

| # | Requirement | Priority |
|---|-------------|----------|
| T1 | Create trip entries with dates, location, description | Must have |
| T2 | Attach photos to trips | Must have |
| T3 | Photo gallery per trip | Must have |
| T4 | Link trips to map locations (country/province/district) | Must have |
| T5 | Timeline/feed view of all trips | Should have |
| T6 | Tags and filtering (solo, family, marathon, etc.) | Nice to have |

## Pages

```
/travel-quokka              — Landing page with world map overview
/travel-quokka/vietnam      — Vietnam map (provinces → districts)
/travel-quokka/marathons    — Marathon map and race list
/travel-quokka/trips        — Trip journal timeline
/travel-quokka/trips/:id    — Individual trip with photos and memories
```

## Data

Trip and location data can be stored as static JSON/MDX files (no backend needed for v1):

```
trips/
  2024-japan/
    index.mdx          — Trip description, dates, metadata
    photos/             — Trip images
  2025-da-nang-marathon/
    index.mdx
    photos/

data/
  countries.json        — List of visited countries with metadata
  vietnam-provinces.json — Visited provinces and districts
  marathons.json        — Race records
```

## Map Technology

- **World map**: React Simple Maps or Mapbox GL for country-level choropleth
- **Vietnam map**: GeoJSON boundaries for provinces/districts (available from government open data)
- **Marathon pins**: Mapbox or Leaflet for point markers

## Integration with Movoz

- Deployed as a Next.js multi-zone under `/travel-quokka` path prefix
- Uses `@movoz/tailwind-config` and `@movoz/theme` for consistent styling
- Nginx route added in `infra/nginx/nginx.conf`

## Open Questions

- [ ] Static data (MDX/JSON) vs API-backed (hustle-turtle or dedicated backend)?
- [ ] Image hosting — local assets, S3, or Cloudinary?
- [ ] Vietnam district-level GeoJSON source?
- [ ] Mapbox (paid beyond free tier) vs Leaflet (free, less polished)?
- [ ] Should marathon tracking live here or in hustle-turtle as a habit?
