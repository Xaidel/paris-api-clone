# Cleanup Plan Status

**Last Updated:** 2026-05-18  
**Overall Progress:** 100% cleanup complete (Phases 1-6 complete)

---

## Phase Overview

| Phase | Title | Status | PR | Branch | Notes |
|-------|-------|--------|----|---------|----|
| 1 | Remove Exhaustive-Analysis Vertical Slice | 🟢 Complete | Merged | `cleanup/main` | Merged into the cleanup integration branch |
| 2 | Remove Legacy List-Matching Stack | 🟢 Complete | Merged | `cleanup/main` | Merged into the cleanup integration branch |
| 3 | Remove ReAct Embedding Fallback | 🟢 Complete | Merged | `cleanup/main` | Merged into the cleanup integration branch ahead of Phase 4 |
| 4 | Remove Legacy/Embedding API Fields | 🟢 Complete | Merged | `cleanup/main` | Merged into the cleanup integration branch and unblocked Phase 5 |
| 5 | Remove Shared Embedding Infrastructure | 🟢 Complete | Merged | `cleanup/main` | Merged into the cleanup integration branch |
| 6 | Final Embedding Schema and pgvector Cleanup | 🟢 Complete | — | `cleanup/phase-6` | Final migration and migration-test coverage implemented |

---

## Phase 1: Remove Exhaustive-Analysis Vertical Slice

**Status:** 🟢 Complete (Merged Into `cleanup/main`)

**Goal:** Remove the offline exhaustive-analysis CLI command without touching HTTP server or ReAct worker.

**Spec Location:** Folded back into `docs/specs/cleanup-plan.md` after merge.

### Current Notes

- Phase 1 is merged into `cleanup/main`
- The dedicated `docs/specs/phase-1/` working folder was removed after merge

---

## Phase 2: Remove Legacy List-Matching Stack

**Status:** 🟢 Complete (Merged Into `cleanup/main`)

**Goal:** Remove the dormant keyword and semantic list-matching path for classification lists.

**Dependencies:**
- ✅ Phase 1 merged into `cleanup/main`

**Spec Location:** See `./cleanup-plan.md` section "Phase 2" for details

### Key Points

- Deletes keyword and semantic matching services
- Removes semantic value objects
- Removes legacy embedding/list matching ports and adapters
- Medium blast radius (affects domain and adapter layers)

### Prerequisites

- [x] Phase 1 merged into `cleanup/main`
- [x] All Phase 1 success criteria verified

### Slice Progress

- [x] Slice 0: Identify Phase 2 legacy code and confirm later-phase boundaries
- [x] Slice 1: Remove dormant `LevenshteinKeywordService` and its tests
- [x] Slice 2: Remove the legacy semantic list-matching flow
- [x] Slice 3: Clean up stale Phase 2 documentation references

### Verification So Far

- [x] `go test ./internal/domain/... -v` passed during implementation
- [x] `go build ./internal/infrastructure/di` passed during implementation
- [x] `go test ./internal/application/use-cases/... -v` passed during implementation
- [x] `go test ./internal/ports/... -v` passed during implementation
- [x] `go test ./internal/adapters/... -v` passed during implementation
- [x] `go test ./internal/application/services -v -run ".*React.*"` passed during implementation
- [x] `go test ./internal/adapters -v -run ".*React.*"` passed during implementation
- [x] Phase 2 docs updated to reflect completed local cleanup state
- [x] Phase 2 merged into `cleanup/main`

### Current Notes

- Phase 2 is merged into `cleanup/main`
- Cleanup branch `cleanup/phase-2` no longer needs tracking here
- The dedicated `docs/specs/phase-2/` working folder was removed after merge
- Phase 3 now stacks on the merged Phase 2 base

---

## Phase 3: Remove ReAct Embedding-Based Historical Fallback

**Status:** 🟢 Complete (Merged Into `cleanup/main`)

**Goal:** Keep ReAct classification but remove its dependency on embedding generation and nearest-neighbor historical lookup.

**Dependencies:**
- Phase 1 ✅
- Phase 2 ✅

**Spec Location:** See `./cleanup-plan.md` section "Phase 3" for details

### Key Points

