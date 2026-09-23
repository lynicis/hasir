# Spec: Release Strategy

> Derived from ADR-0005 (2026-07-14), ADR-0006 (2026-07-14) — Status: Accepted

## Purpose

Define the release strategy, independent per-app versioning using Changesets, and trunk-based development workflow.

## Requirements

### Requirement: Independent Per-App Versioning via Changesets

Each app (`hasir-api`, `hasir-dashboard`) SHALL be versioned independently using Changesets in `independent` mode. Git tags SHALL follow the Changesets native format: `hasir-api@<semver>` and `hasir-dashboard@<semver>`.

#### Scenario: Changeset creation

- **WHEN** a developer makes a releasable change and runs `bunx changeset`
- **THEN** a markdown file is created under `.changeset/`
- **AND** the file records the affected package(s) and bump type (major, minor, patch)

#### Scenario: Version Packages PR

- **WHEN** `.changeset/*.md` files are merged into `main`
- **THEN** the `changeset.yml` GitHub Actions workflow opens or updates a "Version Packages" PR
- **AND** that PR contains computed version bumps and `CHANGELOG.md` updates for each affected app

#### Scenario: Release publish

- **WHEN** a maintainer merges the "Version Packages" PR
- **THEN** Changesets creates git tags (e.g., `hasir-api@1.5.0`) and GitHub Releases automatically
- **AND** the new tags trigger `docker.yml` and `helm-release.yml` downstream workflows

### Requirement: Trunk-Based Development

All development SHALL occur on short-lived feature branches that merge directly to `main`. `main` SHALL always be deployable.

#### Scenario: Feature branch lifecycle

- **WHEN** a developer creates a feature branch
- **THEN** the branch exists for no more than a few days before merging via PR
- **AND** the PR passes all CI gates (lint, typecheck, build, test) before merge

#### Scenario: Main branch stability

- **WHEN** any commit is merged to `main`
- **THEN** the full CI suite (`ci.yml`) passes
- **AND** `main` remains in a deployable state at all times

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Versioning tool | Changesets (`@changesets/cli`) | Native monorepo-aware; independent versioning per package |
| Tag format | `hasir-<app>@<semver>` | Changesets native; downstream workflow triggers on glob `hasir-api@*.*.*` |
| Branching | Trunk-based development | Fastest feedback; no long-lived branches; continuous deployability |
| Release automation | `changesets/action` GitHub Action | Creates Version Packages PR, tags, and GitHub Releases automatically |

## Non-Goals

- Long-lived release branches or staging branches
- Manual version bumping in `package.json`
- Unified monorepo versioning (all apps share a single version)
