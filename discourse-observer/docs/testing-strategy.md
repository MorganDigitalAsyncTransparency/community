# Testing Strategy

This document defines how tests are written in discourse-observer — test types, conventions, and the discipline behind each. For where tests live, how they are named, and how they link to specs, see [documentation-strategy.md](documentation-strategy.md).

The testing approach was shaped by [ADR 0014](decisions/0014-go-test-location-and-tdd-workflow.md), which established the interface-first acceptance testing workflow for Go.

---

## Principles

- **Test behavior, not implementation.** Assert on outputs and observable effects, not internal method calls or private state.
- **Fakes over mocks.** Write simple in-memory implementations of interfaces. No mocking frameworks. Fakes are production-quality code — they satisfy the same contract as real implementations, just without I/O.
- **Deterministic.** No time-dependent, order-dependent, or network-dependent tests. Inject time as a parameter when calculations depend on "now."
- **Readable as documentation.** A test should read like a specification example: given this input, expect this outcome. If a test needs a paragraph of comments to explain what it does, the test is too complicated.
- **Fast in CI.** No live network calls. All external data comes from fixtures or fakes. Target: full suite under 2 minutes.
- **Test the important paths first.** Data transformation, change detection, and domain calculations are the core value — test those heavily. Config parsing, logging, and CLI handling can wait.

---

## Go test types

Go tests fall into four categories. Each has a different purpose, package convention, and place in the delivery workflow.

### Acceptance tests

Acceptance tests prove that a feature satisfies its spec requirements. They are the primary verification artifact — every spec needs at least one.

**When written:** Phase 3 (Validation Strategy), before implementation exists.

**How they work:**

1. Define interfaces and domain types in the target package. These compile on their own — no implementation dependencies.
2. Write tests in an external test package (`package observer_test`) against those interfaces, using fakes.
3. Tests compile and fail. Implementation in Phase 4 makes them pass.

**Conventions:**

- **External package only.** Acceptance tests use `package foo_test`, never the internal package. They test through the public API — if a behavior cannot be observed through exported functions and types, it is not an acceptance-level concern.
- **Fakes, not concrete implementations.** Acceptance tests inject simple in-memory fakes that satisfy the interfaces. They never depend on SQLite, HTTP clients, or external services. This keeps them fast, deterministic, and focused on behavior.
- **One test per use case.** Each test proves one requirement or use case from the spec. The test name describes the use case in domain terms: `TestInitialSyncStoresAllPages`, not `TestRunInitialSyncWithFakeReturningThreePages`.
- **Table-driven for variants.** When the same behavior is tested with different inputs, use table-driven tests with descriptive subtest names.
- **Header comment.** Every acceptance test file starts with `// Spec: specs/<module>/<spec>.md`.

**Example:**

```go
// Spec: specs/observer/initial-delta-sync.md
package observer_test

func TestInitialSyncStoresAllPages(t *testing.T) {
    fake := &FakeFetcher{Pages: map[int][]model.Topic{
        1: {topic("a"), topic("b")},
        2: {topic("c")},
        3: {},
    }}
    store := NewInMemoryStore()

    result, err := observer.RunInitialSync(ctx, fake, store, 2)

    if err != nil { t.Fatal(err) }
    if result.TopicsStored != 3 { t.Errorf("got %d, want 3", result.TopicsStored) }
    if len(store.All()) != 3 { t.Errorf("store has %d, want 3", len(store.All())) }
}
```

### Internal tests

Internal tests verify implementation details within a package. They use the same package name as the code they test, giving access to unexported symbols.

**When written:** Phase 4 (Implementation), alongside or after the code.

**Conventions:**

- **Same package.** Internal tests use `package storage`, `package domain`, etc.
- **File naming:** `<spec>_unit_test.go` — the documentation strategy's "unit test" naming convention applies to these tests.
- **Valuable but optional.** Acceptance tests are the required verification artifact. Internal tests add confidence in tricky implementation logic — algorithms, edge cases in private helpers, state transitions.
- **May use concrete implementations.** Internal tests for `storage` may use a real SQLite database in a temp directory. Internal tests for `domain` test pure functions directly.

