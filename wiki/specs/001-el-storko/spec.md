# Feature Specification: el-storko — Personal Work-Item Tracker

**Feature Branch**: `001-el-storko`

**Created**: 2026-09-21

**Status**: Draft

**Input**: User description: "el-storko — personal work-item tracker (Jira-like), unifying personal tasks/epics with bidirectionally-synced Jira work items, plus a placeholder source for a future AI agent team."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Track personal tasks and epics (Priority: P1)

As the single user, I create, edit, and delete Tasks and Epics (title, description, status), and nest Tasks under an Epic, so personal projects and chores live in one durable place instead of nowhere.

**Why this priority**: This is the core value of the tracker and works with zero external dependencies. Without it there is nothing to build on.

**Independent Test**: Create an Epic, create two Tasks under it, edit one Task's status and description, delete the other Task, delete the Epic. Confirm state persists across a service restart.

**Acceptance Scenarios**:

1. **Given** no existing items, **When** I create an Epic with a title, **Then** it appears in the tracker with status `todo` and no parent.
2. **Given** an existing Epic, **When** I create a Task with that Epic as parent, **Then** the Task appears nested under the Epic.
3. **Given** a Task with a parent Epic, **When** I try to set another Epic as its parent, **Then** the Task's parent is simply reassigned (a Task has at most one Epic parent at a time).
4. **Given** an existing Task or Epic, **When** I edit its title, description, or status, **Then** the change is saved and visible immediately.
5. **Given** an existing Task or Epic, **When** I delete it, **Then** it no longer appears in any view.
6. **Given** the service has been restarted, **When** I view the tracker, **Then** all previously created items are still present.

---

### User Story 2 - See Jira work automatically, without leaving the tracker (Priority: P2)

As the single user, my Jira-assigned issues automatically appear in the tracker, and a local status or description edit on a Jira-sourced item syncs back to Jira, so I have one place to check and update instead of two.

**Why this priority**: This is the feature that eliminates the "scattered across two tools" problem, but it depends on User Story 1's data model existing first.

**Independent Test**: With a Jira issue assigned to the user, wait one polling interval and confirm it appears locally with the correct source and Jira link. Change its status locally and confirm the change appears on the Jira issue within one polling interval. Change the same issue in Jira and confirm the local copy picks up the change on the next poll.

**Acceptance Scenarios**:

1. **Given** a Jira issue assigned to the user that isn't yet in the tracker, **When** a sync cycle runs, **Then** the issue appears locally with `source = jira`, its Jira key, and a link back to Jira.
2. **Given** a Jira-sourced item in the tracker, **When** I change its status or description locally, **Then** the same change appears on the Jira issue within one polling interval.
2b. **Given** a Jira issue's status or description changes directly in Jira, **When** the next sync cycle runs, **Then** the local item reflects that change.
3. **Given** a Jira-sourced item and a conflicting edit made in both places between polls, **When** the sync cycle runs, **Then** the more recently updated side wins and the other is overwritten.
4. **Given** the Jira sync fails (e.g. network error, invalid token), **When** the failure happens, **Then** personal (non-Jira) items remain fully usable and the failure doesn't crash the service.

---

### User Story 3 - Capture and review work quickly (Priority: P2)

As the single user, I can add and list tasks from the CLI for quick capture, and view all work as a kanban board (grouped by status) or a flat list filterable by source or Epic, so I can check or update things without friction regardless of context.

**Why this priority**: This is the day-to-day interaction layer. It depends on User Story 1's data existing, and benefits from — but does not strictly require — User Story 2.

**Independent Test**: From a terminal, add a task and list tasks, confirming the new task appears. Separately, open the web UI and confirm the same task appears on both the kanban board and the flat list, and that filtering by source/Epic narrows results correctly.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** I run the CLI add command with a title, **Then** a new personal Task is created and is immediately visible in the web UI.
2. **Given** existing items, **When** I run the CLI list command, **Then** it prints the same items visible in the web UI.
3. **Given** items in different statuses, **When** I open the kanban board, **Then** items are grouped into columns matching their status.
4. **Given** items from multiple sources and Epics, **When** I apply a source or Epic filter in the flat list view, **Then** only matching items are shown.
5. **Given** items from more than one source, **When** I view the kanban board or flat list, **Then** a Jira-sourced item is visually distinguishable by color from a personal (or agent) item at a glance, without needing to read its source label.

---

### User Story 4 - See burn-rate stats (Priority: P3)

As the single user, I have a Stats tab showing how fast I'm completing work (throughput) and how my open backlog is trending over time, so I can tell at a glance whether I'm keeping up or falling behind — with room to add more metrics to the same tab later.

**Why this priority**: This is an insight/reporting layer on top of data User Story 1 already produces. It adds no new capability to create or sync work, so it's valuable but not required for the tracker to be useful day to day.

**Independent Test**: Complete a few items on different days, open the Stats tab, and confirm the throughput trend reflects the actual completion days and the backlog trend reflects the actual count of open items on each of those days.

**Acceptance Scenarios**:

1. **Given** items completed on different days over the past month, **When** I open the Stats tab, **Then** I see a throughput view showing how many items were completed per day/week over that period.
2. **Given** items created and completed at various points over the past month, **When** I open the Stats tab, **Then** I see a backlog view showing the open (not-done) item count trending over that same period.
3. **Given** a `done` item is edited (title/description) without changing its status, **When** I view the Stats tab, **Then** its completion date used for both metrics remains the day it was originally marked `done`, not the day of the later edit.
4. **Given** a `done` item is moved back to a non-`done` status and later completed again, **When** I view the Stats tab, **Then** it counts toward the day of its most recent completion, not a stale earlier one.
5. **Given** the Stats tab exists today with these two metrics, **When** a future metric is added, **Then** it can be added as another card on the same tab without restructuring the existing ones.

