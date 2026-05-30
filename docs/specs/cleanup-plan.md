# Cleanup Plan: Remove Legacy Embedding, Semantic, Keyword, and Exhaustive-Analysis Code

## Status

Phases 1-5 are complete and merged into the `cleanup/main` integration branch.

Phase 6 complete on `cleanup/phase-6` as the final embedding schema and `pgvector` cleanup phase.

## Summary

This document defines the phased cleanup plan for removing stale, unused, and refactored code related to:

1. Embedding feature code.
2. Semantic matching.
3. Keyword matching.
4. Exhaustive list analysis.

The target end state is a simpler system that keeps:

1. ReAct-based classification.
2. The current step 3 fallback behavior within ReAct handling.
3. Step 4 and step 5 flows.

This plan is intentionally organized as vertical slices aligned to ports and use cases, not by horizontal layer-by-layer deletion. Each phase maps to one sub-issue and one PR. Independent phases should merge independently. Dependent phases must be implemented as stacked PRs.

## Why

The repository still carries two classification eras:

1. Legacy list-based keyword and semantic matching.
2. ReAct-based LLM classification.

The legacy path is intentionally not wired into the primary runtime path, but the supporting code still exists across ports, adapters, domain services, value objects, config, tests, and migrations. In addition, the current ReAct implementation still depends on embedding infrastructure for historical nearest-neighbor reuse. That means the cleanup cannot be completed safely with a single delete-only PR. It requires a phased refactor that first removes isolated slices and then removes the embedding dependency from ReAct itself.

## Planning Constraints

1. Prefer vertical slices by use case or port boundary.
2. Preserve ReAct classification behavior unless a phase explicitly replaces or removes a dependency.
3. Preserve step 4 and step 5 behavior throughout.
4. Avoid broad cross-repo deletions in one PR when a smaller vertical slice is available.
5. Call out stacked-PR dependencies explicitly when phases touch the same persisted models, public result shapes, or DI seams.

## Confirmed Current-State Findings

### Active runtime dependencies

1. The exhaustive-analysis command is isolated and can be removed first.
2. The legacy keyword and semantic matching stack is present but intentionally not wired into the main runtime path.
3. ReAct classification still uses embeddings for historical nearest-neighbor fallback.
4. Public transaction result payloads still expose legacy and embedding-derived fields.

### Key anchors

1. [internal/infrastructure/di/wire.go](../../internal/infrastructure/di/wire.go) shows the composition root, including active ReAct embedding wiring and intentionally unwired legacy matching.
2. [internal/adapters/transaction_classification_react_gateway.go](../../internal/adapters/transaction_classification_react_gateway.go) contains the active ReAct historical matching logic.
3. [internal/application/services/classification_job_handler_react.go](../../internal/application/services/classification_job_handler_react.go) contains the step 3 fallback behavior that must survive cleanup.
4. [internal/domain/value-objects/pipeline_result.go](../../internal/domain/value-objects/pipeline_result.go) is the shared envelope for legacy and ReAct result payloads.
5. [internal/application/use-cases/transaction_result_mapper.go](../../internal/application/use-cases/transaction_result_mapper.go) and [internal/ports/transaction_result.go](../../internal/ports/transaction_result.go) expose legacy and embedding-derived fields through transaction APIs.

## Phase Overview

| Phase | Sub-issue focus | PR shape | Dependency | Parallelizable | Blast radius |
|---|---|---|---|---|---|
| 1 | Remove exhaustive-analysis slice | Standalone delete/refactor PR | None | Yes | Low |
| 2 | Remove legacy list-matching slice | Delete/refactor PR | Phase 1 preferred | Limited | Medium |
| 3 | Remove ReAct embedding fallback | Behavioral refactor PR | Phase 2 | No | High |
| 4 | Remove legacy and embedding-derived API/result fields | Contract cleanup PR | Phase 3 | No | High |
| 5 | Remove shared embedding bootstrap/config/deps | Infra cleanup PR | Phase 4 | No | Medium |
| 6 | Final embedding schema and `pgvector` cleanup | Final migration cleanup PR | Phase 5 | No | High |

