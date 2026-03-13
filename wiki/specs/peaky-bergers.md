# Product Spec: Peaky Bergers

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-03-13
> **Project**: peaky-bergers
> **Location**: `backend/peaky-bergers/`
> **Stack**: TBD

## Overview

Peaky Bergers is an AI software team — a collection of agents that help build software. Instead of being a single assistant, it acts as a team of specialized agents that can take on different roles in the development process.

## Problem Statement

Building software solo means wearing every hat — architect, developer, reviewer, tester, project manager. AI can fill these roles, but current tools are generic single-agent assistants. A purpose-built team of specialized agents, each with defined responsibilities and workflows, is more effective than one generalist.

## Goals

- Assemble a team of AI agents with distinct roles (e.g., architect, developer, reviewer, PM)
- Agents collaborate to complete software tasks end-to-end
- Reduce the overhead of solo development by delegating work to agents
- Integrate with existing development workflows (git, CI/CD, specs)

## Non-Goals

- Not a general-purpose AI platform — focused on software development
- Not a hosted SaaS product — personal tooling
- Not replacing human decision-making — agents assist, user decides

## Team Roles

<!-- Define the agent roles that make up the team. -->

| Role | Responsibility | Description |
|------|---------------|-------------|
| | | |

## Workflows

<!-- Define how agents collaborate on common tasks. -->

### Example: Feature Development
1. PM agent breaks down spec into tasks
2. Architect agent designs the approach
3. Developer agent implements
4. Reviewer agent reviews code
5. PM agent validates against spec

## Integration Points

- **Specs**: Reads from `wiki/specs/` to understand requirements
- **Codebase**: Works within the movoz monorepo
- **Git**: Creates branches, commits, PRs
- **CI/CD**: Triggers and monitors pipelines

## Open Questions

- [ ] What framework/stack for building agents? (Claude Agent SDK, LangGraph, custom?)
- [ ] How do agents communicate with each other? (message passing, shared state, orchestrator?)
- [ ] How much autonomy should agents have? (fully autonomous vs human-in-the-loop?)
- [ ] Should agents have persistent memory across sessions?
- [ ] How to handle agent disagreements (e.g., reviewer rejects developer's code)?
