# Feature Specification: el-storko — Personal Work-Item Tracker

**Feature Branch**: `001-el-storko`

**Created**: 2026-09-21

**Status**: Draft

**Input**: User description: "el-storko — personal work-item tracker (Jira-like), unifying personal tasks/epics with bidirectionally-synced Jira work items, plus a placeholder source for a future AI agent team." Revised 2026-09-22 around a daily-planning workflow: a five-state status model built around a "pick for today" ritual, a simplified Board+Stats-only navigation, per-card Jira distinction, an internal reference key, and a full item detail drawer. Revised again 2026-09-22 with a global Mine/Agent scope toggle, a searchable Epic picker in the drawer, and a globally unique reference key.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Capture and triage work into a Backlog (Priority: P1)

As the single user, I create, edit, and delete Tasks and Epics (title, description, status), and nest Tasks under an Epic, so personal projects and chores land in one durable Backlog instead of nowhere.

**Why this priority**: This is the core value of the tracker and works with zero external dependencies. Without it there is nothing to build on.

**Independent Test**: Create an Epic, create two Tasks under it, edit one Task's status and description, delete the other Task, delete the Epic. Confirm state persists across a service restart.

**Acceptance Scenarios**:

1. **Given** no existing items, **When** I create an Epic with a title, **Then** it appears in the tracker with status `Backlog` and no parent.
2. **Given** an existing Epic, **When** I create a Task with that Epic as parent, **Then** the Task appears nested under the Epic.
3. **Given** a Task with a parent Epic, **When** I try to set another Epic as its parent, **Then** the Task's parent is simply reassigned (a Task has at most one Epic parent at a time).
4. **Given** an existing Task or Epic, **When** I edit its title, description, or status, **Then** the change is saved and visible immediately.
5. **Given** an existing Task or Epic, **When** I delete it, **Then** it no longer appears anywhere.
6. **Given** the service has been restarted, **When** I view the tracker, **Then** all previously created items are still present.

---

### User Story 2 - Plan today's work (Priority: P1)

As the single user, each morning I pick roughly three items out of my Backlog into "Picked for today," and the Board shows only my active work (Picked for today, In progress, Blocked, Done) with the full Backlog listed directly beneath it, so my board reflects what I'm actually doing today instead of everything I might ever do.

**Why this priority**: This daily-planning ritual is the central workflow the whole redesign is built around — it's how the user decides what to work on, not just where finished work gets recorded.

**Independent Test**: Put several items in Backlog, use "Pick for today" on a few, confirm the Board shows exactly the active-status columns with those items and the Backlog list below still shows the rest.

**Acceptance Scenarios**:

1. **Given** an item in Backlog, **When** I use its "Pick for today" action, **Then** its status becomes `Picked for today` and it appears on the Board.
2. **Given** an item in the Backlog list beneath the Board, **When** I drag it onto the Board, **Then** it moves to `Picked for today` — the same effect as the button.
3. **Given** the Board view, **When** I open it, **Then** I see exactly four columns — Picked for today, In progress, Blocked, Done — and Backlog items are never rendered as a Board column.
4. **Given** items sitting in Backlog, **When** I view the Board page, **Then** I see the full Backlog listed directly below the board, not on a separate tab.
5. **Given** I click the app logo from any page, **When** the page loads, **Then** I land on the Board view.

---

### User Story 3 - See Jira work automatically, without leaving the tracker (Priority: P2)

As the single user, my Jira-assigned issues automatically appear in the tracker — visually marked as Jira's own on their card — and a local status or description edit on a Jira-sourced item syncs back to Jira, so I have one place to check and update instead of two.

**Why this priority**: This is the feature that eliminates the "scattered across two tools" problem, but it depends on User Story 1's data model and User Story 2's status model existing first.

**Independent Test**: With a Jira issue assigned to the user, wait one polling interval and confirm it appears locally, tinted as Jira-sourced, showing its real Jira key. Change its status locally and confirm the change appears on the Jira issue within one polling interval. Change the same issue in Jira and confirm the local copy picks up the change on the next poll.

**Acceptance Scenarios**:

