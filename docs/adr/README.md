# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) — short documents
that capture significant architectural decisions made in this project.

---

## What is an ADR?

An ADR documents:
1. **The problem** — the context and the question that needed answering
2. **The decision** — what was decided
3. **The alternatives** — what other options were considered
4. **The consequences** — what becomes easier, harder, or different as a result

ADRs are written at decision time and are **never deleted**. If a decision is
reversed, a new ADR is written superseding the old one. The history of decisions
is as valuable as the decisions themselves.

---

## Why ADRs matter for agent-assisted development

AI agents working in a codebase have no memory of previous sessions. ADRs give
future agents (and new human contributors) the context they need to understand
**why** the codebase is structured the way it is — not just how.

Without ADRs, agents will reconsider settled decisions and may propose changes
that were already evaluated and rejected.

---

## How to create an ADR

1. Copy `template.md` to a new file: `{NNNN}-{short-title}.md`
   (e.g. `0003-use-kafka-for-event-publishing.md`)
2. Increment the number from the last ADR
3. Fill in all sections
4. Set status to `Proposed`, then `Accepted` when agreed upon

---

## Status values

| Status | Meaning |
|---|---|
| `Proposed` | Under discussion — not yet accepted |
| `Accepted` | Decision made — in effect |
| `Deprecated` | Still in use but being phased out |
| `Superseded by ADR-NNNN` | Replaced by a later decision |

---

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](0001-use-hexagonal-architecture.md) | Use Hexagonal Architecture | Accepted |
| [0002](0002-ai-agent-conventions.md) | Establish AI Agent Contracts | Accepted |
