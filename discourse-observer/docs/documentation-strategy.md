# Documentation and Traceability Strategy

This document defines how specifications, tests, and source code are organized and linked in discourse-observer. The approach was decided in [ADR 0008](decisions/0008-documentation-and-traceability-strategy.md).

---

## Directory structure

`specs/` and `tests/` mirror the module layout defined in [ARCHITECTURE.md](../ARCHITECTURE.md). `backend/` and `frontend/` follow the same module boundaries but are free to organize files as the implementation requires.

```text
specs/
  observer/
    change_detection.md
  dashboard/
    queue-visibility.md
    dashboard-components.md
  discourse/
  model/
  storage/
  config/
  use-cases.md                               ← overarching: what users need (not a spec)
  single-forum-scope.md                      ← system-wide constraint
  single-forum-scope_verification.md         ← manual verification for above
  operational-constraints.md                 ← system-wide constraint
  operational-constraints_verification.md    ← manual verification for above
tests/
  observer/
    change_detection_unit_test.go
    change_detection_integration_test.go
  dashboard/
    queue-visibility.unit.test.ts
  discourse/
  model/
  storage/
  config/
backend/
  observer/
    change_detector.go             ← free structure
    diff_calculator.go             ← free structure
    change_types.go                ← free structure
  discourse/
  model/
  storage/
  config/
frontend/
  src/
    components/
      SummaryCards.tsx              ← free structure
      UnrepliedTable.tsx           ← free structure
```

## Deviation from language-default test conventions

Go convention places test files next to source files in the same package. Frontend frameworks (Vitest, Jest) similarly default to colocated test files. This project deliberately separates tests into `tests/` to support the TDD workflow: specifications and tests are written *before* the source files exist. Placing tests next to source files assumes the source structure is known — in a spec-first workflow, it is not. The `tests/` directory mirrors the spec structure, not the source structure.

This means `go test ./...` from the project root does not automatically discover tests in `tests/`. Similarly, frontend test runners need explicit configuration to find tests outside `src/`. Test execution is configured through the Makefile, CI scripts, and test runner config files, which explicitly include the `tests/` directory.

---

## Verification requirements

Every specification — module-level or system-wide — must have at least one corresponding verification artifact. There are no exceptions.

- **Automated tests** are preferred: unit tests, integration tests, contract tests. These live in `tests/`.
- **Manual verification** is acceptable when automation is impractical. Manual verification is documented as a markdown file in `specs/` alongside the spec it verifies, following the naming convention `<spec>_verification.md`. The file describes the verification steps concretely enough that someone unfamiliar with the system can execute them.

Use cases (`specs/use-cases.md`) are not specs — they describe what users need from the system. Use cases drive the creation of module specs but are not subject to the verification requirement themselves. They are validated indirectly: each use case should be traceable to one or more module specs that *are* verified.

A spec without any verification artifact is an open gap. This is expected during development — specs, tests, and implementation may land in separate PRs as the TDD workflow progresses. The CI check reports gaps as information, not as merge blockers.

---

## Test file naming

Test files use the spec filename as a **prefix**, with a suffix indicating the test type. The suffix convention adapts to the language:

**Backend (Go):**

| Test type | Naming pattern | Example |
|-----------|---------------|---------|
| Unit test | `<spec>_unit_test.go` | `change_detection_unit_test.go` |
| Integration test | `<spec>_integration_test.go` | `change_detection_integration_test.go` |
| Contract test | `<spec>_contract_test.go` | `discourse_client_contract_test.go` |

**Frontend (TypeScript):**

| Test type | Naming pattern | Example |
|-----------|---------------|---------|
| Unit test | `<spec>.unit.test.ts` | `queue-visibility.unit.test.ts` |
| Integration test | `<spec>.integration.test.ts` | `queue-visibility.integration.test.ts` |

**Both:**

| Test type | Naming pattern | Example |
|-----------|---------------|---------|
| Manual verification | `<spec>_verification.md` | `single-forum-scope_verification.md` |

Automated test files live in `tests/`, regardless of language. Manual verification documents (markdown) live in `specs/` alongside the spec they verify. This keeps executable code separate from documentation while keeping manual verification close to the requirement it checks.

Multiple test files per spec is expected and correct — different test types verify different aspects of the same responsibility. A spec needs at least one.

---

## Two tiers of specifications

**Module specs** live in `specs/<module>/` and describe responsibilities within that module. Each has at least one corresponding test file in `tests/<module>/`.

**System specs** live in `specs/` root and describe cross-cutting constraints or system-wide properties (e.g., single-forum scope, operational constraints). These also require verification — typically manual verification documents or integration tests that assert the constraint holds.

**Use cases** live in `specs/` root as overarching documents that describe what users need from the system. Use cases are not specs — they drive the creation of module specs but are not subject to the verification requirement. Each use case should be traceable to the module specs it decomposes into.

---

## Spec scale and splitting

Individual spec files should stay readable and navigable. When a spec grows beyond approximately 200 requirements, it is a signal to refactor: identify distinct sub-responsibilities and split into separate spec files, each covering a smaller, cohesive area.

```text
specs/observer/
  change_detection.md              ← original, growing too large — refactor:
  change_detection_diffing.md      ← extracted sub-responsibility
  change_detection_categories.md   ← extracted sub-responsibility
```

The original file is replaced by the sub-files — no overview file is needed. Each sub-file gets its own test files and verification documents following the same prefix convention. This is a refactoring operation: the original spec, its tests, and its verification documents are removed, and new specs with new tests take their place. Requirement numbering restarts at 1 in each new sub-file.

