<!--
SYNC IMPACT REPORT
==================
Version change: 1.0.0 → 1.1.0
Bump rationale: MINOR — added a new core principle (VI. Disciplined Implementation
Workflow) mandating the Superpowers workflow for executing task lists.

Modified principles: None renamed or redefined.
Added sections:
  - Core Principle VI. Disciplined Implementation Workflow (Superpowers)
Removed sections: None

Templates requiring updates:
  - ✅ .specify/templates/plan-template.md (Constitution Check gate added for Principle VI)
  - ✅ .specify/templates/spec-template.md (reviewed — no changes required)
  - ✅ .specify/templates/tasks-template.md (reviewed — no changes required)
  - ✅ .claude/skills/speckit-* command guidance (reviewed — generic, no agent-specific refs to fix)

Runtime guidance: CLAUDE.md remains the agent-facing runtime guidance file and is
consistent with these principles.

Follow-up TODOs: None. RATIFICATION_DATE unchanged (first formal adoption 2026-06-02).
-->

# movoz Constitution

## Core Principles

### I. Spec-Driven Development (NON-NEGOTIABLE)

Every non-trivial feature or change MUST begin with a spec in `wiki/` before any
implementation code is written. The spec defines the *what* and the *why*; it is the
contract that implementation is built against.

- A change is "non-trivial" when it spans 3+ steps, introduces a new component/service,
  or makes an architectural decision. Typo fixes and obvious one-line corrections are exempt.
- Before starting work, existing specs and context in `wiki/` MUST be checked.
- After implementation, the relevant `wiki/` docs (architecture, component specs, technical
  design) MUST be updated to reflect the current state.

Rationale: Writing the spec first removes ambiguity, exposes edge cases early, and keeps the
knowledge base authoritative rather than reconstructed after the fact.

### II. Simplicity First (YAGNI)

Every change MUST be as simple as possible and impact the minimal amount of code required to
satisfy the spec. Build for the requirement in front of you, not for speculative futures.

- No temporary fixes and no band-aids: find and address the root cause.
- Added complexity MUST be justified against a concrete, present requirement. If it cannot be
  justified, it MUST be removed.
- When a fix feels hacky, stop and implement the elegant solution informed by everything now known.

Rationale: Minimal, root-cause changes keep a multi-language personal monorepo maintainable by
one person and prevent accidental architecture.

### III. Verification Before Done (NON-NEGOTIABLE)

No task is complete until its correctness has been demonstrated. Claims of "done", "fixed", or
"passing" MUST be backed by observed evidence, never assertion.

- Run the relevant tests, builds, or the app itself and confirm the output before claiming success.
- Where behavior changes, diff behavior between the baseline and the change.
- If a verification step fails or is skipped, that MUST be stated plainly with the evidence.

Rationale: Evidence-before-assertion is the difference between work that ships and work that
silently regresses.

### IV. Shared Foundations & Consistent Structure

Cross-cutting concerns MUST be shared, and each component MUST follow the established layout for
its language and category. Divergence is a defect unless explicitly justified in a spec.

- Frontend apps MUST consume the shared packages: `@movoz/tailwind-config` (Tailwind preset),
  `@movoz/theme` (theme system), and `@movoz/tsconfig` (TypeScript base).
- New frontend zones MUST follow the Multi-Zones pattern (basePath/assetPrefix + Nginx route).
- Backend services MUST follow the idiomatic layout for their language (e.g. standard Go layout
  for Go services, modular command structure for Rust CLIs).

Rationale: Shared foundations eliminate drift across apps and keep theming, config, and tooling
behaving identically everywhere.

### V. Knowledge Base as Single Source of Truth

`wiki/` is the single source of truth for specs, architecture, design tokens, and component
contracts. Documentation is part of the deliverable, not an afterthought.

