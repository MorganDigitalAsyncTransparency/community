# AI Guidelines

Guidelines for AI tools contributing to this project, including code assistants, pair programming agents, and automated generation tools.

## Project context

discourse-observer is a generic starter project for observing a single Discourse forum. It is not multi-tenant. It is designed to be forked or adapted for specific forums. Keep this in mind when generating code or suggestions.

## Code generation principles

### Keep functions small and focused
Each function should do one thing. If you are generating a function that needs a comment to explain a section, split that section into its own function instead.

### Prefer pure transformations
When possible, write functions that take input and return output without side effects. Pure functions are easier to test, easier to reason about, and compose better. Reserve side effects (API calls, file writes, state mutation) for the edges of the system.

### Respect module boundaries
Each directory under `src/` has a defined responsibility:

- `src/discourse/` — Discourse API integration only
- `src/observer/` — Change detection and normalization only
- `src/model/` — Domain types only, no dependencies on other modules
- `src/config/` — Configuration and adaptation points only
- `src/storage/` — Persistence abstraction only

Do not mix responsibilities. Do not put Discourse API logic in the observer. Do not put observation logic in the model. If you are unsure where something belongs, check the module README files.

### Do not introduce frameworks without reason
Adding a framework is a significant decision. Do not introduce ORMs, web frameworks, dependency injection containers, or similar tools without a clear, documented reason. Simple standard library code is preferred until complexity demands otherwise. When a framework is warranted, record the decision as an ADR.

### Do not mix Discourse API concerns into unrelated modules
The Discourse API has its own data shapes, pagination patterns, and authentication requirements. All of this belongs in `src/discourse/`. Other modules should work with normalized internal types from `src/model/`, not raw API responses.

### Keep the project generic
Do not hardcode forum names, category IDs, tag names, team structures, or community-specific workflows. If forum-specific behavior is needed, it should go through `src/config/` and be documented.

## Documentation expectations

### Update documentation when changing architecture
If your change adds a module, changes a boundary, introduces a dependency, or alters the data flow, update the relevant documentation:

- Module READMEs in `src/`
- [ARCHITECTURE.md](ARCHITECTURE.md) if boundaries change
- A new ADR in `docs/decisions/` if a significant decision was made

### Write meaningful commit messages
Explain why the change was made, not just what changed. The diff shows the what.

## What to avoid

- Generating large amounts of speculative code that is not yet needed
- Adding error handling for scenarios that cannot occur
- Creating abstractions for patterns that appear only once
- Assuming multi-forum or multi-tenant requirements
- Introducing configuration complexity before the feature it configures exists
- Generating placeholder implementations that will need to be entirely rewritten