1. **Given** a Jira issue assigned to the user that isn't yet in the tracker, **When** a sync cycle runs, **Then** the issue appears locally with `source = jira`, mapped into Backlog, In progress, Blocked, or Done based on its Jira status — never into Picked for today, since that state is a personal daily-triage decision with no Jira equivalent.
2. **Given** a Jira-sourced item in the tracker, **When** I change its status or description locally, **Then** the same change appears on the Jira issue within one polling interval.
3. **Given** a Jira issue's status or description changes directly in Jira, **When** the next sync cycle runs, **Then** the local item reflects that change.
4. **Given** a Jira-sourced item and a conflicting edit made in both places between polls, **When** the sync cycle runs, **Then** the more recently updated side wins and the other is overwritten.
5. **Given** the Jira sync fails (e.g. network error, invalid token), **When** the failure happens, **Then** personal (non-Jira) items remain fully usable and the failure doesn't crash the service.
6. **Given** a Jira-sourced item, **When** I view it on the Board or in the Backlog list, **Then** its card has a distinct tint (background/border) and its badge shows the real Jira key (e.g. `AUTH-142`) instead of a generic "Jira" label.

---

### User Story 4 - Capture from the CLI (Priority: P2)

As the single user, I can add and list tasks from the CLI for quick capture, so I can check or update things without opening a browser.

**Why this priority**: This is a convenience layer on top of User Story 1's data. It depends on that data existing, and benefits from — but does not strictly require — User Stories 2 or 3.

**Independent Test**: From a terminal, add a task and list tasks, confirming the new task appears — then open the web UI and confirm the same item shows up there too.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** I run the CLI add command with a title, **Then** a new personal Task is created in Backlog and is immediately visible in the web UI.
2. **Given** existing items, **When** I run the CLI list command, **Then** it prints the same items visible in the web UI, filterable by the same source/status values the API supports.

---

### User Story 5 - Review full item details in a drawer (Priority: P2)

As the single user, I open any item's detail drawer to see and edit every field — including an Estimate in hours and a Due date — plus its internal reference key and, if it's Jira-sourced, its Jira key, all sized to fit without scrolling, so a quick check-in never turns into hunting through a cramped panel.

**Why this priority**: This is where the richer fields (estimate, due date) live, and where the internal reference key becomes visible for future lookups — valuable, but the tracker is still useful without it via quick card-level actions.

**Independent Test**: Open an item's drawer, set an Estimate and a Due date, close and reopen it, confirm both persisted and every field was visible without scrolling the drawer itself.

**Acceptance Scenarios**:

1. **Given** any item, **When** I open its drawer, **Then** I see its title, description, status, Estimate, Due date, internal reference key (`MOVOZ-#`), and — if Jira-sourced — its Jira key, all visible without the drawer needing to scroll.
2. **Given** a Due date in the past, **When** I view the item (on its card or in the drawer), **Then** the Due date is shown in red.
3. **Given** a Due date within the next 3 days, **When** I view the item, **Then** the Due date is shown in yellow.
4. **Given** a Due date more than 3 days away, or no Due date at all, **When** I view the item, **Then** no red/yellow color-coding is applied.
5. **Given** I set an Estimate and save, **When** I reopen the drawer later, **Then** the Estimate is still there.
6. **Given** a Task in the drawer, **When** I type into the Epic field, **Then** the list of candidate Epics filters to matching results as I type, and clicking one assigns it as the parent.
7. **Given** a Task already assigned to an Epic, **When** I use the Epic field's "Clear" action, **Then** the Task becomes unparented.

---

### User Story 6 - See workload and burn-rate stats (Priority: P3)

As the single user, I have a Stats tab showing my open-item count, how much I completed this week, my overall completion rate, my total tracked items, and a breakdown of items by status, so I can gauge my workload and progress at a glance — with room to add more metrics to the same tab later.

**Why this priority**: This is an insight/reporting layer on top of data the other user stories already produce. It adds no new capability to create, plan, or sync work, so it's valuable but not required for the tracker to be useful day to day.

**Independent Test**: Seed items across all five statuses with a mix of completion dates, open the Stats tab, and confirm every number and the status breakdown match a hand count.

