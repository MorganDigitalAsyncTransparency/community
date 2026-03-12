# 3. Programming Languages

**Status:** Accepted
**Date:** 2026-03-12

## Context

The technology choices (ADR 0002) define a backend service with a React frontend and SQLite storage, but do not specify programming languages. The language choice affects contributor onboarding, library ecosystem, runtime characteristics, and how well the codebase works with AI-assisted development.

Key considerations:

- The backend is a long-running service that polls an API, processes data, and serves a REST API
- The frontend is a React single-page application
- Contributors will include AI-assisted developers, so language clarity and tooling matter
- The system runs on a small server with limited resources
- The core work is HTTP, JSON, and data transformation — not compute-heavy processing
- There is no mature Discourse client library in any language; the API client will be custom regardless
- The project starts documentation-heavy with no UI — the frontend comes later

## Alternatives Considered

### TypeScript for both backend and frontend

One language across the entire stack. The strongest choice for maximizing AI-assisted development: TypeScript has the most training data, the most examples online, and the best AI tooling support. Shared types across frontend and backend reduce duplication. However, Node.js is heavier at runtime, brings more dependencies and tooling overhead, and the async-heavy style can obscure control flow in a service that is fundamentally sequential: poll, process, store, serve. Without strict discipline, TypeScript codebases tend toward loose structure.

### Python backend with TypeScript frontend

Quick to prototype HTTP integrations and data transformations. However, it lacks static typing by default, risks looser structure as the codebase grows, and is not a natural fit for a project that prioritizes strict, clean architecture with many contributors.

### Clojure backend with TypeScript frontend

Strong for data transformation and event-driven logic. However, poor AI tooling support, a small contributor pool, and a steep learning curve make it a poor fit for a project that depends on AI-assisted contributions and low onboarding friction.

### Go backend with TypeScript frontend

Go compiles to a single binary with minimal runtime overhead. Its static typing, explicit error handling, and simple language design enforce the kind of strict, clear structure the project values. Go's module system maps naturally to the layered architecture. The cost is two languages in the project, but the backend and frontend are cleanly separated by an API boundary, so shared types offer little practical benefit. Go has less AI training data than TypeScript but is well-supported by modern AI coding tools.

## Decision

**Backend: Go.** The project prioritizes clear structure, explicit boundaries, and strict code over maximum AI convenience. Go's language design enforces these qualities: there is less room for accidental complexity, control flow is visible, and the module system encourages clean layering. The compiled binary simplifies deployment, and the standard library covers HTTP serving, JSON handling, and concurrency without heavy dependencies.

**Frontend: TypeScript with React.** The frontend is written in TypeScript. React is already decided in ADR 0002, and TypeScript is the standard choice for type-safe React applications. When the frontend is added, it will be the layer where TypeScript's AI-tooling advantage is fully leveraged.

The two languages are separated by the backend API boundary. No code or types are shared across that boundary.

## Consequences

**Positive:**

- Go binary deploys as a single file with no runtime dependencies
- Low memory and CPU footprint suits the small server constraint
- Go's strictness reduces the risk of structural drift as contributors add code
- Static typing in both languages catches errors at compile time
- Both languages are well-supported by AI-assisted development tools
- Go's simplicity keeps the backend readable for new contributors

**Negative:**

- Contributors need familiarity with two languages
- No shared type definitions across the API boundary; contracts must be kept in sync manually or via code generation
- Go has less AI training data than TypeScript, which may slow AI-assisted backend work slightly
- Go's error handling style is verbose compared to exception-based languages
- No mature Discourse client library for Go; the API integration layer must be built from scratch