## Dependency and PR Stacking Graph

```text
Phase 1: Exhaustive-analysis removal
  |
  v
Phase 2: Legacy list-matching removal
  |
  v
Phase 3: ReAct embedding-fallback removal
  |
  v
Phase 4: Transaction result/API cleanup
  |
  v
Phase 5: Shared embedding infra/config removal
  |
  v
Phase 6: Final embedding schema and pgvector cleanup
```

### Stacking notes

1. Phase 1 is the only clearly independent phase.
2. Phase 2 and Phase 3 both affect classification-related shared types and tests, so Phase 3 should stack on Phase 2.
3. Phase 4 must stack on Phase 3 because API/result cleanup depends on the remaining shape of ReAct review results.
4. Phase 5 must stack on Phase 4 because embedding bootstrap cannot be removed while active runtime or result-mapping code still depends on it.
5. Phase 6 must happen last because schema cleanup should only follow code cleanup.

## Phase Details

## Phase 1: Remove Exhaustive-Analysis Vertical Slice

### Goal

Remove the offline exhaustive-analysis workflow without changing the HTTP server runtime path.

### Why this is a vertical slice

This phase is centered on the exhaustive-analysis use case and CLI entrypoint, not on deleting one layer at a time.

### Main scope

1. Delete the command entrypoint in [cmd/exhaustive-list-analysis/main.go](../../cmd/exhaustive-list-analysis/main.go) and its tests.
2. Delete the use case in [internal/application/use-cases/exhaustive_list_analysis_run_use_case.go](../../internal/application/use-cases/exhaustive_list_analysis_run_use_case.go) and [internal/application/use-cases/exhaustive_analysis_comparison_mapper.go](../../internal/application/use-cases/exhaustive_analysis_comparison_mapper.go).
3. Delete exhaustive-analysis-specific ports such as `RunExhaustiveListAnalysisPort`, `ExhaustiveAnalysisQuerySource`, and `QueryEmbeddingStore`.
4. Delete exhaustive-analysis-specific adapters such as the file query source and file embedding store.
5. Delete the related domain services and value objects used only by exhaustive analysis.
6. Remove `EXHAUSTIVE_ANALYSIS_*` environment variables from [.env.example](../../.env.example).

### Expected blast radius

Low.

Affected areas:

1. One CLI command.
2. One application use case.
3. A small set of dedicated ports and adapters.
4. Related tests and documentation.

Unaffected areas:

1. ReAct worker.
2. Transaction HTTP APIs.
3. Step 4 and step 5 flows.

### Depends on

None.

### PR stacking note

This PR can merge independently and should go first.

### Suggested sub-issue title

Remove exhaustive-analysis command and supporting slice.

## Phase 2: Remove Legacy List-Matching Vertical Slice

### Goal

Remove the dormant keyword and semantic list-matching path for classification lists.

### Why this is a vertical slice

This phase is centered on the legacy list-matching capability itself, spanning its ports, adapters, services, value objects, and tests.

### Main scope

1. Delete keyword matching services in [internal/domain/services/levenshtein_keyword_service.go](../../internal/domain/services/levenshtein_keyword_service.go).
2. Delete semantic matching services in [internal/domain/services/semantic_similarity_service.go](../../internal/domain/services/semantic_similarity_service.go), [internal/domain/services/cosine_similarity.go](../../internal/domain/services/cosine_similarity.go), and [internal/domain/services/cosine_distance.go](../../internal/domain/services/cosine_distance.go).
3. Delete semantic matching value objects such as `SemanticMatchCandidate`, `SemanticReferenceVector`, `SemanticReferenceDistance`, `SemanticAnalysisResult`, and `ListEntryEmbedding`.
4. Delete legacy embedding/list matching ports such as [internal/ports/classification_entry_embedding_repository.go](../../internal/ports/classification_entry_embedding_repository.go) and [internal/ports/embedding_repository.go](../../internal/ports/embedding_repository.go).
5. Delete supporting adapters such as [internal/adapters/classification_entry_embedding_postgres_repository.go](../../internal/adapters/classification_entry_embedding_postgres_repository.go) and [internal/adapters/embedding_postgres_repository.go](../../internal/adapters/embedding_postgres_repository.go).
6. Remove the semantic nearest-search adapter glue from [internal/infrastructure/di/wire.go](../../internal/infrastructure/di/wire.go).

