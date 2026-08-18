# Product Spec: Drunken Dolphin

> **Status**: Draft (v0.2 — requirements & user stories only)
> **Owner**: Long (single end user)
> **Author**: Drunken Dolphin (Long's personal AI assistant)
> **Date**: 2026-05-03
> **Project**: drunken-dolphin
> **Location**: `backend/drunken-dolphin/` (assistant) + new `apps/` zone (dashboard, name TBD)
> **Stack**: OpenClaw (agent runtime); TBD (dashboard)

---

## 1. Vision

A **private, single-user personal system** that gives Long a single morning glance at his life:

- Today's curated newsletter (auth, AI, tech, finance, markets, Vietnam)
- Personal financial state (today's expenses, monthly bucket breakdown, budget gauge)
- Future modules: tasks, habits, calendar, journal, reading list, etc.

Drunken Dolphin is **one product with two surfaces**:

1. **The assistant** — a background workflow that collects, parses, and stores Long's data (RSS feeds, expense messages from Telegram, scheduled summaries). Hosted on OpenClaw, running on Long's machine.
2. **The dashboard** — a private web app (and eventually mobile) that *views* and *lightly edits* what the assistant has collected. Same data, daily ritual surface.

Long's data is collected by the assistant in the background. The dashboard is the **viewer + light editor**. Telegram remains a primary communication channel for the assistant.

---

## 2. Goals

- **Single source of truth** for everything Long tracks about himself.
- **Low-friction daily ritual** — open the app, see the day in one glance.
- **Long lives in his data**: query, filter, search his own life.
- **Privacy-first** — single user, private auth, no third-party analytics, data stays under Long's control.
- **Composable** — modules (newsletter, expenses, future habits/tasks) are independent; new modules can be added without rewriting the system.
- **Automate Long's workflows** — expense logging, research, summarization — so he doesn't have to think about admin.

## 3. Non-goals (v1)

- Multi-user / sharing / collaboration features
- Social features
- Any public-facing pages
- Replacement for banking apps (no transactions, no payments)
- Native iOS / Android (a PWA is acceptable for v1; native is a v2 question)
- Offline-first sync (network required is fine)
- In-app push notifications (Telegram already handles notification delivery)

---

## 4. Personas

**Long** is the only user. Software engineer in Vietnam, works on auth/identity systems (OAuth/OIDC). Comfortable with code; wants both a polished daily-use product *and* a project he enjoys building.

---

## 5. System architecture (high level)

| Surface | Role | Stack |
|---|---|---|
| **Assistant** (`backend/drunken-dolphin/`) | Background data collector + parser + scheduled jobs. Reads RSS, parses Telegram messages, runs digests, writes to central store. | OpenClaw |
| **Dashboard** (new app, name TBD) | Web app + mobile view of the same data. Light editing (add/edit/delete expenses, manage sources, set budget). | TBD |
| **Telegram bot** | Existing communication channel — plain-language expense messages in, summaries out. The dashboard does **not** replace it. | Existing |
| **Central data store** | One database both surfaces read/write through (the assistant should write through the same APIs the frontend uses, where practical). | TBD |

Authentication separates *Long* (interactive sessions) from *the assistant* (machine credential).

---

## 6. User stories

Each story is tagged with a priority:
- **P0** — must-have for v1 launch (without these, the product doesn't deliver the core value)
- **P1** — important, can ship in v1.1 or later

**Summary:** 12 P0 stories, 10 P1 stories.

### 6.1 Daily ritual

**US-1 — Morning glance** · **P0**
*As Long, I want to open the app each morning and see, in under 5 seconds, today's news digest, yesterday's spending, this month's running total vs budget, and any urgent flags — so I start the day informed without digging through Telegram.*

**US-2 — One-tap drill-in** · **P0**
*As Long, when a digest item catches my eye, I want to open the full original article in one tap and easily come back to where I was — so reading doesn't lose my place in the dashboard.*

**US-3 — Greeting & date context** · **P1**
*As Long, I want the dashboard to greet me with my name, the day, and any meaningful contextual nudges (over budget? unread digest? a saved item I haven't opened yet?) — so the first impression is personal.*

### 6.2 Newsletter

**US-4 — Today's digest** · **P0**
*As Long, I want a clean view of today's curated digest with items grouped by topic (Auth, AI, Tech, Finance, Markets, Vietnam) — so I can scan and prioritize.*

**US-5 — Per-item feedback** · **P0**
*As Long, I want to react to each digest item with "loved", "boring", or "kill this source" — so the curator improves over time without me writing prose feedback.*

**US-6 — Save for later** · **P1**
*As Long, I want to bookmark digest items I don't have time to read now — so I can come back to them later from a "saved" tab.*

**US-7 — Past digests** · **P0**
*As Long, I want to browse past daily digests by date — so I can revisit what I read last week.*

**US-8 — Search across digests** · **P1**
*As Long, I want to search the full text of all digests I've ever received — so I can find that thing I read last month about OAuth flows.*

**US-9 — Source management** · **P1**
*As Long, I want to see all my newsletter sources, adjust their weights, disable some, and add new ones by pasting a URL — so I'm in control of what gets curated.*

**US-10 — Add source by URL** · **P1**
*As Long, when I paste a website URL, I want the system to auto-discover its RSS feed and add it to my sources — so I don't have to hunt for feed URLs.*

### 6.3 Expenses

**US-11 — Today's expenses at a glance** · **P0**
*As Long, I want to see today's expenses (count, total, individual entries) on the dashboard — so I have an instant pulse on my day.*

**US-12 — Monthly summary with charts** · **P0**
*As Long, I want to see this month's spending broken down by bucket (Essentials / Lifestyle / Irregular) and by category, in interactive charts — so I understand where my money is going.*

**US-13 — Budget gauge** · **P0**
*As Long, I want a visible gauge of "month-to-date spending vs my monthly budget" with clear over/under indicators — so I know when to slow down.*

**US-14 — Filterable expense list** · **P0**
*As Long, I want to filter and sort my full expense history by date range, category, tag, and bucket — so I can answer specific questions about my spending.*

**US-15 — Tag rollup** · **P1**
*As Long, I want to see all expenses for a given hashtag (like `#ustravel2026`) with a running total — so I can track trip- or project-specific costs.*

**US-16 — Add expense from the app** · **P0**
*As Long, I want to add expenses from a form in the app — and I want it to also accept plain language ("lunch 120k", "flight 8M #ustravel2026") — so I'm not forced to use Telegram for every entry.*

**US-17 — Edit / delete expenses** · **P0**
*As Long, I want to fix or remove an expense entry from the app with confirmation — so I can correct mistakes without leaving the dashboard.*

**US-18 — Adjust monthly budget** · **P1**
*As Long, I want to change my monthly budget value from settings — so I can recalibrate as my life changes.*

### 6.4 Cross-cutting

**US-19 — Cross-module search** · **P1**
*As Long, I want a single search box that searches both digests and expenses — so I can find anything in my data without picking a category first.*

**US-20 — Agent status** · **P1**
*As Long, I want to see the assistant's last activity (last digest run, last expense ingested, any errors) — so I trust the system is working.*

**US-21 — Authentication** · **P0**
*As Long, I want the dashboard to be unreachable from the public internet — only my devices on my private VPN (Tailscale) can connect. Once a device is on the VPN, a one-time "approve this device" setup grants it a long-lived trusted-device cookie so the dashboard opens instantly forever after, no login screen ever.*

**US-22 — Settings & preferences** · **P1**
*As Long, I want to manage my profile (name, timezone), budgets, newsletter topic preferences and anti-list, and authentication settings — all in one place. (Auth-related settings like "sign out / revoke devices" are P0 by virtue of US-21.)*

### 6.5 Future modules (out of scope for v1, but the product should accommodate)

- Tasks / todos
- Habits / streaks
- Calendar / events
- Journal entries
- Health metrics (sleep, weight, steps from a wearable)
- Reading list / book log
- Watch list (movies / TV)
- Ask-the-assistant chat inside the app
- Investments (record positions, track value, gains/losses, allocation vs target)
- Research (define and run research tasks, store findings, summarize, compare options)

---

## 7. Functional requirements

### 7.1 Authentication
- Single user (Long). No public signup. No external identity provider.
- The dashboard is **only reachable on Long's private network** (Tailscale tailnet). It does not bind to a public interface and has no public DNS record.
- On a new device, a **one-time setup link** (URL containing a setup secret) registers the device by setting a long-lived signed cookie. Subsequent visits from that device pass automatically. Daily use never sees a login screen.
- All views and actions require either the trusted-device cookie (browser) or an agent bearer token (machine).
- The data-collector agent (the assistant) writes via a separate, distinguishable credential — a long-lived bearer token — distinct from the user's trusted-device cookie. When the agent and the dashboard run on the same host, the agent uses localhost; otherwise it presents the bearer token over the tailnet.
- **Revocation**: rotating one env var (`DEVICE_SECRET`) immediately invalidates every trusted device; revoking a row in `api_tokens` invalidates the agent.

### 7.2 Data freshness
- When the agent writes new data, it should appear in the dashboard within minutes (no manual refresh required for the daily-use surfaces).
- The daily digest must reliably appear by 7am Asia/Saigon time on >95% of days.

### 7.3 Performance targets
- Dashboard first contentful paint: < 1.5s on a typical mobile connection.
- Page navigation between core views: < 800ms p95.
- Search results: < 1s p95 for the most recent month of data.

### 7.4 Reliability
- Operations that mutate data (edit/delete an expense, adjust a budget, change a source weight) must require confirmation and be reversible where possible.
- Data must not be lost on transient failures; writes are idempotent.
- The system must back up the database at least daily.

### 7.5 Privacy
- All user data stays under Long's control.
- No third-party analytics that observe user data.
- No PII or content sent to external services beyond what's already in use (the LLM provider used for digest summarization, the RSS feeds being read, and the auth provider).

### 7.6 Cost
- Total monthly running cost (hosting + DB + LLM usage) for v1 should stay under $25.

### 7.7 Internationalization
- UI is in English.
- Source data may be in Vietnamese (e.g. expense notes, Vietnamese news sources). The agent translates Vietnamese newsletter content to English in the digest summary; the original Vietnamese text remains available via the source link.

### 7.8 Accessibility
- Semantic HTML, keyboard navigation, sufficient color contrast.
- Not the primary focus, but the app shouldn't be unusable on a screen reader.

### 7.9 Maintainability
- Clear separation between the data-collector agent and the user-facing app.
- The agent should write through the same APIs the frontend uses, where practical.
- New modules (tasks, habits, etc.) should be addable without modifying the existing modules.

---

## 8. Assistant: detailed behavior

This section pins down the assistant's interactive behavior. Both Telegram and the dashboard's "add expense from the app" flow (US-16) share these rules.

### 8.1 Expense parsing

**Input**: a natural-language message describing an expense. May include:
- **Description** (required) — what the expense is for
- **Amount** (required) — supports shorthand: `50k` = 50,000 · `1.5k` = 1,500 · `1.5m` = 1,500,000 · plain numbers (`50000` = 50,000) · bare small numbers taken literally (`50` = 50)
- **Category** (optional) — if not provided, the assistant auto-categorizes
- **Date** (optional) — defaults to today if not specified
- **Trip tag** (optional) — a hashtag like `#ustravel2026` recognized as a trip

**Buckets vs categories**: expenses roll up into three high-level **buckets** for the dashboard's summary view — *Essentials*, *Lifestyle*, *Irregular*. Each **category** below maps to one bucket. The mapping lives in the assistant.

| Category | Examples |
|----------|----------|
| Health | Doctor visit, medicine, health checkup |
| Food | Lunch, coffee, groceries, dinner |
| Entertainment | Movies, games, concerts, Netflix |
| Transportation | Grab, taxi, bus, fuel, parking |
| Sports | Running shoes, gym, trekking gear, marathon entry, trail run |
| Utilities | Electricity, water, internet, phone bill |
| Personal Projects | Domain names, hosting, hardware, tools |
| Other | Anything that doesn't fit above categories |

**Categorization rules**:
- Sports-related activities (trekking, running, marathon, cycling, gym, sports equipment) are **always Sports**, never Entertainment.
- If the user specifies a category explicitly, use it as-is — no override.
- If no category can be determined, assign **Other**.

**Examples**:

| Message | Description | Amount | Category | Date |
|---------|-------------|--------|----------|------|
| `grab to office 50k` | Grab to office | 50,000 | Transportation | Today |
| `lunch with team 200k food` | Lunch with team | 200,000 | Food | Today |
| `gym membership 500k yesterday` | Gym membership | 500,000 | Sports | Yesterday |
| `trail running shoes 2m` | Trail running shoes | 2,000,000 | Sports | Today |
| `netflix subscription 70k` | Netflix subscription | 70,000 | Entertainment | Today |
| `coffee 45k 2026-03-10` | Coffee | 45,000 | Food | 2026-03-10 |
| `gym 100k entertainment` | Gym | 100,000 | Entertainment | Today |
| `birthday gift for mom 500k` | Birthday gift for mom | 500,000 | Other | Today |
| `flight 8M #ustravel2026` | Flight | 8,000,000 | Transportation (tag: ustravel2026) | Today |

**Output**: the assistant confirms back with the parsed expense (description, formatted amount, category, date, any tags). If the message is missing an amount or is unparseable, the assistant asks for clarification instead of recording.

**Currency**: VND (default, single currency for now).

**Acceptance criteria**:
- Expense is persisted after confirmation.
- Amount shorthand (k, m) is parsed correctly.
- Auto-categorization follows the rules above (Sports priority over Entertainment).
- User-specified category is respected without override.
- Date defaults to today when not provided.
- Confirmation message is sent back with all parsed fields.
- Missing amount or unparseable message triggers a clarification request.
- Unrecognized expenses are categorized as Other.
- Trip tags (hashtags) are extracted and attached.

### 8.2 Newsletter digest

The assistant runs a daily job that:
- Pulls items from configured RSS sources (24+ across Auth, AI, Tech, Finance, Markets, Vietnam).
- Ranks/filters items per source weights and Long's anti-list.
- Summarizes via an LLM provider; translates Vietnamese content to English.
- Persists the digest, then delivers it via Telegram (existing) and surfaces it in the dashboard.
- Reliability target: 7am Asia/Saigon, ≥95% of days (see §7.2).

Long's per-item feedback (US-5) feeds back into source weighting and the curator's prompt context.

### 8.3 Telegram channel

Telegram is the existing primary I/O channel for the assistant. Plain-language expense messages in; daily/weekly summaries and check-ins out. Telegram remains supported alongside the dashboard — see Appendix A for the live schedule.

### 8.4 Future agent capabilities (not in v1)

- **Investments** — record positions (asset, quantity, cost basis), track portfolio value over time, view gains/losses, fetch live market data, compare allocation vs target.
- **Research** — define and run research tasks, store and retrieve findings, summarize results, compare options side-by-side, integrate with web sources.
- **In-app chat** — "ask Drunken Dolphin" surface inside the dashboard (today, only Telegram).

---

## 9. UX & design notes (non-binding, vibes)

- **Tone:** quiet, calm, considered. This is the assistant Long checks in the morning, not a hype dashboard. Lots of whitespace.
- **Color palette:** carry over from the existing chart palette already used in Telegram — green for Essentials, orange for Lifestyle, purple for Irregular. A primary accent color for the brand is TBD.
- **Typography:** clean sans-serif (Inter or similar). Generous line height. Body copy 15–16px.
- **Layout:** vertical stack of cards on mobile; side-by-side cards on desktop.
- **No cookie banners or consent flows** — single-user app, no third parties watching.
- **One small Easter egg:** a 🐬 dolphin appears somewhere subtle (loading spinner, 404 page).

---

## 10. Open questions (decide before tech design starts)

1. ~~**Auth provider preference?**~~ — **Resolved 2026-05-10**: No external IdP. **Tailscale-only network access** + a **trusted-device cookie** signed with a local secret. Google sign-in was considered (2026-05-08) but rejected after the user requested zero third-party auth dependency. This is fully cloud-agnostic — the app runs anywhere reachable on Long's tailnet, no managed service in the path.
2. **Hosting target?** (Vercel / Fly.io / self-hosted / something else)
3. **Dashboard product/zone name?** ("Drunken Dolphin" covers the system; the dashboard zone may want its own short name)
4. **Domain?** (Subdomain Long owns, or buy something new)
5. **Repo location & visibility?** (GitHub public / private / self-hosted)
6. **Mobile priority for v1?** (PWA-only, or push for native sooner)
7. **In-app chat with the assistant?** (Currently Telegram-only. Should the dashboard also have a "ask Drunken Dolphin" chat surface in v1, or defer?)
8. **Central data store?** (SQLite vs Postgres; same DB for assistant + dashboard, or assistant's local store + dashboard reads via API?)
9. **Multi-currency support?** (VND only today; deferred)

---

## 11. Success metrics (evaluated 4 weeks after launch)

- Long opens the app at least **5 days/week** without prompting.
- He files at least **one piece of newsletter feedback per week** (loved / boring / kill / save).
- He logs at least one expense **from inside the app** (signal that the app is nicer than Telegram for some flows).
- The 7am digest succeeds on **27 of 28 days**.
- Long can articulate one thing the app does **better** than just receiving Telegram messages.

---

## 12. Glossary

- **Bucket** — high-level grouping of expense categories: Essentials, Lifestyle, Irregular.
- **Category** — finer-grained expense label (Food, Sports, Transportation, …) that maps to a bucket.
- **Digest** — the daily curated newsletter delivered both to Telegram and the app.
- **News item** — one article from one source.
- **Source** — a feed Long has subscribed to (currently RSS; future: email, X, etc.).
- **Trip tag** — a hashtag like `#ustravel2026` that the existing expense system recognizes as a trip and categorizes accordingly.
- **VND** — Vietnamese Dong, the default currency.

---

## Appendix A — Existing infrastructure already running (context only — not requirements)

The following is what the assistant already operates on Long's behalf today. It exists; the dashboard should integrate with it, not replace it.

| Thing | Status |
|---|---|
| Daily expense logging via Telegram (plain language) | live |
| Daily 8am expense summary with chart on Telegram | live |
| Daily 8pm expense check-in | live |
| Daily 10pm follow-up nudge (if no reply) | live |
| Daily 7am newsletter digest on Telegram | live |
| Weekly Sunday 7pm newsletter feedback prompt | live |
| 24+ RSS sources across Auth, AI, Tech, Finance, Markets, Vietnam | live |
| Plain-language expense parser (`lunch 120k`, `flight 8M #ustravel2026`) | live |
| 156+ historical expense rows | already in storage |
| 6+ months of digest archives planned | starting now |

---

## Appendix B — Migration notes

The Drunken Dolphin codebase originated as a fitness tracking CLI (pushups, situps, pullups). Fitness tracking has migrated to Hustle Turtle. This spec defines the current and forward scope: a single personal-assistant product with an agent backend and a dashboard frontend.
