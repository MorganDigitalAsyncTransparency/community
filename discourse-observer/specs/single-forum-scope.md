# Single Forum Scope

This document explains the scoping decision that discourse-observer targets one Discourse forum per deployment and why this matters for the project's design.

## One forum per codebase

Each deployment of discourse-observer connects to exactly one Discourse forum. The configuration, data model, storage, and observation logic all assume a single forum as the data source.

This means:

- There is one set of API credentials
- There is one base URL for the Discourse instance
- There is one set of categories, tags, and topic namespaces
- There is no need for tenant isolation, per-forum routing, or cross-forum data separation

## Why not multi-tenant

Multi-tenant systems add significant complexity:

- Data isolation and access control between tenants
- Per-tenant configuration and credential management
- Routing and namespace disambiguation
- Shared infrastructure concerns (rate limiting per tenant, storage quotas)
- Deployment and upgrade coordination across tenants

None of this complexity serves the core goal of observing a single forum well. Adding it prematurely would slow down development and make the codebase harder to understand.

If a team needs to observe multiple forums, they deploy separate instances — one per forum. Each instance is independently configured, deployed, and maintained.

## Designed for forking and adaptation

The single-forum scope does not mean the project is only useful for one forum. It means:

- The **codebase** is generic and reusable as a starting point
- A team adapting it for their forum configures it for their specific Discourse instance
- Different teams can fork the project and diverge as needed
- Shared improvements can be contributed back to the upstream project

This is similar to how many open-source tools work: one codebase, deployed independently by each user, configured for their environment.

## Forum-specific workflows

Every Discourse forum develops its own workflows over time: specific categories for support requests, tags for priority levels, conventions for topic lifecycle, team-based routing patterns.

In discourse-observer, these workflows are not built into the core. Instead:

- The core observes generic Discourse entities (topics, categories, tags, revisions)
- Forum-specific interpretation happens through configuration in `backend/config/`
- If a workflow requires code changes (not just configuration), those changes should be documented and kept isolated so they do not break the generic foundation

This approach keeps the core reusable while allowing each deployment to layer on its own domain knowledge.

## When to reconsider

The single-forum scope should be reconsidered only if:

- A concrete, validated need arises for observing multiple forums in a single deployment
- The overhead of managing separate deployments becomes a demonstrated problem
- The forums share enough structure that a multi-forum model would be simpler than separate instances

Until then, the single-forum scope keeps the project simple, focused, and easy to reason about.
