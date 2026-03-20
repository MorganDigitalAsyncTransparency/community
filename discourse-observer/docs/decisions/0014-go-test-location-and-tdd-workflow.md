# 0014. Go Test Location and TDD Workflow

**Status:** Proposed
**Date:** 2026-03-20

## Context

[ADR 0008](0008-documentation-and-traceability-strategy.md) decided that all tests — Go and frontend — live in a `tests/` directory that mirrors the spec structure. The rationale was that specs and tests are written before source files exist, so tests cannot live next to source files whose location is not yet known.

In practice, the project has diverged from this decision for Go:

- All 16 Go test files live in `backend/`, colocated with source files.
- The `tests/` directory contains only frontend tests (`tests/dashboard/`).
- The Go test runner (`scripts/test-go.sh`) runs `go test ./backend/...` — it does not reference `tests/` at all.
- The CI configuration, Makefile, and IDE tooling all assume Go tests are colocated.

The divergence is not accidental sloppiness — it reflects real constraints of the Go toolchain and a genuine gap in how ADR 0008 modeled TDD for a statically typed language.

### Go toolchain constraints

Go's testing model is tightly coupled to the package system:

1. **Internal test packages** (`package foo`) can access unexported symbols. Moving these tests outside the package directory breaks this access. The only workaround — exporting everything — degrades the API surface.
2. **`go test ./...`** discovers `_test.go` files by walking the module directory tree. Tests in `tests/` outside the Go module root require separate `go.mod` files or symlinks — both add complexity with no benefit.
3. **IDE integration** (gopls, VS Code Go extension) provides go-to-test, coverage overlay, and test-on-save based on colocated `_test.go` files. Tests in a separate directory lose all of these.

### TDD in a statically typed language

ADR 0008's TDD rationale ("tests are written before source files exist") works naturally in TypeScript where imports are resolved at runtime and test files can reference not-yet-existing modules. In Go, the compiler rejects imports of nonexistent packages — you cannot write `import "backend/observer"` before the `observer` package exists.

However, this does not mean Go cannot support test-before-code workflows. The approach is different:

1. **Define interfaces and domain types first** — a thin contracts file that compiles.
2. **Write acceptance tests against those interfaces** using fakes — tests express desired behavior before any implementation exists.
3. **Implement the concrete types** until the tests pass.

The tests live in external test packages (`package observer_test`) that only access the public API. They prove use cases, not implementation details. The key insight: TDD in Go is about **when** you write the test and **what** it tests (behavior via interfaces), not **where** the test file lives.

### The actual problem

ADR 0008 conflated two separate concerns:

1. **Test location** — where test files live on disk.
2. **Test ordering** — when tests are written relative to implementation.

The `tests/` directory was the proposed solution to both. For frontend, this works. For Go, the location constraint creates real toolchain friction while the ordering goal can be achieved independently through interface-first testing.

## Alternatives Considered

### A. Move Go tests to `tests/` (enforce ADR 0008 as written)

Move all Go test files from `backend/` to `tests/observer/`, `tests/api/`, `tests/storage/`, etc. Update `scripts/test-go.sh` and CI to discover tests in both locations.

**Pros:** Consistent with ADR 0008 and the documentation strategy. One rule for all languages.

**Cons:**

- 10 of 16 Go test files use internal test packages (`package storage`, `package domain`, `package api`) that access unexported symbols. Moving them breaks compilation. The fix — exporting internal types — weakens encapsulation across the entire backend.
- `go test ./backend/...` would no longer find any tests. The test script would need `go test ./backend/... ./tests/...` with a separate `go.mod` or module replacement directives.
- IDE features (go-to-test, coverage, test-on-save) stop working for tests in `tests/`.
- Creates a Go project structure that no Go developer would recognize, increasing onboarding friction.

### B. Colocated Go tests with acceptance-test discipline (recommended)

Go tests stay in `backend/` per Go convention. The TDD workflow is preserved through a different mechanism: acceptance tests written against interfaces in external test packages (`_test`), before implementation. The documentation strategy distinguishes Go and frontend test location conventions.

**Pros:**

- Aligns with Go convention — zero toolchain friction.
- Preserves TDD: acceptance tests are written in Phase 3, before implementation in Phase 4. They use interfaces and fakes, so they compile and fail before the concrete implementation exists.
- Tests prove use cases through behavior, not implementation details.
- Internal package tests remain available for implementation-level verification where valuable.
- No changes to existing test files, CI, or tooling.

**Cons:**

- Two different conventions for test location (Go in `backend/`, frontend in `tests/`). The documentation strategy must be explicit about which applies where.
- Spec-to-test traceability for Go relies on `// Spec:` header comments rather than directory mirroring.

### C. Hybrid — acceptance tests in `tests/`, unit tests colocated

Move only cross-package integration tests (`pipeline_test.go`, `sync_test.go`, `pipeline_report_test.go`) to `tests/observer/`. Keep package-internal tests colocated.

**Pros:** Partially satisfies ADR 0008. The tests that *can* move do move.

**Cons:**

- Two locations for Go tests creates confusion about where new tests go.
- The cross-package tests (`package main_test`) would need import path changes but no fundamental restructuring — so this is feasible but adds complexity for marginal benefit.
- Does not address the TDD workflow question — just shuffles files.

### D. Separate Go module for tests

Create a `tests/go.mod` that imports `backend/` as a dependency. Tests in `tests/` import backend packages through the module system.

**Pros:** Tests are physically in `tests/` while still compiling against backend packages.

**Cons:**

- Cannot access unexported symbols — same limitation as Alternative A for internal tests.
- Module dependency management becomes complex (replace directives, version synchronization).
- Unusual Go project structure with no ecosystem precedent.
- Significant engineering overhead for a documentation-organizational goal.

## Decision

*Awaiting decision.*

## Consequences

*To be completed after decision.*
