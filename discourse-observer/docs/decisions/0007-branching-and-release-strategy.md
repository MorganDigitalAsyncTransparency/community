# 7. Branching and Release Strategy

**Status:** Proposed
**Date:** 2026-03-13

## Context

The repository is hosted on GitHub with `main` as the default branch. It is maintained with AI assistance (Claude Code).

The project is in foundation stage: specifications, documentation, and architecture decisions exist, but no application code has been written yet. The established delivery workflow already prescribes feature branches, rebase onto main, and pull requests — but does not specify branch protection, release mechanics, or CI expectations.

Future deployment is a Docker container on a small server (ADR 0002). That deployment does not exist yet, and adding infrastructure before there is code to deploy would be premature. The strategy defined here must work for the current documentation-only phase and scale to cover application code, automated testing, and deployment without requiring a rewrite.

## Alternatives Considered

### Long-lived develop and release branches (Git Flow)

A branching model with long-lived `develop` and `release/*` branches, hotfix branches, and tagged releases from `release/*`.

Designed for projects with parallel release streams and large teams. The overhead of maintaining `develop`, creating release branches, and merging back hotfixes is unjustified for a single developer. The model adds ceremony without benefit at this scale.

### Trunk-based development (commit directly to main)

All changes go directly to `main` without branches or pull requests.

Fast, but sacrifices the audit trail that pull requests provide. The established delivery workflow already depends on PRs. Direct commits also prevent CI from validating changes before they reach `main`.

### PR branch deploy to QA (deploy-before-merge)

Deploy the feature branch to a QA environment before merging to `main`. The branch is tested in QA, then merged if acceptable or closed if not. Production deploys automatically on merge to `main`.

This gives higher confidence than CI alone because the actual deployment is tested before it reaches `main`. However, it requires a QA environment, a deploy workflow with branch input, and a convention for which branch currently owns the QA slot. For a single developer in foundation stage this is premature — CI checks provide sufficient confidence.

**Deferred.** Expected to become relevant once application code is running and deployable. The branching model does not need to change to support this — it only requires a new CI workflow (`deploy-qa.yml` with `workflow_dispatch` and branch input) and a QA environment.

### Short-lived feature branches with PRs (GitHub Flow) — selected

One long-lived branch (`main`). All work happens on short-lived feature branches that merge into `main` via pull request. `main` is always the latest stable state.

Simple, well-suited to a single developer, and already aligned with the established delivery workflow. Pull requests create a natural review checkpoint and CI trigger point. No parallel release streams to manage.

## Decision

Adopt **short-lived feature branches with PRs (GitHub Flow)** as the branching and release model for the repository.

```text
feature/x ── PR against main
                  │
            CI checks (must pass)
                  │
            merge to main
                  │
            main updated
                  │
         ┌────────┴────────┐
    foundation stage     after deployment
         │                    │
    nothing more         auto-deploy to production
```

Hotfixes follow the same path — a short-lived branch with a PR. No special process is needed because every merge to `main` is a potential deploy.

### Branching

- **`main` is the single long-lived branch.** It represents the current accepted state of the project. No commits are made directly to `main`.
- **All changes arrive via short-lived feature branches** created from `main` and merged back through pull requests.
- **Branch naming:** `<type>/<short-description>` where type is one of `feature`, `fix`, `docs`, `refactor`, `chore`. Examples: `docs/adr-0007-branching`, `feature/discourse-polling`, `fix/revision-dedup`.
- **Branches are rebased onto `main` before merging.** Merge commits are acceptable when rebase would rewrite shared history, but rebase is the default.
- **Merge method: merge commit.** After rebasing, the PR is merged with a merge commit (GitHub's "Create a merge commit" option). This preserves individual commits from the branch while creating a clear merge point in the history. Squash merge is acceptable for single-commit branches but is not the default.
- **Verify locally before pushing.** Run available linting and tests before creating the PR. CI is the final gate, not the first.
- **Branches are deleted after merge.**

### Branch protection

Enable the following rules on `main`:

- **Require pull request before merging.** No direct pushes.
- **Require status checks to pass.** Once CI workflows exist, they must pass before merge.
- **Do not require reviews.** A single developer cannot meaningfully approve their own PRs. Self-review before merge substitutes for formal code review.
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

**If needed later:** A staging environment can be added by deploying PR branches to a separate container before merge. See the deferred "PR branch deploy to QA" alternative above.

## Consequences

**Positive:**

- Aligns with the existing delivery workflow without changes
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

### Required changes

| What | Change |
|------|--------|
| GitHub repo settings | Enable branch protection on `main`: require PR, require status checks, disallow force push |
| `.github/workflows/markdown-lint.yml` | Add workflow to lint Markdown on PRs targeting `main` |
