# Spec: Protobuf Pipeline

> Derived from ADR-0002 (2026-07-14), ADR-0003 (2026-07-14), ADR-0004 (2026-07-14) — Status: Accepted

## Purpose

Define the Protobuf schema repository layout, local code generation policies (uncommitted generated code), and Go module import patterns.

## Requirements

### Requirement: Flat Proto Directory Layout

Protobuf schemas SHALL live under `packages/proto/proto/<domain>/v1/*.proto` (double-nested inside the package, single at the repo root level). No intermediate `hasir-proto/` directory nesting is permitted.

#### Scenario: Proto import paths

- **WHEN** a `.proto` file references another schema
- **THEN** the import path uses `<domain>/v1/<file>.proto` with no `hasir-proto/` prefix
- **AND** buf lint exits 0 with no warnings

### Requirement: Generated Code Never Committed

Generated code (`packages/proto/gen/`) SHALL never be committed to the repository. It is rebuilt on demand.

#### Scenario: Fresh clone build

- **WHEN** a developer clones the repository and runs `make setup`
- **THEN** `buf generate` runs and populates `packages/proto/gen/`
- **AND** no generated files are tracked by git

#### Scenario: CI build

- **WHEN** CI runs on a pull request
- **THEN** `make proto` runs before any compile or test step
- **AND** the generated files are available for the rest of the pipeline

### Requirement: Local Proto Generation for Go API

The `apps/api` Go module SHALL import generated Go code from `packages/proto/gen/go` via a `go.work` workspace or `replace` directive. It SHALL NOT import from the Buf Schema Registry (BSR) Go dependency.

#### Scenario: Local protobuf change

- **WHEN** a developer modifies a `.proto` file
- **THEN** running `make proto` regenerates the Go and TypeScript bindings locally
- **AND** `apps/api` picks up the change without publishing to BSR or installing a new package

#### Scenario: CI proto validation

- **WHEN** any file under `packages/proto/**` changes in a pull request
- **THEN** CI runs `buf lint` and `buf breaking --against .git#branch=main`
- **AND** the pipeline fails if either check reports an error

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Proto directory | `packages/proto/proto/<domain>/v1/` | Flat, idiomatic; no redundant nesting |
| Generated code | Never committed; `.gitignore`d | Avoids merge conflicts and repo bloat |
| Go import source | Local `packages/proto/gen/go` | Eliminates BSR dependency; self-contained builds |
| Breaking change detection | `buf breaking` vs `main` in CI | Prevents accidental wire incompatibilities |

## Non-Goals

- Publishing generated bindings to npm (TypeScript) — handled by the `@hasir/proto` package build
- Releasing proto schemas to BSR during development — BSR publish is a release-channel concern only