- Specs live in `wiki/specs/`; technical and architecture docs live in `wiki/technical/`.
- When implementation diverges from the docs, the docs MUST be updated in the same unit of work.
- Decision writeups, bug-fix analyses, and technical designs MUST use the templates in `wiki/`.

Rationale: A trustworthy knowledge base is what makes spec-driven development possible over time;
stale docs are worse than none.

### VI. Disciplined Implementation Workflow (NON-NEGOTIABLE)

Implementation of any task list MUST follow the Superpowers workflow, in order:

1. **Worktree** — isolate the work in a dedicated git worktree before touching code.
2. **TDD (red-green-refactor)** — write a failing test first, make it pass with the simplest
   change, then refactor. No implementation code is written before a failing test exists.
3. **Subagent-driven execution** — execute independent tasks via focused subagents (one task per
   subagent) to keep the main context clean and parallelize where possible.
4. **Code review** — request and address review before integrating; the work MUST be verified
   against the spec.
5. **Finish branch** — complete the work through the structured finish-branch flow (merge, PR, or
   cleanup) rather than leaving the branch dangling.

Rationale: A fixed, isolated, test-first execution path keeps changes safe to integrate, keeps the
main workspace clean, and guarantees every task list ends in a reviewed, intentionally integrated
state rather than ad-hoc edits.

## Technology & Architecture Standards

- **Monorepo**: pnpm workspaces + Turborepo for the frontend; backend services live under
  `backend/`, infrastructure under `infra/`, knowledge base under `wiki/`.
- **Frontend**: Next.js (App Router) using the Multi-Zones architecture; standalone output for
  Docker deployment; theme state synced across zones via shared localStorage.
- **Backend**: Each service is self-contained with its own build/test tooling
  (Rust/Cargo, Go modules). Services expose clear interfaces (CLI or REST).
- **Infrastructure**: Terraform (AWS, us-west-2) with workspace-based dev/prod environments;
  Nginx reverse proxy + Docker Compose for multi-zone deployment.
- **CI/CD**: GitHub Actions in `.github/workflows/` governs infrastructure plan/apply, cost
  analysis on PRs, and scheduled dev-resource cleanup.

## Development Workflow & Quality Gates

- **Plan first**: Non-trivial work (3+ steps or architectural decisions) MUST be planned before
  implementation. If work goes sideways, stop and re-plan rather than pushing forward.
- **Spec gate**: Implementation MUST NOT begin until a reviewed spec exists in `wiki/`.
- **Verification gate**: Tests/builds MUST pass and behavior MUST be demonstrated before a task
  is marked complete (see Principle III).
- **Documentation gate**: Relevant `wiki/` docs MUST be updated as part of completing the work.
- **Delivery**: Changes are committed and pushed at the end of a plan's execution. Outward-facing
  or hard-to-reverse actions require confirmation unless explicitly authorized.
- **Self-improvement**: After any correction, the lesson MUST be captured so the same mistake is
  not repeated.

## Governance

This constitution supersedes other ad-hoc practices in the repository. Where guidance conflicts,
the order of precedence is: explicit user instruction → this constitution → runtime agent
guidance (`CLAUDE.md`) → default tooling behavior.

- **Amendments**: Changes to this constitution MUST be documented in the Sync Impact Report,
  version-bumped per the policy below, and propagated to dependent templates and guidance docs in
  the same change.
- **Versioning policy** (semantic):
  - MAJOR — backward-incompatible governance/principle removals or redefinitions.
  - MINOR — a new principle/section is added or guidance is materially expanded.
  - PATCH — clarifications, wording, and non-semantic refinements.
- **Compliance review**: Every plan and review MUST verify that the work complies with these
  principles. Any deviation MUST be justified in the plan's Complexity Tracking / Constitution
  Check section, or the deviation MUST be removed.
- **Runtime guidance**: `CLAUDE.md` provides day-to-day operational guidance and MUST remain
  consistent with this constitution.

**Version**: 1.1.0 | **Ratified**: 2026-06-02 | **Last Amended**: 2026-06-02
