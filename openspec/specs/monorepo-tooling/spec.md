# Spec: Monorepo Structure and Tooling

> Derived from ADR-0001 (2026-07-14) — Status: Accepted

## Purpose

Define the unified monorepo repository structure, build orchestration via Turborepo, package management with Bun workspaces, and Go module isolation.

## Requirements

### Requirement: Unified Monorepo Layout

The platform SHALL consolidate all apps, packages, deployment assets, and tooling into a single git repository under `apps/`, `packages/`, `deploy/`, and `scripts/`.

#### Scenario: Build orchestration

- **WHEN** a developer runs `make build` or `turbo run build`
- **THEN** Turborepo orchestrates task execution across all workspaces in dependency order
- **AND** only changed workspaces and their dependents are rebuilt (affected mode)

#### Scenario: Package management

- **WHEN** a developer runs `bun install` from the monorepo root
- **THEN** Bun workspaces links all local packages without manual symlinking
- **AND** the lock file is updated atomically for all workspaces

### Requirement: Go Module Isolation

Go services SHALL remain independent Go modules, participating in the Turborepo task graph via thin `package.json` wrappers that expose `build` and `test` scripts.

#### Scenario: Go build integration

- **WHEN** `turbo run build --filter=hasir-api` is run
- **THEN** Turborepo invokes the `build` script in `apps/api/package.json`
- **AND** that script delegates to `go build ./...`

### Requirement: Single Source of Truth

All code, configuration, CI/CD pipelines, and deployment assets SHALL live in this monorepo.

#### Scenario: Cross-cutting changes

- **WHEN** a change touches both `apps/api` and `packages/proto`
- **THEN** both are versioned and merged atomically in a single pull request
- **AND** CI validates the combined change end-to-end

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Task orchestrator | Turborepo | Best balance of speed, Go + Next.js compatibility, minimal config |
| Package manager | Bun workspaces | Fastest install times; aligns with Bun as JS runtime |
| Build isolation | Per-workspace Turbo tasks | Only rebuild affected packages |

## Non-Goals

- Replacing application business logic
- Selecting a cloud provider or Kubernetes distribution
- Changing the authentication/authorization architecture