**Acceptance Scenarios**:

1. **Given** items in various statuses, **When** I open the Stats tab, **Then** I see the current count of open (not-`Done`) items.
2. **Given** items completed within the last 7 days, **When** I open the Stats tab, **Then** I see how many were completed in that window.
3. **Given** all tracked items, **When** I open the Stats tab, **Then** I see an overall completion rate (share of all tracked items that are `Done`) and the total tracked item count.
4. **Given** items spread across the five statuses, **When** I open the Stats tab, **Then** I see a bar chart breaking down item counts by status.
5. **Given** the Stats tab exists today with these metrics, **When** a future metric is added, **Then** it can be added as another card without restructuring the existing ones.

---

### User Story 7 - Switch between my work and agent-produced work (Priority: P3)

As the single user, I flip one global Mine/Agent switch at the top of the app (defaulting to Mine) and both the Board and the Stats tab immediately scope to that data — my own work (personal and Jira-sourced items) versus whatever the future AI agent team has produced — without navigating to a separate screen.

**Why this priority**: The Agent source has no producer yet (see Assumptions), so today this mostly means the toggle exists and defaults correctly. It becomes more valuable once `backend/peaky-bergers/` ships agent-sourced items, so it's forward-looking rather than day-to-day critical right now.

**Independent Test**: With items in both scopes, toggle to Agent and confirm the Board and Stats both show only agent-sourced data; toggle back to Mine and confirm both return to personal+Jira data — using the exact same Board/Stats layout and components in both scopes.

**Acceptance Scenarios**:

1. **Given** the app just loaded, **When** I haven't touched the switch, **Then** it defaults to `Mine`.
2. **Given** items with `source = personal` or `source = jira`, **When** the switch is set to `Mine`, **Then** the Board and Stats both scope to exactly those items.
3. **Given** items with `source = agent`, **When** I set the switch to `Agent`, **Then** the Board and Stats both scope to exactly those items, using the same layout and components as `Mine` — no separate UI is built for this scope yet.
4. **Given** I flip the switch, **When** the new scope has zero items, **Then** the Board/Stats render their normal empty states rather than an error.

---

### Edge Cases

