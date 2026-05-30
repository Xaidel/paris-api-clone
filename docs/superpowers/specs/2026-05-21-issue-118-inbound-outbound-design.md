# Issue 118 Design: Inbound/Outbound Directory Refactor

## Summary

Refactor `internal/ports` and `internal/adapters` so files are organized into
`inbound/` and `outbound/` subdirectories while preserving existing package names
(`package ports` and `package adapters`) and keeping runtime behavior unchanged.

The implementation will be delivered as two commits:

1. inbound refactor for both ports and adapters
2. outbound refactor for both ports and adapters

Some files are package-level shared files rather than strictly inbound or outbound.
Those files will be included in whichever of the two commits is required to keep
that commit buildable. The commit messages will explicitly call out that
justification.

## Goals

- Make directional boundaries clearer in the filesystem layout
- Preserve existing package names and behavior
- Keep the refactor limited to file organization and any minimal reference updates
- Produce exactly two logical commits aligned to inbound and outbound work

## Non-Goals

- No business logic changes
- No API contract changes
- No persistence behavior changes
- No package renaming
- No architectural redesign beyond directory organization

## Constraints

- This is a Go codebase using flat `package ports` and `package adapters`
- Go package names must remain unchanged after moving files into subdirectories
- The two commits must remain coherent and should be kept buildable
- Shared package files may need to move in the commit where they are required for
  package integrity

## Current State

Today, `internal/ports/*.go` mixes inbound ports, outbound ports, and shared port
types in one directory. `internal/adapters/*.go` mixes inbound HTTP adapters,
outbound persistence and gateway adapters, and shared adapter helpers in one
directory.

Examples:

- inbound ports: `*_create_port.go`, `*_get_port.go`, `*_list_port.go`,
  `*_update_port.go`, `*_delete_port.go`
- outbound ports: `*_repository.go`, `*_gateway.go`, `*_store.go`,
  `*_queue.go`, `*_manager.go`, `*_parser.go`, `event_publisher.go`,
  `password_hasher.go`
- inbound adapters: `*_http_adapter.go`, `http_common.go`
- outbound adapters: `*_postgres_repository.go`, `*_gateway.go`, `*_store.go`,
  `*_queue.go`, `*_parser.go`, `*_pgx.go`, `password_hasher_bcrypt.go`

## Proposed Structure

### Ports

- `internal/ports/inbound/`
- `internal/ports/outbound/`

All files in both directories will continue using `package ports`.

### Adapters

- `internal/adapters/inbound/`
- `internal/adapters/outbound/`

All files in both directories will continue using `package adapters`.

## File Classification Rules

### Inbound Ports

Move use-case entrypoint interfaces into `internal/ports/inbound/`.

This includes files such as:

- `*_create_port.go`
- `*_get_port.go`
- `*_list_port.go`
- `*_update_port.go`
- `*_delete_port.go`
- other query/action ports like `transaction_navigation_get_port.go`

### Outbound Ports

Move infrastructure dependency contracts and port-adjacent shared types into
`internal/ports/outbound/` when they belong with outbound behavior or are needed
to keep the package buildable.

This includes files such as:

- `*_repository.go`
- `*_gateway.go`
- `*_store.go`
- `*_queue.go`
- `*_manager.go`
- `*_parser.go`
- `event_publisher.go`
- `actor_directory.go`
- `password_hasher.go`
- shared result and helper types that are consumed across package files if they
  must move with the outbound commit to keep compilation intact

### Inbound Adapters

Move inbound transport adapters into `internal/adapters/inbound/`.

This includes files such as:

- `*_http_adapter.go`
- `*_http_adapter_test.go`
- `http_common.go` when needed with inbound handlers

### Outbound Adapters

Move outbound infrastructure adapters into `internal/adapters/outbound/`.

This includes files such as:

- `*_postgres_repository.go`
- `*_postgres_repository_test.go`
- `*_gateway.go`
- `*_gateway_test.go`
- `*_store.go`
- `*_store_test.go`
- `*_queue.go`
- `*_queue_test.go`
- `*_parser.go`
- `*_parser_test.go`
- `*_pgx.go`
- `*_pgx_test.go`
- `password_hasher_bcrypt.go`
- `password_hasher_bcrypt_test.go`
- `pgvector.go` when required by outbound package compilation

## Shared Files and Commit Placement

Some files are not purely directional, for example:

- `internal/adapters/doc.go`
- `internal/adapters/http_common.go`
- `internal/adapters/pgvector.go`
- `internal/ports/doc.go`
- `internal/ports/*_result.go`
- `internal/ports/classification_job.go`
- related package-level tests that exercise shared package behavior

These files will not be forced into an artificial split. Instead, each shared
file will be placed in the commit where it is needed to keep that commit cleanly
buildable.

Each commit message will explain this explicitly, for example by noting that the
commit also includes shared package files required to keep the directional move
buildable.

## Implementation Sequence

### Commit 1: Inbound Refactor

- Create `internal/ports/inbound/`
- Create `internal/adapters/inbound/`
- Move all clearly inbound port files and inbound adapter files
- Move any shared package files required so the repository still builds after the
  first commit
- Update any affected imports or references caused by directory moves
- Verify the code still compiles and tests pass at this boundary

### Commit 2: Outbound Refactor

- Create `internal/ports/outbound/` if not already fully populated
- Create `internal/adapters/outbound/` if not already fully populated
- Move all remaining outbound port files and outbound adapter files
- Move remaining shared package files
- Update any affected imports or references caused by directory moves
- Verify the final repository still compiles and tests pass

## Import and Package Considerations

- Package declarations remain unchanged after moves
- Any imports from `internal/ports` or `internal/adapters` will need to use the
  new directory paths when referencing moved packages from other directories
- Files within the same package must still resolve package-local types correctly
- Tests must move with their corresponding production files when directory layout
  changes require it

## Verification Plan

Verification should happen at both commit boundaries and after the full refactor.

Minimum verification:

- `go test ./...`

If a narrower check is useful while staging a commit, it can be used during the
work, but completion requires the full test run.

## Risks and Mitigations

### Risk: package breakage during partial move

Because `ports` and `adapters` are flat packages today, moving only some files can
temporarily break imports or package-local symbol resolution.

Mitigation:

- allow shared files to move in whichever commit keeps the package buildable
- verify after each commit boundary instead of only at the end

### Risk: misclassifying shared files as directional

Some result, helper, or doc files are used by both sides.

Mitigation:

- classify by buildability and package coupling, not only file name
- document the reason directly in commit messages

### Risk: unintended behavior change

Even a structural refactor can accidentally alter imports or test coverage.

Mitigation:

- keep edits minimal and constrained to moves and reference updates
- run full test verification before completion

## Expected Outcome

After the refactor:

- `internal/ports` is organized into `inbound/` and `outbound/`
- `internal/adapters` is organized into `inbound/` and `outbound/`
- package names remain unchanged
- behavior remains unchanged
- git history shows two directional commits with explicit justification for any
  shared files included to keep each commit buildable