- Refactors ReAct gateway to use exact historical match only (no embedding fallback)
- Removes transaction description embedding repository
- Updates ReAct tests
- High blast radius (affects active classification behavior)

### Critical Preservation

- ✅ Exact historical match path remains
- ✅ ReAct LLM classification path remains
- ✅ Step 3 fallback remains intact

### Slice Progress

- [x] Slice 1: Remove active ReAct embedding fallback wiring and repository slice
- [x] Slice 2: Drop embedding-derived internal ReAct metadata plumbing
- [x] Slice 3: Verify Phase 3 boundaries and update status docs

### Verification So Far

- [x] Boundary scan confirms `TransactionDescriptionEmbeddingRepository`, `PipelineResultSourceEmbeddingHistoricalMatch`, and `IncReactHistoricalEmbeddingMatchTotal` are gone from `internal/`
- [x] Boundary scan confirms no remaining production `EmbeddingSimilarity()` usage in `internal/`
- [x] `embedding_similarity` remains only as a compatibility field in `internal/adapters/transaction_postgres_repository.go`
- [x] Shared `EmbeddingService` bootstrap remains for Phase 5
- [x] `go test ./internal/adapters -run 'TestReActTransactionClassificationGateway(ReusesExactHistoricalMatchWithoutLLM|FallsBackToLLMAfterExactMiss|SkipsHistoricalNextStepReuse|LeavesExitStepUnsetForNextStep|PreservesClassifierVersionInReactPayload|LogsBatchTokenUsageAtDebug|RetriesTransientLLMFailures)' -count=1`
- [x] `go test ./internal/application/services -run TestReActClassificationJobHandlerHandleBatch -count=1`
- [x] `go test ./internal/application/use-cases -run 'Test(NewTransactionResultPreservesNonTerminalClassification|NewTransactionResultOverridesPipelineClassificationAfterProfessionalReview|NewTransactionResultLeavesEmbeddingSimilarityUnsetForReactPayload)' -count=1`
- [x] `go test ./... -v`
- [ ] `go test ./... -v -race` blocked locally by Windows C toolchain error: `cc1.exe: sorry, unimplemented: 64-bit mode not compiled in`
- [x] `go build ./...`
- [x] `go build ./cmd/server`

### Current Notes

- Phase 3 is merged into `cleanup/main`
- Shared embedding bootstrap, config, and dependencies were intentionally deferred and are handled in Phase 5
- Local race verification remained blocked by the current Windows toolchain during implementation, but non-race verification completed before merge

---

## Phase 4: Remove Legacy and Embedding-Derived Transaction Result Fields

**Status:** 🟢 Complete (Merged Into `cleanup/main`)

**Goal:** Stop exposing legacy score/detail fields and embedding-derived metadata in transaction read models and HTTP responses.

**Dependencies:**
- Phase 1–3 ✅

**Spec Location:** See `./cleanup-plan.md` section "Phase 4" for details

### Key Points

- Simplifies transaction result DTOs
- Removes legacy score fields from API responses
- Reduces `PipelineResult` envelope
- High blast radius (affects public API contracts)

### Slice Progress

- [x] Slice 1: Remove deprecated outward transaction-result fields and mapper-only legacy score details
- [x] Slice 2: Stop writing deprecated legacy matcher detail and score blobs for new payloads
- [x] Slice 3: Drop redundant outward pipeline-result metadata from mapper and HTTP adapter
- [x] Slice 4: Remove deprecated legacy persistence compatibility fields and simplify legacy constructor signatures

### Verification So Far

- [x] `go test ./internal/application/use-cases`
- [x] `go test ./internal/adapters`
- [x] `go test ./internal/domain/value-objects ./internal/domain/entities ./internal/application/use-cases`
- [x] `go test ./...`
- [x] `go build ./...`

### Current Notes

- Phase 4 is merged into `cleanup/main`
- Old stored pipeline-result rows still parse, but new payloads no longer emit deprecated legacy matcher, score, confidence, or embedding-derived metadata fields
- Shared embedding bootstrap, config, and dependencies were deferred and are now implemented in Phase 5

---

## Phase 5: Remove Shared Embedding Infrastructure and Config

**Status:** 🟢 Complete (Merged Into `cleanup/main`)