---

### Edge Cases

- What happens when a Jira issue assigned to the user is later unassigned or closed? It stops appearing as an active sync target on the next poll, but the local copy already created is not silently deleted (the user can still see and act on it).
- What happens when the Jira API token is missing, expired, or revoked? Jira sync cycles fail silently in the background (logged without the token value); personal items remain fully functional.
- How does the system handle deleting an Epic that still has Tasks nested under it? The Tasks become unparented (no Epic) rather than being deleted along with the Epic.
- What happens if a Jira-sourced item is deleted locally but the Jira issue still exists? It reappears on the next sync cycle, since Jira remains the source of truth for issue existence.
- What happens if two edits (local and Jira-side) land in the same polling interval? Last-write-wins by update timestamp, as described in User Story 2's acceptance scenario 3.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow creating, viewing, editing, and deleting work items of type Epic or Task.
- **FR-002**: The system MUST allow a Task to be assigned exactly one Epic as a parent, or none; an Epic MUST NOT have a parent.
- **FR-003**: Every work item MUST have a status that is one of: `todo`, `in_progress`, `blocked`, `done` — the same fixed set for every item, regardless of type or source.
- **FR-004**: Every work item MUST have a `source` of `personal`, `jira`, or `agent`, where `agent` is a valid value but is not produced by anything in this feature.
- **FR-005**: The system MUST periodically poll Jira for issues assigned to the user and create or update corresponding local items with `source = jira`, retaining the Jira issue key and a link to the issue.
- **FR-006**: The system MUST push local status and description changes on `source = jira` items back to the corresponding Jira issue within one polling interval.
- **FR-007**: When a conflicting change exists on both sides between poll cycles, the system MUST resolve it by keeping whichever side was updated more recently.
- **FR-008**: The system MUST provide a command-line interface capable of, at minimum, adding a new personal Task and listing existing work items.
- **FR-009**: The system MUST provide a web interface offering both a kanban board (grouped by status) and a flat list view, with filtering by source and by Epic.
- **FR-010**: The CLI and web UI MUST read and write the same underlying data, with no separate or divergent storage.
- **FR-011**: The system MUST run continuously in the background and resume automatically after a reboot or crash, without the user manually restarting it.
- **FR-012**: The system MUST NOT expose the Jira API token in logs, in version control, or in any API response.
- **FR-013**: A failure in Jira connectivity or authentication MUST NOT prevent creating, editing, or viewing personal (non-Jira) work items.
- **FR-014**: The web UI MUST render Jira-sourced work items with a distinct color treatment from personal/agent items, in both the kanban board and the flat list view.
- **FR-015**: The system MUST record the point in time a work item most recently transitioned to `done`, distinct from any later, unrelated edit to that item.
- **FR-016**: The web UI MUST provide a Stats tab showing, at minimum, a completion-throughput trend and an open-backlog trend over a recent time window, structured so additional metrics can be added to the same tab later without reworking the existing ones.

### Key Entities

- **Work Item**: A unit of trackable work — either an Epic or a Task. Attributes: title, description, status (one of the four fixed states), source (`personal` / `jira` / `agent`), optional parent (Epic reference, Task only), creation timestamp, last-updated timestamp, and last-completed timestamp (set when status becomes `done`, cleared if it moves away from `done`). When sourced from Jira, also carries the Jira issue key and a link to the issue in Jira.
- **Epic**: A Work Item that groups related Tasks; has no parent of its own.
- **Task**: A Work Item that may belong to at most one Epic.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The user can record a new personal task from either the CLI or the web UI in under 10 seconds.
- **SC-002**: A Jira issue newly assigned to the user appears in the tracker within one polling interval (no manual refresh or Jira lookup needed).
- **SC-003**: A status or description change made locally on a Jira-sourced item is visible on the actual Jira issue within one polling interval, eliminating the need to open Jira to make that same edit.
- **SC-004**: The tracker is available (respondable via CLI or web UI) at any time without the user having manually started it that session, including immediately after a machine reboot.
- **SC-005**: Zero occurrences of the Jira credential appearing in logs, committed files, or any API response, verified by inspection.
- **SC-006**: The user can locate "everything I need to do today" — across personal and Jira-sourced items — in a single view, without switching tools.
- **SC-007**: The user can answer "am I completing work faster or slower than before, and is my backlog growing or shrinking" from the Stats tab alone, without manually tallying items.

## Assumptions

- Single user, no authentication or multi-user access control is needed for this tracker.
- The hierarchy is exactly two levels (Epic → Task); no deeper nesting is supported in v1.
- The workflow state set (`todo` / `in_progress` / `blocked` / `done`) is fixed and not user-configurable in v1.
- Jira sync only covers issues assigned to the user, not all issues in any Jira project.
- Jira sync is poll-based, not webhook-based, since the service is not publicly reachable from Jira.
- The tracker runs locally on the user's own machine(s), bound to localhost or the local network — not deployed publicly.
- The `agent` source is reserved for a future AI agent team (`backend/peaky-bergers/`) that does not exist yet; no producer of `agent`-sourced items is built in this feature.
- Last-write-wins by timestamp is an acceptable conflict resolution strategy given this is a single-user tool with infrequent simultaneous edits.
- Burn-rate stats are computed on demand from existing item timestamps (created/completed) rather than a separate historical snapshot table — accurate enough for a single-user tool with tens to low hundreds of items, and avoids a second source of truth.
- A default trailing window (e.g. the last 30 days) is an acceptable scope for v1 of the Stats tab; a user-configurable window is not required yet.