### Expected blast radius

Medium.

Affected areas:

1. Domain services and supporting value objects.
2. Classification-list-related ports and adapters.
3. DI glue that still references legacy semantic search.
4. Related unit and integration tests.

Risk points:

1. Shared types and repository stubs in tests.
2. Any overlooked references in old transaction pipeline code.

Unaffected areas when done correctly:

1. ReAct LLM batching.
2. Step 3 fallback.
3. Step 4 and step 5 handling.

### Depends on

Phase 1 preferred.

### PR stacking note

This can theoretically be started separately, but Phase 1 should merge first to avoid duplicate conflicts around classification-entry embeddings and exhaustive-analysis consumers.

### Suggested sub-issue title

Remove dormant keyword and semantic list-matching stack.

## Phase 3: Remove ReAct Embedding-Based Historical Fallback

### Goal

Keep ReAct classification, but remove its dependency on embedding generation and nearest-neighbor historical lookup.

### Why this is a vertical slice

This phase is centered on the active ReAct classification gateway behavior and the specific outbound port it uses for embedding-based historical reuse.

### Main scope

1. Refactor [internal/adapters/transaction_classification_react_gateway.go](../../internal/adapters/transaction_classification_react_gateway.go) so historical resolution keeps exact goods-description lookup but removes the embedding nearest-neighbor branch.
2. Delete [internal/ports/transaction_description_embedding_repository.go](../../internal/ports/transaction_description_embedding_repository.go).
3. Delete [internal/adapters/transaction_description_embedding_postgres_repository.go](../../internal/adapters/transaction_description_embedding_postgres_repository.go).
4. Remove embedding historical source paths and similarity plumbing from [internal/domain/value-objects/pipeline_result.go](../../internal/domain/value-objects/pipeline_result.go) where no longer needed.
5. Update ReAct gateway and worker tests to reflect exact historical reuse only.

### Expected blast radius

High.

Affected areas:

1. Active classification behavior.
2. ReAct historical reuse behavior.
3. Persisted review-result metadata.
4. ReAct gateway and handler tests.

Risk points:

1. Regressing historical reuse behavior.
2. Accidentally breaking step 3 fallback in [internal/application/services/classification_job_handler_react.go](../../internal/application/services/classification_job_handler_react.go).
3. Leaving stale persisted payload fields behind.

Must remain intact:

1. Exact historical match path.
2. ReAct LLM classification path.
3. Step 3 fallback behavior.

### Depends on

Phase 2.

### PR stacking note

This PR must stack on Phase 2. Both phases change classification-adjacent shared types, tests, and assumptions.

### Suggested sub-issue title

Remove embedding-based historical fallback from ReAct classification.

## Phase 4: Remove Legacy and Embedding-Derived Transaction Result Fields

### Goal

Stop exposing legacy score/detail fields and embedding-derived metadata in transaction read models and HTTP responses.

### Why this is a vertical slice

This phase is centered on the transaction read model and outward-facing contract for classification results.

### Main scope

1. Simplify [internal/ports/transaction_result.go](../../internal/ports/transaction_result.go).
2. Simplify [internal/application/use-cases/transaction_result_mapper.go](../../internal/application/use-cases/transaction_result_mapper.go).
3. Remove no-longer-valid legacy/embedding fields from [internal/adapters/transaction_postgres_repository.go](../../internal/adapters/transaction_postgres_repository.go).
4. Update API-facing transaction response behavior in [internal/adapters/transaction_http_adapter.go](../../internal/adapters/transaction_http_adapter.go) as needed.
5. Reduce `PipelineResult` to the remaining supported result model.

### Expected blast radius

High.

Affected areas:

