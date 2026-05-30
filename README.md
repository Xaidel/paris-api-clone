# Hexagonal Architecture Template

A language-agnostic backend template structured around **Hexagonal Architecture**
(Ports & Adapters). Designed for maintainability in teams where both human engineers
and AI coding agents contribute code.

---

## What this is

This is a **template repository** — not a runnable application. It defines:

- A canonical folder structure for hexagonal architecture
- Layer-by-layer contracts that both humans and AI agents can follow
- Pseudocode examples illustrating patterns for each layer
- Documentation on naming conventions, forbidden patterns, and the port-first workflow
- Architecture Decision Records (ADRs) explaining the design choices

Clone this, delete what you don't need, add your language's tooling, and start building.

---

## Why hexagonal architecture?

The core promise: **business logic never knows how it is delivered or stored**.

- Swap Postgres for SQLite without touching a use case
- Add a gRPC endpoint alongside an existing REST API without duplicating logic
- Test every use case without starting a server or a database
- Onboard a new contributor (human or AI) to a layer without needing to understand all others

---

## Why this template is built for AI agents

In the "vibe coding" era, AI agents generate large amounts of code quickly. Without
structure, that speed creates problems: business logic leaks into HTTP handlers, SQL
appears in domain objects, naming becomes inconsistent across sessions.

This template treats AI agents as **first-class contributors** by providing:

- `AGENTS.md` at the root and in each layer — machine-readable contracts an agent
  can load as context before generating code
- Explicit forbidden patterns per layer — negative constraints agents respond well to
- Canonical naming conventions — prevents drift between agent sessions
- A decision flowchart — tells an agent exactly where to put new code
- A self-audit checklist — an agent can verify its own output before submitting

See [`docs/development-workflow.md`](docs/development-workflow.md) for the recommended
human + AI agent development loop.

---

## Repository structure

```
tpl-hexagonal-arch/
│
├── AGENTS.md                    ← AI agent contract (start here)
├── README.md                    ← This file
│
├── src/
│   ├── domain/                  ← Pure business logic. Zero dependencies.
│   │   ├── entities/            ← Objects with identity and lifecycle
│   │   ├── value-objects/       ← Immutable descriptive values
│   │   ├── events/              ← Things that happened (past tense)
│   │   └── services/            ← Domain logic that spans entities
│   │
│   ├── application/             ← Orchestration. Depends on domain + ports only.
│   │   ├── use-cases/           ← Named user actions (commands and queries)
│   │   └── services/            ← Cross-cutting application concerns
│   │
│   ├── ports/                   ← Contracts (interfaces). No implementations.
│   │   ├── inbound/             ← What the app offers to the outside world
│   │   └── outbound/            ← What the app needs from the outside world
│   │
│   ├── adapters/                ← Concrete implementations of ports.
│   │   ├── inbound/             ← Receive input: HTTP, gRPC, WebSocket, CLI, queues
│   │   └── outbound/            ← Call external systems: DB, cache, APIs, queues
│   │
│   └── infrastructure/          ← Wiring only. The composition root.
│       ├── config/              ← Environment-based configuration loading
│       ├── di/                  ← Dependency injection / service wiring
│       └── observability/       ← Logging, tracing, metrics bootstrap
│
├── tests/
│   ├── unit/                    ← Domain + application; no I/O
│   ├── integration/             ← Adapter tests; real or in-memory infrastructure
│   └── e2e/                     ← Full-stack via inbound adapter
│
└── docs/
    ├── architecture.md          ← Hex arch overview and layer diagram
    ├── dependency-rules.md      ← The Dependency Rule, machine-scannable
    ├── naming-conventions.md    ← Canonical naming per layer
    ├── anti-patterns.md         ← Forbidden patterns with explanations
    ├── development-workflow.md  ← Port-first loop; human + AI agent workflow
    └── adr/                     ← Architecture Decision Records
```

---

## Getting started

### 1. Clone the template

```sh
git clone https://github.com/your-org/tpl-hexagonal-arch my-service
cd my-service
rm -rf .git && git init
```

### 2. Add your language's tooling

This template has no language-specific files. Add what you need:

- **Python**: `pyproject.toml`, `uv.lock`, `ruff.toml`
- **Go**: `go.mod`, `go.sum`
- **Rust**: `Cargo.toml`, `Cargo.lock`
- **TypeScript**: `package.json`, `tsconfig.json`

### 3. Read the layer contracts

Before writing any code, read:

1. [`AGENTS.md`](AGENTS.md) — the full architecture contract
2. The `AGENTS.md` in the layer you are working in
3. The `EXAMPLE.md` files in the relevant subdirectory

### 4. Follow the port-first rule

For every new feature:

1. Define the outbound port interface (`src/ports/outbound/`)
2. Write the use case against the port (`src/application/use-cases/`)
3. Implement the adapter (`src/adapters/outbound/`)
4. Wire everything in (`src/infrastructure/di/`)

See [`docs/development-workflow.md`](docs/development-workflow.md) for the full loop.

---

## Key documents

| Document | Purpose |
|---|---|
| [`AGENTS.md`](AGENTS.md) | Root AI agent contract — read first |
| [`docs/architecture.md`](docs/architecture.md) | Full architecture explanation with diagrams |
| [`docs/naming-conventions.md`](docs/naming-conventions.md) | Canonical naming rules |
| [`docs/anti-patterns.md`](docs/anti-patterns.md) | What not to do, and why |
| [`docs/development-workflow.md`](docs/development-workflow.md) | How to build features in this architecture |
| [`docs/adr/`](docs/adr/) | Why these decisions were made |

---

## Language-specific notes

### Python
- Place `__init__.py` files in each layer directory
- Use abstract base classes (`ABC`) or `Protocol` for port interfaces
- Recommended: `uv` for package management, `ruff` for linting

### Go
- Each layer maps naturally to a Go package
- Use Go interfaces for ports
- Recommended: one `main.go` in `src/infrastructure/` as the entry point

### Rust
- Each layer maps to a module or crate
- Use traits for ports
- Recommended: workspace with one crate per layer for strict compile-time enforcement

### TypeScript
- Use TypeScript interfaces or abstract classes for ports
- Each layer can be a separate module with its own `index.ts`
- Recommended: `bun` for runtime and package management

---

## Contributing

When contributing code:

1. Read `AGENTS.md` before writing anything
2. Run the self-audit checklist in Section 7 of `AGENTS.md` before submitting
3. Add an ADR in `docs/adr/` for any significant architectural decision
