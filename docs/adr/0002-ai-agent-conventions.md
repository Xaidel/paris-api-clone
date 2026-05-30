# ADR-0002: Establish AI Agent Contracts (AGENTS.md)

**Date**: 2026-03-03
**Status**: Accepted
**Deciders**: Project founders

---

## Context

AI coding agents (GitHub Copilot, Claude, GPT-4, Cursor, and similar tools) are
increasingly used to generate, modify, and review code. In the "vibe coding" era,
these agents are capable of producing large amounts of code quickly — but without
architectural awareness, that speed creates new maintenance problems:

1. **Layer violations**: Agents place code where it's most convenient, not where it
   architecturally belongs. Business logic ends up in HTTP handlers. SQL appears in
   domain objects.

2. **Naming drift**: Each agent session may invent different names for the same
   concept. `OrderManager`, `OrderService`, `OrderHandler`, and `OrderController`
   accumulate as synonyms for the same abstraction.

3. **Missing ports**: Agents implement adapters without defining port interfaces
   first, coupling use cases directly to concrete implementations.

4. **No memory across sessions**: An agent working in session N has no knowledge
   of decisions made in session N-1. Without explicit written contracts, settled
   decisions are re-litigated.

5. **Incomplete wiring**: Agents write new components but forget to register them
   in the composition root.

We needed a mechanism to communicate architectural constraints to AI agents in a
format they can load, understand, and self-verify against — before generating code.

---

## Decision

Adopt **AGENTS.md files** as machine-readable architectural contracts placed at:
- The repository root (`AGENTS.md`) — full architecture overview
- Each layer directory (`src/{layer}/AGENTS.md`) — layer-specific rules

Each AGENTS.md file contains:
1. A description of the layer's purpose
2. An explicit allowed/forbidden imports table
3. Naming conventions for the layer
4. Forbidden patterns (explicit negative constraints)
5. A self-audit checklist the agent runs before submitting output

The root AGENTS.md additionally contains:
- A decision flowchart ("where does this code go?")
- The full dependency rule table
- Cross-layer naming conventions
- A pointer to layer-specific AGENTS.md files

---

## Alternatives considered

### Option A: README.md only (human documentation)

Document the architecture in README files written for human readers.

- **Pros**: Familiar format, easy to write
- **Cons**: README files are narrative — they explain, they don't constrain.
  Agents read them but don't treat them as executable rules. Negative constraints
  ("never do X") are less prominent than positive descriptions. Agents frequently
  ignore or override them when they conflict with "making the code work".

### Option B: Linter rules only (automated enforcement)

Rely entirely on architectural linters (import-linter, dependency-cruiser, go-arch-lint)
to enforce rules.

- **Pros**: Machine-enforced, no room for exceptions
- **Cons**: Linters catch violations after the fact — they don't prevent agents from
  writing the wrong code in the first place. An agent that violates rules must iterate
  through failures. More importantly, linters don't communicate naming conventions,
  decision flowcharts, or the "why" behind constraints.

### Option C (chosen): AGENTS.md as layer contracts

- **Pros**: Agents can load the relevant AGENTS.md before generating code, giving them
  the constraints in a format they respond well to (explicit rules, forbidden lists,
  checklists). Works with any AI agent tool. Complements linters (which enforce after
  the fact) with guidance before the fact. Scales — each layer has its own scoped
  contract so agents don't need to load the full picture when working in one layer.
- **Cons**: Requires discipline to keep AGENTS.md files up to date as the architecture
  evolves. An outdated AGENTS.md gives agents incorrect guidance. Mitigation: treat
  AGENTS.md updates as part of any architectural change.

---

## Consequences

### Positive
- AI agents can be pointed at `AGENTS.md` before generating code — reducing
  layer violations, naming drift, and missing ports
- Human engineers have a single authoritative reference for architectural rules
- Self-audit checklists make code review faster: reviewer can verify the checklist
  was followed rather than discovering violations from scratch
- Layer-specific contracts keep context size small — agents working in one layer
  load only that layer's AGENTS.md
- The pattern is tool-agnostic: works with any AI coding assistant

### Negative
- AGENTS.md files must be maintained as the architecture evolves. Stale contracts
  are worse than no contracts — they actively mislead.
- Agents are not guaranteed to follow AGENTS.md — they can still produce violations.
  AGENTS.md reduces the frequency of violations; linters must catch what slips through.
- Adds documentation overhead: every new layer or significant convention change
  requires an AGENTS.md update.

### Neutral / risks
- Risk that AGENTS.md files grow too large and agents skip reading them.
  Mitigation: keep each layer's AGENTS.md focused and scannable. The root AGENTS.md
  is the comprehensive reference; layer files are the scoped quick-reference.
- Different AI agent tools may read AGENTS.md differently (some tools automatically
  inject it as context; others require the user to include it explicitly).

---

## Notes

The `AGENTS.md` filename convention is compatible with tools that auto-inject
project-level context files (such as OpenCode's `AGENTS.md` support, Cursor's
`.cursorrules`, and similar). Projects using tools with different convention
filenames can symlink or copy the content.
