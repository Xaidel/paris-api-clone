# Dependency Injection / Composition Root

The `di/` directory contains the **composition root** — the single place where
all components are instantiated and wired together.

---

## Rules

1. This is the only place where `new`/`construct` is called on concrete classes
2. All dependencies are injected through constructors
3. No global singletons or service locators
4. No business logic
5. Wire once at startup; adapters and use cases never instantiate their dependencies

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
