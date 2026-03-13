# 7. Branching and Release Strategy

**Status:** Proposed
**Date:** 2026-03-13

## Context

The repository is hosted on GitHub (`MorganDigitalAsyncTransparency/community`) with `main` as the default branch. It is maintained by a single developer with AI assistance (Claude Code).

The project is in foundation stage: specifications, documentation, and architecture decisions exist, but no application code has been written yet. The delivery workflow in CLAUDE.md (phases 0–8) already prescribes feature branches, rebase onto main, and pull requests — but does not specify branch protection, release mechanics, or CI expectations.

Future deployment is a Docker container on a small server (ADR 0002). That deployment does not exist yet, and adding infrastructure before there is code to deploy would be premature. The strategy defined here must work for the current documentation-only phase and scale to cover application code, automated testing, and deployment without requiring a rewrite.

This ADR applies to the entire repository, not only to discourse-observer. It is the first repository-level decision record, stored in `docs/decisions/` at the repository root.

## Alternatives Considered

### Git Flow (develop + release branches)

A branching model with long-lived `develop` and `release/*` branches, hotfix branches, and tagged releases from `release/*`.

Designed for projects with parallel release streams and large teams. The overhead of maintaining `develop`, creating release branches, and merging back hotfixes is unjustified for a single developer. The model adds ceremony without benefit at this scale.

### Trunk-based development (commit directly to main)

All changes go directly to `main` without branches or pull requests.

Fast, but sacrifices the audit trail that pull requests provide. The CLAUDE.md delivery workflow already depends on PRs for Phase 7–8. Direct commits also prevent CI from validating changes before they reach `main`.

### GitHub Flow (short-lived feature branches + PRs to main)

One long-lived branch (`main`). All work happens on short-lived feature branches that merge into `main` via pull request. `main` is always the latest stable state.

Simple, well-suited to a single developer, and already aligned with the CLAUDE.md delivery workflow. Pull requests create a natural review checkpoint and CI trigger point. No parallel release streams to manage.

## Decision

Adopt **GitHub Flow** as the branching and release model for the repository.

### Branching

- **`main` is the single long-lived branch.** It represents the current accepted state of the project.
- **All changes arrive via short-lived feature branches** merged through pull requests, as prescribed by the CLAUDE.md delivery workflow (Phase 7).
- **Branch naming:** `<type>/<short-description>` where type is one of `feature`, `fix`, `docs`, `refactor`, `chore`. Examples: `docs/adr-0007-branching`, `feature/discourse-polling`, `fix/revision-dedup`.
- **Branches are rebased onto `main` before merging** (Phase 7). Merge commits are acceptable when rebase would rewrite shared history, but rebase is the default.
- **Branches are deleted after merge** (Phase 8).

### Branch protection

Enable the following rules on `main`:

- **Require pull request before merging.** No direct pushes.
- **Require status checks to pass.** Once CI workflows exist, they must pass before merge.
- **Do not require reviews.** A single developer cannot meaningfully approve their own PRs. The CLAUDE.md Phase 5–6 review process substitutes for formal code review.
- **Allow force push: no.** Protect commit history on `main`.

### Release strategy

Releases are phased to match the project's maturity:

**Foundation stage (now):** No releases. `main` is the latest state. Changes are tracked through merged pull requests and ADRs. No tags, no version numbers.

**First application code:** Introduce semantic versioning (`MAJOR.MINOR.PATCH`) and tag releases on `main`. A release is created by tagging a commit and generating a GitHub Release with release notes. Tags follow the format `v0.1.0`. Start at `0.x` to signal pre-stable status.

**Server deployment (ADR 0002):** Adopt continuous deployment from `main`. Every merge to `main` that passes CI triggers a build and deploy pipeline. The tagged release process remains available for marking significant milestones but is not required for every deployment.

### CI workflows (GitHub Actions)

Workflows are introduced incrementally as the codebase warrants them:

**Now (foundation stage):**

- **Markdown lint** — validate documentation formatting on PRs. Lightweight and immediately useful since the project is documentation-heavy.

**When application code is added:**

- **Lint and type check** — run the tooling defined in ADR 0004.
- **Test** — run the automated test suite.
- **Build** — verify the Docker image builds successfully.

**When server deployment begins:**

- **Deploy** — build and push a Docker image, then deploy to the target server. Triggered on merge to `main` after all checks pass.

Each workflow runs on pull requests targeting `main` and on pushes to `main`.

### Environments

**Now:** Local development only. No remote environments.

**When deployment begins:** Two environments:

- **Local** — developer workstation, used for development and manual testing.
- **Production** — Docker container on the target server (ADR 0002). Deployed from `main`.

**If needed later:** A **staging** environment can be added between local and production. This would be a second container on the same server (or a separate one) running a build from a PR branch or a release candidate tag. This is not built until there is a concrete need — the single-developer workflow and CI checks provide sufficient confidence for now.

## Consequences

**Positive:**

- Aligns with the existing CLAUDE.md delivery workflow without changes
- Simple model with minimal overhead for a single developer
- Pull requests create an audit trail of every change and its rationale
- Branch protection prevents accidental direct pushes to `main`
- CI grows incrementally — no upfront infrastructure investment
- Release strategy scales from documentation-only to continuous deployment without model changes
- Staging can be added later without altering the branching model

**Negative:**

- No formal review gate beyond self-review (acceptable for single developer, would need revisiting if the team grows)
- No staging environment initially — issues may only surface in production when deployment begins
- Continuous deployment from `main` means a broken merge can reach production; CI quality directly determines production stability
- Semantic versioning requires discipline to tag releases at meaningful points rather than every merge
