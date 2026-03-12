# 4. Code Quality Tooling

**Status:** Accepted
**Date:** 2026-03-12

## Context

The project has established coding standards (CLAUDE.md), an engineering strategy with CI expectations, and language choices (ADR 0003: Go backend, TypeScript frontend). There is no tooling yet to enforce these standards automatically.

Without automated enforcement, code quality depends entirely on contributor discipline. This is especially risky in a project designed for AI-assisted contributions, where automated guardrails prevent drift more reliably than instructions alone.

The tooling must cover two languages (Go and TypeScript), plus Markdown documentation. It must also work within the project's constraints: simple setup, minimal dependencies, and no heavy infrastructure.

## Alternatives Considered

### Go linting

**staticcheck** — A standalone Go linter focused on correctness. High-quality checks, but only runs one set of analyses. Does not aggregate multiple linters, so you need to run several tools separately to get comparable coverage.

**go vet + individual linters** — Running `go vet`, `errcheck`, `ineffassign`, etc. as separate commands. No configuration overhead per tool, but requires managing multiple invocations, and there is no unified way to configure or suppress findings across them.

**golangci-lint** — A meta-runner that aggregates 50+ Go linters behind one command and one configuration file. De facto standard in the Go ecosystem. Supports selective enabling, per-linter configuration, and caching. The trade-off is a larger binary and its own configuration surface area.

### Markdown linting

**remark-lint** — Part of the remark/unified ecosystem. Powerful and extensible via plugins, but requires a Node pipeline with multiple packages to configure. More complexity than needed for enforcing basic Markdown consistency.

**markdownlint-cli** — Standalone CLI with a single JSON config file. Covers all common Markdown style rules. Simple to install, simple to configure, no plugin ecosystem to manage.

### TypeScript linting (deferred until frontend exists)

**Biome** — Fast all-in-one linter and formatter for JS/TS. Written in Rust, very fast. Younger ecosystem with fewer rules and plugins than ESLint. Less community adoption and fewer integration examples.

**ESLint** — The established standard for JavaScript and TypeScript linting. Broad rule set, large plugin ecosystem, well-understood by contributors and AI tools. The v9+ flat config format simplifies configuration. The trade-off is slower execution compared to Biome.

### CSS linting (deferred until frontend exists)

**Stylelint** — The standard CSS linter. Extensible via shared configs (`stylelint-config-standard`). No real competitor for dedicated CSS linting. Biome does not yet support CSS (as of early 2026).

### HTML validation (deferred until frontend exists)

**html-validate** — Offline HTML validator with a recommended rule set. Checks accessibility (WCAG), semantics, and spec compliance without requiring a running browser. Configurable via a single JSON file.

**Nu HTML Checker (v.Nu)** — The W3C's official validator. Thorough spec compliance checking, but requires a Java runtime or a hosted service. Heavier to run locally and in CI.

### Git hooks

**Husky** — Popular Node-based hook manager. Adds a dependency and its own lifecycle (`prepare` script, `.husky/` directory). Works well in pure Node projects but adds a moving part for a problem git solves natively.

**lefthook** — Go-based hook manager. Requires a separate binary install. More powerful than needed for running a few lint commands.

**Native git hooks with `core.hooksPath`** — Git natively supports pointing hooks to a committed directory. No external dependencies. The hooks are visible, versionable shell scripts. The only setup step is a one-time git config command, which a Makefile target can automate.

## Decision

Use **standard, language-native linters** for each file type, enforced by **native git hooks** via `core.hooksPath`.

### Go (backend)

- **golangci-lint** as the lint runner. It aggregates multiple Go linters behind one command and is the de facto standard in the Go ecosystem.
- Configuration in `.golangci.yml` at the project root.
- Invoked via `make lint` or `go run` with the golangci-lint binary.

### TypeScript (frontend, deferred)

- **ESLint** with flat config (`eslint.config.js`). ESLint v9+ flat config, no legacy `.eslintrc` files.
- Deferred until the frontend exists. Configuration is added when there is TypeScript code to lint.

### CSS (frontend, deferred)

- **Stylelint** extending `stylelint-config-standard`.
- Deferred until the frontend exists. Configuration is added when there are CSS files to lint.

### HTML validation (frontend, deferred)

- **html-validate** for offline HTML validation against the spec and WCAG rules.
- Deferred until the frontend produces HTML output.

### Markdown (all documentation)

- **markdownlint-cli** for Markdown files. Relevant immediately — the project is documentation-heavy.
- Relaxed rules for line length (MD013), multiple H1s (MD025), ordered list style (MD029), inline HTML (MD033), and empty links (MD042).

### Git hooks

- Hooks live in `.githooks/` committed to the repository.
- `core.hooksPath` is set to `.githooks/` via a setup target (Makefile or npm prepare, depending on which toolchain is active).
- The pre-commit hook runs linters and tests. It grows as toolchains are added.
- No external hook managers (Husky, lefthook, pre-commit framework).

### CI integration

- CI runs the same lint commands as the pre-commit hook. No separate CI-only lint configuration.
- This keeps local and CI enforcement identical.

## Consequences

**Positive:**

- Automated enforcement catches issues before code reaches review
- Language-native tools follow each ecosystem's conventions and update paths
- Native git hooks require zero external dependencies beyond the linters themselves
- Hooks are plain shell scripts — easy to read, debug, and modify
- Markdown linting is immediately useful for the documentation-heavy foundation stage
- CI and local hooks run the same checks, preventing environment-specific surprises

**Negative:**

- Contributors must run a setup step once (`make setup` or equivalent) to activate hooks
- Two separate lint toolchains (Go and Node) to maintain as the project grows
- golangci-lint has its own configuration surface area to learn
- Deferred frontend tooling means those configs must be added later in a follow-up change
