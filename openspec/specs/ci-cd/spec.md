# Spec: CI/CD Pipeline

> Derived from ADR-0008 (2026-07-14), ADR-0009 (2026-07-14), ARCHITECTURE.md §23 — Status: Accepted

## Purpose

Define CI/CD pipeline architecture, Turborepo remote caching via Vercel, git-filter-repo history preservation, and distinct release workflows.

## Requirements

### Requirement: Turborepo Remote Cache via Vercel

CI/CD pipelines and local developer machines SHALL share a Turborepo remote cache backed by Vercel Remote Cache. All cache operations SHALL be authenticated via `VERCEL_TOKEN` and integrity-checked via `TURBO_REMOTE_CACHE_SIGNATURE_KEY`.

#### Scenario: Cache hit in CI

- **WHEN** a pull request is opened for a commit whose task graph was previously computed
- **THEN** Turborepo restores all unchanged tasks from the remote cache
- **AND** CI wall-clock time for unchanged packages is under 30 seconds

#### Scenario: Cache invalidation

- **WHEN** a file in `packages/shared` is modified
- **THEN** all downstream packages that depend on `shared` are invalidated and rebuilt
- **AND** unchanged packages are still restored from cache

### Requirement: Git History Preserved via git-filter-repo

The one-time monorepo history import from five standalone repositories SHALL use `git-filter-repo` to rewrite paths and tags before merging. No legacy git tool (`filter-branch`, `subtree`) SHALL be used for this migration.

#### Scenario: History import validation

- **WHEN** the migration is complete
- **THEN** `git log --oneline apps/api/` shows commits originally from the `api` repository
- **AND** tags from the source `api` repo appear as `api/v*` in the monorepo

### Requirement: Separate CI and Release Workflows

- `ci.yml` SHALL run on every PR and push to `main`: lint, typecheck, test, proto lint, `buf breaking`.
- `docker.yml` SHALL run on `hasir-api@*.*.*` / `hasir-dashboard@*.*.*` tags: build and push Docker images to GHCR.
- `helm-release.yml` SHALL run on `hasir-helm@*` tags: package and publish the Helm chart.
- `proto-release.yml` SHALL run on `proto/v*` tags: push schemas to BSR.

#### Scenario: PR CI gates

- **WHEN** a pull request is opened
- **THEN** `ci.yml` runs lint, typecheck, and tests for all affected packages
- **AND** the PR cannot be merged until all checks pass

#### Scenario: Docker publish on release

- **WHEN** a `hasir-api@1.5.0` tag is pushed
- **THEN** `docker.yml` builds the API image and pushes `ghcr.io/lynicis/hasir-api:1.5.0` to GHCR
- **AND** the image is also tagged `latest`

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Remote cache | Vercel Remote Cache | Native Turborepo integration; zero maintenance |
| History import | `git-filter-repo` | Modern, fast, safe; supports path + tag rewriting |
| Container registry | GHCR | Native GitHub Actions integration; unified permissions |
| CI concurrency | Cancel in-progress runs on same branch | Avoid stale runs consuming minutes |

## Non-Goals

- Self-hosted CI runners
- Self-hosted S3-compatible remote cache
- Canary or blue-green deployment automation (Helm rollback is the rollback strategy)
