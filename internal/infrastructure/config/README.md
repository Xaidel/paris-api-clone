# Configuration

The `config/` directory loads environment variables and external config sources
into **typed, validated configuration structs** that are passed to adapters and
infrastructure components at startup.

---

## Rules

1. All configuration loaded here — never `os.getenv()` scattered in adapter code
2. Typed structs — never raw strings passed around
3. Validated at startup — fail fast if required config is missing or invalid
4. No business logic in config
5. No domain types in config — config is infrastructure

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
