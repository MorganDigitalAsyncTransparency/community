# 1. Project Foundation

**Status:** Accepted
**Date:** 2026-03-12

## Context

We want to build tools for observing and analyzing activity from a Discourse forum. There are several approaches we could take:

1. Start by building a full application and refactor later
2. Start with a specific forum's needs and generalize later
3. Start with project structure, documentation, and architecture boundaries first

Each approach has trade-offs. Starting with a full application risks building the wrong thing before the domain is understood. Starting with a specific forum's needs risks hardcoding assumptions that make the project difficult to reuse. Starting with structure risks over-planning, but for a project that will involve AI-assisted contributors, clarity of structure directly reduces the cost of every future contribution.

Additionally, this project is expected to be forked or adapted for different Discourse forums. A well-structured foundation makes that adaptation straightforward. A poorly structured one makes every fork diverge unnecessarily.

## Decision

We start with project structure, documentation, and architecture boundaries before writing application code.

The project is designed as a **generic starter for single-forum deployments**:

- **One forum per codebase/deployment.** This is not a multi-tenant system. Each deployment targets exactly one Discourse forum. This simplifies configuration, authentication, data modeling, and deployment.
- **Generic by default.** The project does not hardcode forum names, categories, tags, or workflows. Forum-specific adaptation happens through configuration or by forking the project.
- **Documentation as infrastructure.** Module READMEs, architecture decisions, contributor guidance, and AI guidelines are treated as first-class project artifacts, not afterthoughts. They reduce onboarding friction for every contributor.
- **Layered architecture from the start.** The source code is organized into clear layers (discourse, observer, model, config, storage) with defined responsibilities. This prevents the tangling of concerns that makes projects hard to extend.
- **Low-friction local development.** The project should be easy to clone, understand, and run locally without complex infrastructure. Platform and deployment decisions are deferred until the application layer warrants them.
- **Forum-specific assumptions come later.** When a team adapts this project for their forum, they add their specific categories, tags, workflows, and reporting needs on top of the generic foundation. These assumptions are documented when introduced.

## Consequences

**Positive:**
- Every future contribution has a clear place to land
- AI-assisted contributors can understand the project from documentation alone
- Different forums can fork from a clean, well-documented starting point
- Architectural mistakes are cheaper to fix before application code exists
- The single-forum scope keeps the system simple and avoids premature multi-tenancy complexity

**Negative:**
- The project does not do anything functional yet
- Contributors must read documentation before contributing (this is intentional)
- Some structural decisions may need revision once implementation reveals new constraints
- The layered structure adds some overhead for very simple use cases
