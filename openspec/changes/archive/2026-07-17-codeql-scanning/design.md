# Design: CodeQL Code Scanning

## Context

GitHub Actions code scanning via CodeQL provides SAST (Static Application Security Testing) for both Go and TypeScript. The main challenge is that the Go analyzer requires compiled types, including the protobuf-generated code which is never committed. Pre-generation must run before CodeQL's build step.

## Goals / Non-Goals

**Goals:**
- Enable GitHub code scanning with zero false-positive noise from generated code
- Analyze both `go` and `javascript-typescript` in a single matrix workflow

**Non-Goals:**
- Replacing `golangci-lint` or `gosec` (those remain in `ci.yml`)
- Enabling CodeQL auto-fix suggestions

## Decisions

### Decision 1: Matrix strategy for language coverage

Use a matrix with `language: [go, javascript-typescript]` so each language runs in isolation with appropriate build tooling.

### Decision 2: Buf generation before Go analysis

Add a pre-build step that installs Bun, Go, and Buf, then runs `make proto` so `packages/proto/gen/go` is populated before CodeQL initializes the Go build. Without this the Go analyzer fails to resolve generated types.

### Decision 3: Permissions

Set `security-events: write` and `actions: read` permissions so CodeQL can upload SARIF results to GitHub's security dashboard.

## PostgreSQL Migration

Not required.