**Example:**

```go
package domain

func TestComputeQueueSummaryEmpty(t *testing.T) {
    now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
    result := ComputeQueueSummary(nil, now)
    if result.UnrepliedCount != 0 || result.UntaggedCount != 0 {
        t.Errorf("expected zeros for empty input")
    }
}
```

### Contract tests

Contract tests verify that API response shapes match the spec. They exercise real HTTP handlers with seeded data and assert on response structure and status codes.

**When written:** When an API endpoint is implemented or changed.

**Conventions:**

- **Internal package.** Contract tests typically use `package api` to access handler setup helpers.
- **Seeded data.** Use known test fixtures (from `mock.Topics()`) loaded into a temporary SQLite database. Assert on response shape, not exact values — the point is that the contract holds, not that a specific number appears.
- **Header comment.** Link to the API contract spec.

### Integration tests

Integration tests verify end-to-end workflows across module boundaries. They use real implementations (mock HTTP server, real SQLite) wired together.

**When written:** When the workflow being tested spans multiple modules and cannot be verified by acceptance or internal tests alone.

**Conventions:**

- **External package.** Integration tests typically use `package main_test` since they import from multiple internal packages.
- **Real implementations, controlled environment.** Use the mock Discourse server (`mockserver`), real SQLite in a temp directory, and the real observer. No fakes — the point is to verify that the real components work together.
- **Header comment.** Link to the spec that defines the workflow.

---

## Writing fakes

Fakes are simple in-memory implementations of interfaces. They satisfy the same contract as real implementations but without I/O, making tests fast and deterministic.

**Rules for fakes:**

- A fake implements the full interface. It does not panic on unimplemented methods — it returns zero values or errors as appropriate.
- A fake stores state in memory (slices, maps). It is inspectable — tests can read back what was stored without going through the interface if needed.
- Fakes live in the test file that uses them, or in a shared `_test.go` file within the same package if multiple test files need them. They do not live in production code.
- Keep fakes minimal. A fake that grows complex enough to need its own tests is a signal that the interface is too large.

**Example:**

```go
type FakeFetcher struct {
    Pages map[int][]model.Topic
}

func (f *FakeFetcher) FetchPage(ctx context.Context, page int) ([]model.Topic, error) {
    return f.Pages[page], nil
}

type InMemoryStore struct {
    topics []model.Topic
}

func (s *InMemoryStore) StoreTopics(_ context.Context, topics []model.Topic) error {
    s.topics = append(s.topics, topics...)
    return nil
}

func (s *InMemoryStore) All() []model.Topic { return s.topics }
```

---

## Frontend tests

Frontend tests live in `tests/dashboard/` mirroring the spec structure ([ADR 0008](decisions/0008-documentation-and-traceability-strategy.md)). They use Vitest and follow the same behavioral principles — test what the user sees, not component internals.

Frontend testing conventions are not further specified here because the frontend is a thin rendering client with no domain logic. As the frontend grows, this section should expand.

---

## What does not need heavy testing

- Config file parsing — validate manually until it stabilizes.
- Logging and formatting.
- CLI argument handling.
- Exact error message wording.

---

## Avoiding brittle tests

- Do not assert on exact JSON structures when only a few fields matter.
- Do not mock internals — fake at interface boundaries.
- Do not test private functions directly — test through the public interface.
- Do not assert on ordering unless ordering is a requirement.
- Do not couple tests to implementation structure — test the what, not the how.

---

## Test type summary

| Type | Package | Uses fakes | When written | Required | Proves |
|------|---------|-----------|-------------|----------|--------|
| Acceptance | External (`_test`) | Yes | Phase 3 | Yes | Spec requirements through behavior |
| Internal | Same as source | No (real impls OK) | Phase 4 | No | Implementation correctness |
| Contract | Internal | No (seeded data) | With API work | For API endpoints | Response shapes match spec |
| Integration | External (`main_test`) | No (real impls) | Cross-module work | Case by case | Components work together |
