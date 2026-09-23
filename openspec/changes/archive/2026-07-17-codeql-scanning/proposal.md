# Proposal: CodeQL Code Scanning

## Why

The codebase had no automated static analysis for security vulnerabilities beyond `golangci-lint`. Adding CodeQL via GitHub's code scanning gives us SAST coverage for both Go and TypeScript/JavaScript, surfacing CVEs and code patterns that linters miss.

## What Changes

- New `.github/workflows/codeql.yml` workflow configuring GitHub code scanning for `go` and `javascript-typescript` languages
- Buf proto generation step added as a pre-build setup so the Go analyzer has access to generated types

## Capabilities

### New Capabilities
- `ci-cd`: CodeQL scanning integrated into the security posture of every PR and push to `main`

## Impact

- `.github/workflows/codeql.yml`: New file

## Implementation Branch

`codeql-scanning` (archived — completed)