1. Transaction response DTOs.
2. Persisted JSON serialization and deserialization.
3. Route tests and repository tests.
4. Any consumers expecting legacy score fields.

Risk points:

1. Breaking public or integration-facing response contracts.
2. Incorrectly parsing historical stored pipeline payloads.
3. Leaving mismatch between persistence and mapper layers.

### Depends on

Phase 3.

### PR stacking note

This PR must stack on Phase 3 because the final supported shape of ReAct results is determined there.

### Suggested sub-issue title

Remove legacy and embedding-derived fields from transaction classification results.

## Phase 5: Remove Shared Embedding Infrastructure and Config

### Goal

Delete the shared embedding service, DI wiring, environment parsing, and package dependencies once no active path depends on them.

### Why this is a vertical slice

This phase is centered on the shared embedding capability as an application dependency, spanning config, ports, adapters, and composition root wiring.

### Main scope

1. Delete [internal/ports/embedding_service.go](../../internal/ports/embedding_service.go).
2. Delete [internal/adapters/embedding_service_eino.go](../../internal/adapters/embedding_service_eino.go).
3. Remove `EmbeddingConfig` and embedding-related classification env/config fields from [internal/infrastructure/config/config.go](../../internal/infrastructure/config/config.go).
4. Remove the matching tests from [internal/infrastructure/config/config_test.go](../../internal/infrastructure/config/config_test.go).
5. Remove embedding service/repository construction and exposed application fields from [internal/infrastructure/di/wire.go](../../internal/infrastructure/di/wire.go).
6. Remove obsolete env vars from [.env.example](../../.env.example).
7. Remove unused embedding package dependencies from [go.mod](../../go.mod).
8. Update docs such as [README.md](../../README.md) if they still mention embedding setup.

### Expected blast radius

Medium.

Affected areas:

1. Bootstrap and composition root.
2. Environment/config loading.
3. Dependency graph.
4. Developer setup docs.

Risk points:

1. Accidentally removing configuration still needed by remaining ReAct code.
2. Missing a test helper or stub still referencing embedding ports.

### Depends on

Phase 4.

### PR stacking note

This PR must stack on Phase 4. Shared embedding primitives cannot be removed before result and runtime cleanup are complete.

### Suggested sub-issue title

Remove shared embedding bootstrap, config, and dependencies.

## Phase 6: Final Embedding Schema and pgvector Cleanup

### Goal

Drop the obsolete embedding tables, indexes, and `pgvector`-backed projection artifacts that are no longer referenced after Phases 1-5 removed all application dependencies.

### Why this is a vertical slice

This phase is centered on the last database boundary for the removed embedding capabilities and finishes the cleanup by removing the now-dead schema artifacts.

### Main scope

1. Add [internal/infrastructure/db/migrations/000205_drop_unused_embedding_artifacts.up.sql](../../internal/infrastructure/db/migrations/000205_drop_unused_embedding_artifacts.up.sql) to drop:
   1. `transaction_description_embeddings` and its exact-lookup/HNSW indexes.
   2. `classification_entry_embedding` and its active/HNSW indexes.
   3. `classification_entry` and its active index.
   4. `list_entry_embeddings` and its HNSW index.
2. Add [internal/infrastructure/db/migrations/000205_drop_unused_embedding_artifacts.down.sql](../../internal/infrastructure/db/migrations/000205_drop_unused_embedding_artifacts.down.sql) to always recreate the base `list_entry_embeddings` table, and to recreate the `vector`-backed column/index plus `classification_entry`, `classification_entry_embedding`, and `transaction_description_embeddings` only when the `vector` extension exists.
3. Preserve the historical migrations that originally introduced those tables:
   1. [internal/infrastructure/db/migrations/000014_create_list_entry_embeddings.up.sql](../../internal/infrastructure/db/migrations/000014_create_list_entry_embeddings.up.sql)
   2. [internal/infrastructure/db/migrations/000015_promote_list_entry_embeddings_to_pgvector.up.sql](../../internal/infrastructure/db/migrations/000015_promote_list_entry_embeddings_to_pgvector.up.sql)
   3. [internal/infrastructure/db/migrations/000016_create_classification_entry_projection.up.sql](../../internal/infrastructure/db/migrations/000016_create_classification_entry_projection.up.sql)
   4. [internal/infrastructure/db/migrations/000019_add_react_transaction_classification_support.up.sql](../../internal/infrastructure/db/migrations/000019_add_react_transaction_classification_support.up.sql)
