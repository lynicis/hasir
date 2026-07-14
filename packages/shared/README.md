# path: packages/shared/README.md
# Stub package `@hasir/shared` — re-exported cross-cutting helpers.
# Currently: `log`, `otel`. The Go side has equivalents in `apps/api/pkg/log`.
# See ADR-0001 §13 (Baseline Production Hooks) for the convention.

This package is private and consumed only via the `workspace:` protocol by
`apps/*` workspaces in this repo. It is not published to a registry.