- What happens when a Jira issue assigned to the user is later unassigned or closed? It stops appearing as an active sync target on the next poll, but the local copy already created is not silently deleted (the user can still see and act on it).
- What happens when the Jira API token is missing, expired, or revoked? Jira sync cycles fail silently in the background (logged without the token value); personal items remain fully functional.
- What happens when a Jira issue reports a status this tracker doesn't recognize? It falls back to `Backlog`, the same way an unrecognized status previously fell back to the old default state.
- How does the system handle deleting an Epic that still has Tasks nested under it? The Tasks become unparented (no Epic) rather than being deleted along with the Epic.
- What happens if a Jira-sourced item is deleted locally but the Jira issue still exists? It reappears on the next sync cycle, since Jira remains the source of truth for issue existence.
- What happens if two edits (local and Jira-side) land in the same polling interval? Last-write-wins by update timestamp, as described in User Story 3's acceptance scenario 4.
- What happens if the user picks more than the usual ~3 items for today? Nothing is blocked — there's no enforced limit, it's a personal guideline, not a system constraint.
- What happens when an item has no Due date? No red/yellow color-coding is applied; it renders like any other item.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow creating, viewing, editing, and deleting work items of type Epic or Task.
- **FR-002**: The system MUST allow a Task to be assigned exactly one Epic as a parent, or none; an Epic MUST NOT have a parent.
- **FR-003**: Every work item MUST have a status that is one of: `Backlog`, `Picked for today`, `In progress`, `Blocked`, `Done` — the same fixed set for every item, regardless of type or source.
- **FR-004**: Every work item MUST have a `source` of `personal`, `jira`, or `agent`, where `agent` is a valid value but is not produced by anything in this feature.
- **FR-005**: The system MUST periodically poll Jira for issues assigned to the user and create or update corresponding local items with `source = jira`, mapping each into `Backlog`, `In progress`, `Blocked`, or `Done` — never into `Picked for today`.
- **FR-006**: The system MUST push local status and description changes on `source = jira` items back to the corresponding Jira issue within one polling interval.
- **FR-007**: When a conflicting change exists on both sides between poll cycles, the system MUST resolve it by keeping whichever side was updated more recently.
- **FR-008**: The system MUST provide a command-line interface capable of, at minimum, adding a new personal Task and listing existing work items.
- **FR-009**: The web UI MUST offer exactly two views — a Board (showing only the `Picked for today` / `In progress` / `Blocked` / `Done` columns, with the full Backlog listed directly beneath it) and a Stats tab — with no separate List, Epics, or Jira-sync-settings views.
- **FR-010**: The CLI and web UI MUST read and write the same underlying data, with no separate or divergent storage.
- **FR-011**: The system MUST run continuously in the background and resume automatically after a reboot or crash, without the user manually restarting it.
- **FR-012**: The system MUST NOT expose the Jira API token in logs, in version control, or in any API response.
- **FR-013**: A failure in Jira connectivity or authentication MUST NOT prevent creating, editing, or viewing personal (non-Jira) work items.
- **FR-014**: The user MUST be able to move an item from `Backlog` to `Picked for today` via an explicit "Pick for today" action or by dragging it onto the Board.
- **FR-015**: The web UI MUST render Jira-sourced items with a distinct tint (background/border) directly on their card, and MUST show the item's real Jira key (e.g. `AUTH-142`) as its badge instead of a generic "Jira" label.
- **FR-016**: The system MUST record the point in time a work item most recently transitioned to `Done`, distinct from any later, unrelated edit to that item.
- **FR-017**: The web UI's Stats tab MUST show, at minimum: current open-item count, items completed in the trailing 7 days, an overall completion rate, total tracked items, and a status-breakdown bar chart — structured so additional metrics can be added later without reworking the existing ones.
- **FR-018**: Every work item MUST display an internal reference key in the form `MOVOZ-#` on its Board/Backlog card and in its detail drawer, shown alongside the Jira key when the item is Jira-sourced. This key MUST be unique across every work item regardless of type — an Epic and a Task MUST NEVER share the same number — and MUST be assigned from a single persistent global sequence, not derived from any other per-row identifier.
- **FR-019**: The web UI MUST provide an item detail drawer exposing every field — title, description, status, Estimate (hours), Due date, internal reference key, and Jira key when present — sized so all fields are visible without the drawer itself needing to scroll.
- **FR-020**: A Due date MUST be color-coded red when overdue and yellow when due within the next 3 days; otherwise it MUST render with no special color.
- **FR-021**: Clicking the app logo MUST navigate to the Board view.
- **FR-022**: The web UI MUST provide a single global Mine/Agent scope switch (default `Mine`) that filters both the Board and the Stats tab by data scope — `Mine` covering `personal` and `jira` sourced items, `Agent` covering `agent`-sourced items — without navigating to a separate screen.
- **FR-023**: The Board and Stats views MUST use the same layout and components in both the `Mine` and `Agent` scopes for now; only the underlying filtered data differs. Diverging UI per scope is explicitly allowed in the future once the Agent producer exists, not required now.
- **FR-024**: The item detail drawer's Epic-assignment control MUST be a searchable, type-to-filter picker rather than a fixed list of all Epics, and MUST include a "Clear" action to unassign the Epic.

### Key Entities