**Goal:** Delete the shared embedding service, DI wiring, environment parsing, and dependencies once no active path depends on them.

**Dependencies:**
- Phase 1–4 ✅

**Spec Location:** See `./cleanup-plan.md` section "Phase 5" for details

### Key Points

- Deletes embedding service and adapters
- Removes embedding config parsing
- Removes DI wiring for embedding
- Removes unused package dependencies
- Updates documentation

### Slice Progress

- [x] Slice 1: Remove shared embedding port, adapter, and direct test coverage
- [x] Slice 2: Reshape runtime config so ReAct keeps only the still-live OpenAI transport settings
- [x] Slice 3: Remove embedding DI wiring, application exposure, and unused module dependencies
- [x] Slice 4: Update cleanup status docs and verify boundary searches
- [x] Slice 5: Add regression coverage for the moved ReAct bootstrap config seam and remove stale semantic/embedding leftovers

### Verification So Far

- [x] `go test ./internal/infrastructure/config ./internal/ports ./internal/infrastructure/di ./internal/adapters`
- [x] `go test ./internal/adapters`
- [x] `go test ./internal/infrastructure/di -count=1`
- [x] `go test ./internal/domain`
- [x] `go test ./...`
- [x] `go build ./...`
- [x] Boundary scan confirms no production `EmbeddingService`, `EmbeddingConfig`, `buildEmbeddingService`, `embeddingModel`, `WithEmbeddingBatchSize`, or `WithEmbeddingMetrics` symbols remain
- [x] Phase 5 env-name boundary scan is clean in committed code; remaining hits are the spec's search instructions and the local developer `.env`
- [x] Boundary scan confirms stale semantic/embedding-only domain errors are gone

### Current Notes

- Phase 5 is merged into `cleanup/main`
- Phase 5 removes the shared embedding bootstrap entirely rather than leaving a degraded-mode no-op service behind
- ReAct keeps the still-live OpenAI connection settings under `ClassificationConfig`; the deleted embedding config no longer owns runtime AI transport settings
- `go mod tidy` refreshed `go.sum` after removing the direct embedding dependencies
- A targeted regression test now covers the ReAct OpenAI bootstrap config seam in `internal/infrastructure/di`

---

## Phase 6: Final Embedding Schema and pgvector Cleanup

**Status:** 🟢 Complete (`cleanup/phase-6` complete)

**Goal:** Drop the final unused embedding tables, indexes, and `pgvector`-backed schema artifacts after the application cleanup is complete.

**Dependencies:**
- Phase 1–5 ✅

**Spec Location:** See `./cleanup-plan.md` section "Phase 6" for details

### Key Points

- Adds `000205_drop_unused_embedding_artifacts.up.sql` and `.down.sql`
- Drops `transaction_description_embeddings`, `classification_entry_embedding`, `classification_entry`, and `list_entry_embeddings`
- Removes the related exact-lookup and HNSW indexes while preserving historical migrations
- Restores `list_entry_embeddings` on rollback unconditionally, while restoring the vector-backed column/index plus `classification_entry`, `classification_entry_embedding`, and `transaction_description_embeddings` only when the `vector` extension is available
- Closes the final cleanup phase across code, config, and schema

### Final Scope

- Removes the last embedding-era schema artifacts introduced earlier in the cleanup plan
- Preserves historical migration files `000014`, `000015`, `000016`, and `000019` unchanged
- Verifies the new Phase 6 migration statements in `internal/infrastructure/db/migrate_test.go`

### Historical Migrations Referenced By Phase 6

- `000014_create_list_entry_embeddings.up.sql`
- `000015_promote_list_entry_embeddings_to_pgvector.up.sql`
- `000016_create_classification_entry_projection.up.sql`
- `000019_add_react_transaction_classification_support.up.sql`

### Verification So Far

- [x] `go test ./...`
- [x] `go build ./...`
- [x] Boundary scan over `go.mod`, `go.sum`, `docker-compose.yml`, `internal`, `docs/specs`, and `.env.example` confirms `DATABASE_USE_PGVECTOR`, `github.com/pgvector/pgvector-go`, and `pgvector/pgvector:pg17` remain only in migration history, migration tests, or status/spec documentation
- [x] Boundary scan over `internal` and `*.go` files confirms `list_entry_embeddings`, `classification_entry_embedding`, and `transaction_description_embeddings` remain only in migration history and migration-test coverage

