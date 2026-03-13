# Engineering Strategy

This document defines the engineering foundation for discourse-observer.
It covers structure, integration, testing, CI, and delivery discipline.

The goal is a simple, professional baseline — not a process framework.

---

## 1. Engineering approach

**Keep it simple. Make it work. Keep it clean.**

- Start with the smallest useful implementation and grow incrementally.
- Every module has one job. Every file has one reason to change.
- Core logic (model, observer) has zero external dependencies.
- Integrations (Discourse API, storage) sit behind interfaces.
- No framework until one is clearly needed. Standard library first.
- No abstraction without a second use case.

### Module dependency rule

```text
model         ← has no dependencies (leaf)
observer      ← depends on model
discourse     ← depends on model (implements interfaces defined in observer or model)
storage       ← depends on model (implements interfaces defined in observer or model)
config        ← depends on nothing; values injected into all modules at startup
```

- `model` depends on nothing. It is the innermost layer.
- `observer` depends on `model` only. It defines interfaces for fetching and storing data but does not import `discourse` or `storage` directly.
- `discourse` and `storage` implement interfaces defined by `observer` or `model`. They depend on `model` for types. At runtime, they are injected into the observer — the observer never imports them.
- `config` provides values at startup. Modules read config but do not depend on it structurally.

### File and function discipline

Follow CLAUDE.md rules. In short:

- Functions under 20 lines, files under 200 lines.
- No deep nesting. No boolean flag parameters.
- Names reveal intent.

---

## 2. Integration strategy

External systems (Discourse API, storage backends, future APIs) must be isolated.

### Rules

1. **Interface first.** Define what the core needs as a contract. Implement the adapter separately.
2. **No platform details in core logic.** The observer and model never import HTTP clients, database drivers, or API-specific types.
3. **Adapters are replaceable.** Swapping Discourse API versions, storage backends, or output formats should not touch core logic.
4. **Failures are explicit.** Adapters return typed results, not raw exceptions. The caller decides what to do with errors.
5. **Rate limits and retries live in the adapter**, not in business logic.

### Integration boundaries

| Boundary | Adapter responsibility | Core expectation |
|---|---|---|
| Discourse API | HTTP, auth, pagination, rate limits | Receives raw API data; normalizes internally |
| Storage | File/DB writes, schema, migrations | Receives and returns domain types |
| Output/Reporting | Formatting, delivery | Receives structured data |
| Configuration | File loading, env vars, validation | Provides typed config values |

### Partial data and failure

- Fetching may return incomplete data. The observer must handle partial results without crashing.
- Failed fetches should be logged and retryable. No silent data loss.
- The system should be resumable: a crash mid-sync should not corrupt state.

---

## 3. Testing strategy

### Layers

| Layer | What it tests | Speed | Priority |
|---|---|---|---|
| **Unit** | Model types, transformations, observer logic, change detection | Fast (<1s) | First |
| **Adapter** | Discourse client, storage adapters against recorded responses | Fast (<5s) | Second |
| **Contract** | API response shapes match expected schema | Fast | When API integration is built |
| **Smoke** | Key workflow: fetch → observe → store round-trip | Medium | When storage exists |

### Principles

- **Test behavior, not implementation.** Assert on outputs and side effects, not internal method calls.
- **Use fixtures and recorded responses.** No live API calls in CI.
- **Keep tests deterministic.** No time-dependent, order-dependent, or network-dependent tests.
- **Tests run in CI on every PR.** If tests are too slow for that, they are too slow.
- **Test the important paths first.** Data transformation and change detection are the core value — test those heavily. Configuration loading and logging can wait.

### What does NOT need heavy testing early

- Config file parsing (validate manually until it stabilizes).
- Logging and formatting.
- CLI argument handling.
- Exact error message wording.

### Avoiding brittle tests

- Do not assert on exact JSON structures when only a few fields matter.
- Do not mock internals; mock at boundaries (API client, storage).
- Do not test private functions directly. Test through the public interface.

---

## 4. CI strategy

### Workflows

All workflows use GitHub Actions. Keep them minimal and fast.

#### `ci.yml` — Pull request validation

Runs on every push to a PR branch and on `main`.

Jobs:

1. **lint** — Static analysis and formatting check.
2. **test** — Run all unit and adapter tests.
3. **build** — Verify the project compiles/builds cleanly.

All three must pass to merge. Each job runs independently (parallel).

#### `release.yml` — Tagged releases

Runs when a version tag (`v*`) is pushed.

Steps:

1. Run full test suite.
2. Build artifacts.
3. Create GitHub release with changelog.

#### `deps.yml` — Dependency updates (optional)

Weekly schedule. Dependabot or Renovate config to open PRs for dependency updates.
These PRs go through normal CI before merge.

### Branch protection

- `main` requires PR with passing CI.
- No direct pushes to `main`.
- At least one approval on PRs (when team > 1).
- Force-push to `main` is disabled.

### Keeping CI fast

- Target: full CI under 2 minutes.
- No live network calls. All external data is fixtures.
- Cache dependencies between runs.
- Only run affected tests if the project grows large enough to justify path filtering.

---

## 5. Delivery discipline

### How changes land

1. **One concern per PR.** A PR does one thing: a feature, a fix, a refactor, or a doc update. Not a mix.
2. **Requirements before code.** For anything non-trivial, write down what the change should achieve before implementing. This can be a PR description, a spec update, or an ADR.
3. **Small PRs.** Prefer multiple small PRs over one large one. Aim for under 300 lines changed.
4. **Docs stay current.** If a change affects architecture, behavior, or boundaries, update the relevant docs in the same PR.
5. **ADR for structural decisions.** Any decision that changes module boundaries, adds a dependency, or alters the architecture gets an ADR.

### Merge expectations

- CI passes.
- PR description explains what and why.
- Code follows CLAUDE.md and CONTRIBUTING.md.
- No unrelated changes smuggled in.

### Post-merge confidence

- `main` is always in a working state.
- If `main` breaks, fixing it is the top priority.
- Every merged PR should be independently revertable.

---

## 6. Decisions and deferrals

Some concerns have been decided. Others are deliberately deferred.

### Decided

- **Language/runtime choice** — Go backend, TypeScript frontend ([ADR 0003](decisions/0003-programming-languages.md)).
- **Storage** — Raw observations persisted as NDJSON files ([ADR 0005](decisions/0005-storage-format.md)); derived analytical data persisted in SQLite ([ADR 0006](decisions/0006-analytical-storage.md)).
- **Code quality tooling** — golangci-lint, markdownlint-cli, native git hooks ([ADR 0004](decisions/0004-code-quality-tooling.md)).

### Deliberately deferred

These are real concerns that do not need solutions yet:

- **Deployment strategy** — not relevant until there is something to deploy.
- **Monitoring and alerting** — not relevant until the system runs in production.
- **Multi-repo or monorepo decisions** — single repo until there is a reason to split.
- **Performance optimization** — correctness first, optimize when measured.

---

## Summary

The strategy is: start simple, keep boundaries clean, test the core, automate CI, and deliver in small pieces. Complexity is added only when needed, never preemptively.
