# Language-Specific Supplements

This directory contains language-specific supplements to the architecture contracts
defined in the root `AGENTS.md` and each layer's `AGENTS.md`. They narrow the
language-neutral rules into concrete idioms, tooling choices, and real-code examples
for each supported language.

---

## Agent loading order

When working in a specific language, load files in this order:

1. **`/AGENTS.md`** — architecture rules, layer map, naming conventions (language-neutral)
2. **`src/{layer}/AGENTS.md`** — rules scoped to the layer you are working in
3. **`docs/languages/{language}/AGENTS.md`** ← load this file last; it narrows and specialises the above

Each file in the chain inherits from the previous. Later files override earlier ones
for idioms and tooling only. Architecture rules from step 1 are never overridden.

---

## Precedence rule

> Language supplements are authoritative for idioms, tooling, and framework choices.
> They **never** override architecture rules.
>
> If anything in a language supplement appears to conflict with the root `AGENTS.md`
> or a layer `AGENTS.md`, the architecture rule wins. Flag the conflict rather than
> silently choosing the language supplement.

---

## Available supplements

| Language | AGENTS.md | Key frameworks | DB driver |
|---|---|---|---|
| Python | [`python/AGENTS.md`](python/AGENTS.md) | FastAPI, asyncpg | asyncpg |
| TypeScript | [`typescript/AGENTS.md`](typescript/AGENTS.md) | TanStack Start | postgres.js |
| Go | [`go/AGENTS.md`](go/AGENTS.md) | Gin | pgx |
| Rust | [`rust/AGENTS.md`](rust/AGENTS.md) | Axum | sqlx |

---

## Example files

Each supplement has a `examples/` sub-directory with real, runnable code (not
pseudocode) for every architectural layer. All examples use the same `Order` /
`Payment` domain so patterns are consistent across layers.

| File | Layer it illustrates |
|---|---|
| `examples/domain.md` | `src/domain/` — entities, value objects, events, errors |
| `examples/application.md` | `src/application/` — use cases, application services |
| `examples/ports.md` | `src/ports/` — inbound and outbound port contracts |
| `examples/adapters.md` | `src/adapters/` — inbound (HTTP) and outbound (DB, payment) adapters |
| `examples/infrastructure.md` | `src/infrastructure/` — config, DI wiring, observability |
| `examples/tests.md` | `tests/` — in-memory doubles, unit, integration, and E2E tests |

---

## Cross-Language Comparison Matrix

Fast-lookup for key conventions that differ across languages. Always load the full language supplement before writing code.

| Concept | Python | TypeScript | Go | Rust |
|---|---|---|---|---|
| **Root directory** | `src/` | `src/` | `internal/` | `src/` |
| **Entry point** | app factory via `uvicorn` / ASGI | framework entry (TanStack Start) | `cmd/myapp/main.go` | `src/main.rs` |
| **Port mechanism** | `typing.Protocol` + `@runtime_checkable` | `interface` | `interface` (implicit structural) | `trait` + `#[async_trait]` |
| **Port directory structure** | `src/ports/inbound/` + `src/ports/outbound/` | `src/ports/inbound/` + `src/ports/outbound/` | flat `internal/ports/` (`package ports`) | `src/ports/inbound/` + `src/ports/outbound/` |
| **Adapter conformance** | implicit — adapters do NOT inherit from Protocol | explicit `implements InterfaceName` | implicit — Go structural interfaces | explicit `impl TraitName for StructName` |
| **Value object pattern** | `@dataclass(frozen=True)` | `class` + `private constructor` + `readonly` fields | struct + unexported fields + value receivers | `struct` + private fields + `derive(Clone, PartialEq, Eq)` |
| **Monetary type** | `decimal.Decimal` | `decimal.js` | `shopspring/decimal` | `rust_decimal::Decimal` |
| **Absence / null** | `T \| None` | `T \| null` | `(*T, nil)` — `nil` for not found, `err` for failure | `Option<T>` |
| **Error propagation** | `raise` / `except`; typed exception hierarchy | `throw` / `try-catch`; typed exception hierarchy | `(T, error)` return pairs + `fmt.Errorf("%w")` wrapping | `Result<T, E>` + `?` operator; `thiserror` enums |
| **Async model** | `async def` throughout all layers | `async`/`await` throughout all layers | synchronous/blocking; `context.Context` for cancellation | domain + application sync; `async fn` at port traits + adapters only |
| **DI approach** | manual wiring; `dependency_injector` for complex cases | manual wiring; TSyringe for complex cases | manual constructor wiring only (no DI library) | manual constructor wiring + `Arc<T>` for shared state |
| **Module / package system** | `__init__.py` per directory; `src/` on `PYTHONPATH` (not a package itself) | `index.ts` barrel per directory + `tsconfig paths` aliases | short flat `package` names; one package per layer directory | `mod.rs` per directory + explicit `pub mod` in parent; or Cargo workspace |
| **Unit test placement** | `tests/unit/` | `tests/unit/` | colocated `*_test.go` files in `internal/` | inline `#[cfg(test)] mod tests { ... }` blocks |
| **Integration test placement** | `tests/integration/` | `tests/integration/` | colocated `*_test.go` files in adapter packages | Cargo's `tests/` directory |
| **Workspace / crate isolation** | n/a | n/a | n/a | Cargo workspace is the default structure; each crate maps to one architecture layer |