### Current Notes

- Phase 6 completes the repository cleanup without editing historical migrations in place
- The remaining embedding-era names in committed code are limited to migration history, migration-test coverage, and cleanup documentation
- Local `go test ./... -v -race` remains blocked by the current Windows C toolchain, so non-race verification is the strongest evidence captured on this machine

---

## Summary Dashboard

```
Legend:
  🔴 Blocked      = Waiting for dependencies
  🟡 Planned      = Ready to start, not started
  🟠 In Progress  = Actively being worked on or implemented locally
  🟣 Pending Review = Implemented locally and awaiting verification/review
  🟢 Complete     = Finished in its tracked branch and, when applicable, merged into `cleanup/main`
  ⚫ Deferred      = Intentionally postponed

Current State (2026-05-18):
  Phase 1: 🟢 Complete     (merged into cleanup/main)
  Phase 2: 🟢 Complete     (merged into cleanup/main)
  Phase 3: 🟢 Complete     (merged into cleanup/main)
  Phase 4: 🟢 Complete     (merged into cleanup/main)
  Phase 5: 🟢 Complete     (merged into cleanup/main)
  Phase 6: 🟢 Complete     (final embedding schema cleanup implemented)
  ───────────────────────────────
  Overall: Cleanup complete
```

---

## How to Update This File

When a phase is started, in progress, or completed:

1. Update the **Last Updated** and **Overall Progress** fields at the top of the file
2. Update the **Status** column in the Phase Overview table
3. Update the relevant phase's **Slice Progress** section with completed checkboxes
4. Update the relevant phase's **Verification So Far** section with verification results
5. Add notes under the relevant phase's **Current Notes** section about what was done
6. Update the **Summary Dashboard** at the end

### Status Codes

- `🔴 Blocked` → Waiting for dependencies
- `🟡 Planned` → Ready to start but not yet begun
- `🟠 In Progress` → Currently being worked on or implemented locally
- `🟣 Pending Review` → Implemented locally and awaiting verification/review
- `🟢 Complete` → Finished in its tracked branch and, when applicable, merged into `cleanup/main`
- `⚫ Deferred` → Intentionally postponed

### PR Template

When creating a Phase PR, include:

```markdown
## Phase [N] Completion Checklist

- [x] All tasks completed and tested
- [x] Success criteria verified
- [x] `go test ./... -v -race` passes, or the blocking environment limitation is documented with substitute verification evidence
- [x] `go build ./...` succeeds
- [x] Code review completed
- [x] STATUS.md updated

## Changes Summary
[List key deletions and modifications]

## Verification
[Output from verification commands]
```

---

## Current Blockers & Issues

No active cleanup blockers are recorded in the plan.

Local `go test ./... -v -race` remains blocked by the Windows C toolchain in this
environment (`cc1.exe` lacks 64-bit support), so full non-race verification is
the strongest evidence currently available on this machine.

---

## Next Action

**Immediate:** Treat the cleanup plan as complete and carry forward only normal
post-cleanup verification or release work.

**Option A (Human-Driven):**
1. Review the final Phase 6 migration and doc updates on `cleanup/phase-6`
2. Confirm the repository-level verification evidence in this file
3. Merge the finished cleanup branch when ready

**Option B (Agent-Driven):**
1. Use the finished cleanup state as the base for any unrelated follow-up work
2. Rerun verification in a Linux/macOS environment if race coverage is required
3. Leave the cleanup docs stable unless new regressions or rollback notes appear

**Target Timeline:**
- Phase 1: Complete
- Phase 2: Complete
- Phase 3: Complete
- Phase 4: Complete
- Phase 5: Complete
- Phase 6: Complete

---

## References

- **Main Cleanup Plan:** `./cleanup-plan.md`
- **Hexagonal Architecture Contract:** `../../AGENTS.md`
- **Codebase Structure:** `../../README.md`

---

## Questions or Updates?

Update this file whenever:
- A phase starts, progresses, or completes
- A blocker is discovered
- A PR is merged
- Timeline estimates change
- New dependencies or issues arise

Keep this file as the single source of truth for cleanup progress.
