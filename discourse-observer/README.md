# discourse-observer

A generic foundation for observing and analyzing activity from a single Discourse forum.

## What this is

discourse-observer is a starter project for building tools that watch a Discourse forum, normalize its data, and prepare it for analysis, reporting, or visualization. It provides the structural foundation — architecture boundaries, contributor guidance, and module layout — so that implementation can proceed incrementally and clearly.

## Design principles

- **One forum per deployment.** This is not a multi-tenant platform. Each deployment targets a single Discourse forum. Different forums fork or adapt this project for their own needs.
- **Generic and reusable.** The project does not hardcode forum names, categories, team structures, or workflows. Forum-specific configuration is added during adaptation, not baked into the core.
- **Structure first.** The project starts with documentation, architecture decisions, and clean module boundaries before building features. This makes it easier for both human and AI contributors to extend the project without breaking its design.
- **Incremental growth.** The intended evolution is: Discourse API integration → observation/sync layer → normalized model → event/history store → backend/API → dashboard/reporting UI. Each layer is added when needed, not before.

## Why documentation comes first

This project is designed for AI-assisted contribution. That means:

- Module responsibilities must be explicit
- Architecture decisions must be recorded
- Boundaries between layers must be clear
- Conventions must be documented close to the code

Ambiguity in structure leads to inconsistent contributions. By establishing the foundation first, every future change has a clear place to land.

## Project structure

```text
discourse-observer/
  backend/
    discourse/    # Raw Discourse API integration
    observer/     # Turning source data into observed changes
    model/        # Internal normalized types and domain concepts
    config/       # Forum-specific configuration and adaptation
    storage/      # Abstraction for persisting observed data
  frontend/       # React dashboard UI
  specs/          # Behavior and model specifications
  docs/           # Purpose, context, and architecture decisions
  tests/          # Tests focused on transformation and observer logic
```

## Current status

This project is at the **foundation stage**. The structure, documentation, and architecture boundaries are in place. Implementation of the Discourse client, observer logic, and data model will follow as next steps.

## Getting started

```sh
make start
```

This copies `.env.example` to `.env` (if needed), builds and starts the containers, and opens the dashboard at <http://localhost:3000>. Edit `.env` with your Discourse credentials before the first run.

After code changes, use `make restart` to rebuild and relaunch.

See [docs/getting-started.md](docs/getting-started.md) for prerequisites, configuration details, and troubleshooting.

### Project orientation

1. Read [ARCHITECTURE.md](ARCHITECTURE.md) to understand the module boundaries.
2. Read [docs/purpose.md](docs/purpose.md) for project goals and direction.
3. Read [CONTRIBUTING.md](CONTRIBUTING.md) before making changes.
4. Check [docs/decisions/](docs/decisions/) for recorded architecture decisions.

## License

To be determined based on the adopting organization's needs.