4. Keep migration verification in [internal/infrastructure/db/migrate_test.go](../../internal/infrastructure/db/migrate_test.go) aligned with the final Phase 6 statements.

### Expected blast radius

High.

Affected areas:

1. Database schema.
2. Unconditional rollback of the base `list_entry_embeddings` table, plus `pgvector`-dependent rollback of the vector-backed column/index and the remaining embedding projection tables.
3. Migration testing.
4. Any historical data inspection workflows that still referenced the removed tables.

Risk points:

1. Data loss if removed too early.
2. Rollback complexity.
3. Hidden operational dependencies outside application code.

### Depends on

Phase 5.

### PR stacking note

This phase must be last because it removes the final schema artifacts after all code references are gone.

### Suggested sub-issue title

Finish embedding schema and `pgvector` cleanup after application cleanup.

## Out of Scope

1. Changing the business meaning of step 4 or step 5.
2. Replacing ReAct with another classification architecture.
3. Rewriting historical migrations.
4. Unrelated infrastructure or API cleanup.

## Verification Strategy

### Per-phase verification

1. Run targeted `go test` on touched packages first.
2. Run `go test ./...` after each phase compiles.
3. Run `go build ./...` after phases that affect DI, runtime wiring, or commands.
4. Use repo-wide searches after each phase to confirm deleted concepts are truly gone.

### Targeted searches

1. After Phase 1, search for `exhaustive` and `EXHAUSTIVE_ANALYSIS`.
2. After Phase 2, search for `LevenshteinKeyword`, `SemanticSimilarity`, `ClassificationEntryEmbeddingRepository`, `cosineSimilarity`, and `cosineDistance`.
3. After Phase 3, search for `TransactionDescriptionEmbeddingRepository`, `embedding_historical_match`, and `EmbeddingSimilarity`.
4. After Phase 4, search for `CombinedKeywordScore`, `CombinedSemanticScore`, `KeywordDecision`, and `SemanticDecision` in transaction result and adapter code.
5. After Phase 5, search for `EMBEDDING_`, `CLASSIFICATION_KEYWORD_`, `CLASSIFICATION_SEMANTIC_`, and `OPENAI_EMBEDDING_MODEL`.
6. After Phase 6, search for `DATABASE_USE_PGVECTOR`, `github.com/pgvector/pgvector-go`, `pgvector/pgvector:pg17`, `list_entry_embeddings`, `classification_entry_embedding`, and `transaction_description_embeddings` in live code/config surfaces.

### Final success criteria

1. `go test ./...` passes.
2. `go build ./...` passes.
3. ReAct classification remains functional.
4. Step 3 fallback remains functional.
5. Step 4 and step 5 remain untouched.
6. No HTTP payload exposes legacy keyword/semantic or embedding-derived data.
7. No startup or runtime path depends on embedding infrastructure.
8. The final embedding schema cleanup migration removes obsolete embedding tables and indexes without editing historical migrations.

## Recommended Execution Order

1. Merge Phase 1 independently.
2. Stack Phase 2 on Phase 1.
3. Stack Phase 3 on Phase 2.
4. Stack Phase 4 on Phase 3.
5. Stack Phase 5 on Phase 4.
6. Execute Phase 6 last to finish schema cleanup after the application cleanup is complete.

## Decision Record

1. The plan removes embeddings from ReAct too, not only the dormant legacy matching path.
2. Exact historical match reuse remains part of ReAct unless a separate product decision removes it.
3. Step 3 fallback remains in scope to preserve.
4. Step 4 and step 5 remain out of scope for modification.
5. API/result cleanup is treated as a first-class phase because it changes the outward contract and has its own blast radius.