---

## Requirement numbering within specs

Individual requirements within a spec file are numbered sequentially (1, 2, 3, ...). Numbers are assigned in order of creation and never reused. When a requirement is removed, its number is retired — gaps in the sequence are expected and indicate removed requirements. New requirements always receive the next number after the highest previously used, regardless of gaps.

This convention makes it possible to reference specific requirements (e.g., "change_detection requirement 14") in commit messages, PR discussions, and test names without ambiguity.

---

## Spec lifecycle

Specs are living documents that evolve with the project:

- **Creation:** A spec is written in Phase 1 before tests and implementation.
- **Update:** A spec may be updated in the same PR that implements it, when implementation reveals that the original requirements were incomplete or incorrect. The spec, tests, and code must be consistent within the PR.
- **Requirement removal:** When individual requirements become obsolete, they are deleted from the spec. Their numbers are retired — never reused.
- **Spec removal:** When an entire spec's responsibility is removed or superseded, the spec, its verification documents, and its corresponding test files are all deleted. The removal is done in its own PR with a commit message explaining what was removed and why. The git history preserves the decision trail — obsolete specs do not need to remain in the working tree.

---

## Traceability chain

Traceability works in two directions:

**Spec → tests** — linked by naming convention. The spec filename is the prefix for all related test files:

| Artifact | Location | Relationship |
|----------|----------|--------------|
| Specification | `specs/<module>/<responsibility>.md` | Defines the responsibility |
| Tests (Go) | `tests/<module>/<responsibility>_*_test.go` | Verify the specification |
| Tests (TS) | `tests/<module>/<responsibility>.*.test.ts` | Verify the specification |
| Manual verification | `specs/[<module>/]<responsibility>_verification.md` | Documents manual checks |

**Code → spec** — linked by header comment in each source file:

```go
// Package observer implements change detection for forum observations.
//
// Spec: specs/observer/change_detection.md
// Tests: tests/observer/change_detection_*_test.go
package observer
```

```ts
// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.*.test.ts
```

Every source file declares which spec it serves and which tests verify that spec. Multiple source files may reference the same spec — this is expected and correct when a responsibility is implemented across several files.

Header comments will drift over time — this is accepted as technical debt, not a design flaw. The traceability matrix generated by CI catches broken references. Reviews catch incorrect references. Periodic cleanup addresses accumulated drift.

---

## Automated verification

A CI check verifies the traceability chain:

- Every module spec in `specs/<module>/` has at least one corresponding test file in `tests/<module>/`
- Every system spec in `specs/` root has a corresponding manual verification document or integration test
- Every source file in `backend/` and `frontend/src/` contains a `Spec:` header comment pointing to a valid spec file
- Broken references (header comments pointing to non-existent specs) are reported as errors
- Orphaned specs (no test artifacts) are reported

This check runs on every PR. Gaps are reported as information, not as merge blockers — the TDD workflow means specs, tests, and implementation may arrive in separate PRs. The traceability matrix provides visibility into the current state of the chain, not a gate that prevents incremental progress.

---

## Workflow

The TDD workflow follows the delivery phases defined in CLAUDE.md:

1. Write the specification in `specs/<module>/<responsibility>.md` (Phase 1 — Requirements)
2. Document design decisions if needed (Phase 2 — Design)
3. Define validation strategy: determine what tests are needed and write failing tests in `tests/<module>/` (Phase 3 — Validation Strategy)
4. Implement in `backend/<module>/` or `frontend/src/` until tests pass — structure source files freely (Phase 4 — Implementation)
5. Add `Spec:` and `Tests:` header comments to each source file (Phase 4 — part of implementation)
6. Update the spec if implementation revealed gaps or corrections (Phase 4 — part of implementation)
7. Verify that spec, tests, and source references are consistent (Phase 5 — Review)

Each step may be its own PR. A spec can be merged without tests. Tests can be merged without implementation. The traceability matrix shows the current state of completeness — it does not enforce that the full chain exists in every PR.

The reverse is not equally acceptable: implementation code merged without a corresponding spec and tests weakens the traceability chain and creates unverified behavior. Code without specs should be flagged in review and treated as technical debt requiring prompt follow-up.

---

## Parallel work

The module-mirrored structure in `specs/` and `tests/` isolates parallel work streams. Each task involves a spec and its test files within one module. Two agents working on different responsibilities — even within the same module — touch different files.

Shared files (`ARCHITECTURE.md`, `mkdocs.yml` navigation, module-level `README.md` files) are updated in the same PR that introduces the change requiring the update — but only after the implementation is complete. If a task reveals a new architectural boundary or module change, the discovering agent documents it in its PR rather than deferring to a later merge. This keeps shared documentation current without requiring coordination between streams.

---

## Publishing

Documentation is published via **MkDocs** with the **Material** theme, deployed to **GitHub Pages**. MkDocs reads markdown files directly from the repository.

A `mkdocs.yml` at the project root defines the navigation structure. The navigation references files from `specs/`, `docs/`, and project-level markdown files.

How published pages are grouped — one page per spec file, aggregated per module, or a combination — is deliberately left open. MkDocs supports both direct file listing and include-based aggregation via plugins. The decision will be made when there are enough spec files to evaluate what reads best.

This setup is compatible with Backstage TechDocs, which uses MkDocs as its rendering engine. Migration to TechDocs requires adding a `catalog-info.yaml` and pointing it at the existing `mkdocs.yml`.