- **Work Item**: A unit of trackable work — either an Epic or a Task. Attributes: title, description, status (one of the five fixed states), source (`personal` / `jira` / `agent`, grouped for view purposes into `Mine` = personal+jira, or `Agent` = agent), optional parent (Epic reference, Task only), Estimate (hours, optional), Due date (optional), a globally unique reference number (`MOVOZ-#`, drawn from one shared sequence across every Epic and Task), creation timestamp, last-updated timestamp, and last-completed timestamp (set when status becomes `Done`, cleared if it moves away from `Done`). When sourced from Jira, also carries the Jira issue key and a link to the issue in Jira.
- **Epic**: A Work Item that groups related Tasks; has no parent of its own. Retained in the data model for future use, though no dedicated Epic-progress view exists now that the Epics tab is removed (see Assumptions).
- **Task**: A Work Item that may belong to at most one Epic.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The user can record a new personal task from either the CLI or the web UI in under 10 seconds.
- **SC-002**: A Jira issue newly assigned to the user appears in the tracker within one polling interval (no manual refresh or Jira lookup needed).
- **SC-003**: A status or description change made locally on a Jira-sourced item is visible on the actual Jira issue within one polling interval, eliminating the need to open Jira to make that same edit.
- **SC-004**: The tracker is available (respondable via CLI or web UI) at any time without the user having manually started it that session, including immediately after a machine reboot.
- **SC-005**: Zero occurrences of the Jira credential appearing in logs, committed files, or any API response, verified by inspection.
- **SC-006**: The user can identify exactly what to work on right now from the Board alone, and what's available to pick up next from the Backlog list beneath it, without switching tabs.
- **SC-007**: The user can answer "how much do I have open, how much did I finish this week, and how am I trending overall" from the Stats tab alone, without manually tallying items.
- **SC-008**: An overdue or soon-due item is noticeable at a glance from its Due-date color, without opening its drawer.
- **SC-009**: The user can find any item's full details — Estimate, Due date, reference key — without leaving the Board, via its drawer.
- **SC-010**: The user can switch between their own work and agent-produced work with one global toggle, without leaving the Board or Stats view.
- **SC-011**: No two work items — Epic or Task — ever display the same `MOVOZ-#`, even as the tracker grows.

## Assumptions

- Single user, no authentication or multi-user access control is needed for this tracker.
- The hierarchy is exactly two levels (Epic → Task); no deeper nesting is supported in v1.
- The workflow state set (`Backlog` / `Picked for today` / `In progress` / `Blocked` / `Done`) is fixed and not user-configurable in v1.
- Jira sync only covers issues assigned to the user, not all issues in any Jira project.
- Jira sync is poll-based, not webhook-based, since the service is not publicly reachable from Jira; it stays implicit and background — there is no dedicated Jira-connection settings screen (no site/token/poll-interval UI), only the git-ignored `.env`.
- The tracker runs locally on the user's own machine(s), bound to localhost or the local network — not deployed publicly.
- The `agent` source is reserved for a future AI agent team (`backend/peaky-bergers/`) that does not exist yet; no producer of `agent`-sourced items is built in this feature.
- Last-write-wins by timestamp is an acceptable conflict resolution strategy given this is a single-user tool with infrequent simultaneous edits.
- Burn-rate and workload stats are computed on demand from existing item timestamps (created/completed) rather than a separate historical snapshot table — accurate enough for a single-user tool with tens to low hundreds of items, and avoids a second source of truth.
- A default trailing window (e.g. the last 30 days for trend data, 7 days for "done this week") is acceptable scope for v1 of the Stats tab; a user-configurable window is not required yet.
- "Done this week" means completed in the trailing 7 days, not the calendar week, for consistency with the existing trend-window approach.
- Overall completion rate is the share of all tracked items (all-time) whose status is `Done` — a lifetime percentage, distinct from the weekly "done this week" raw count.
- "Picked for today" has no enforced item-count limit; the "roughly three items" guideline is a personal practice, not a system constraint.
- The internal reference key (`MOVOZ-#`) is assigned from its own persistent global sequence at creation time, not derived from the item's row id — deliberately decoupled so it stays stable and collision-free across Epics and Tasks regardless of how the underlying storage evolves.
- A task-labels feature was explored and explicitly rejected — not part of this feature.
- A separate "Today" view was explored and rejected in favor of the Board (active statuses only) plus the Backlog list beneath it doing that job instead.
- The Mine/Agent scope split is a view-level filter over the existing `personal`/`jira`/`agent` source values, not a new data field — `Mine` means `source != agent`, `Agent` means `source = agent`.
- Mine and Agent are allowed to render with identical UI today; this is intentional, not an oversight, since the Agent producer doesn't exist yet and its real needs aren't known.
- Epic/Task hierarchy remains part of the data model for future use, but the dedicated Epics-progress view and the flat List/filter view are both removed in this revision — Epic association may still be visible on a card, but no dedicated view surfaces it.
