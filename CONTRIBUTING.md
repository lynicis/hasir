# Contributing to Hasir

Thanks for your interest in contributing! This document covers the practical
guidelines for working with this monorepo.

## Table of Contents

- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Database Migrations](#database-migrations)
- [Pull Requests](#pull-requests)
- [Issues](#issues)
- [Code of Conduct](#code-of-conduct)

## Project Structure

```
hasir/
├── apps/
│   ├── api/           # Go 1.26 API + Git-over-SSH (:8080, SSH :2222)
│   ├── dashboard/     # Next.js 16 management UI (:3000)
│   └── landing/       # Next.js 16 marketing site (:3001, Vercel deploy)
├── packages/
│   ├── proto/              # Protobuf schemas + buf workspace → @hasir/proto
│   ├── ui/                 # Shared component library (shadcn/ui + Radix)
│   ├── shared/             # TS shared utils (log, otel)
│   ├── tooling/            # CLI bins (proto-gen.mjs, release.mjs)
│   ├── eslint-config/      # Shared ESLint preset
│   └── typescript-config/  # Shared tsconfig.base
├── deploy/
│   ├── helm/          # Kubernetes Helm chart
│   └── docker/        # Docker Compose + nginx
├── docs/              # Legacy documentation & ADRs
├── openspec/          # OpenSpec standard: specs, changes, config.yaml
└── scripts/
```

Each app has its own `Makefile`, dependency file, and tests. The root
`Makefile` delegates to each via [Turborepo](https://turbo.build).

## Getting Started

### Prerequisites

- Go 1.26+
- Bun 1.4+
- Docker & Docker Compose (required for API integration tests and local Postgres)
- Helm (for deployment work)
- Buf CLI (for protobuf changes)

### Quick Setup

```bash
# 1. Install all dependencies and generate protobuf code
make setup          # bun install + buf generate + helm dep update

# 2. Start local Postgres (required for API)
cd apps/api && make run-postgres   # docker run -p 5432:5432 postgres:alpine

# 3. Edit dev config (committed with safe defaults)
# apps/api/config.json — set MODE=development; loaded automatically in dev

# 4. Start all apps in parallel
make dev            # turbo run dev --parallel (API :8080, Dashboard :3000)
```

## Development Workflow

1. **Branch**: Create a feature branch from `main` — no long-lived branches.
2. **Small commits**: Keep commits atomic and well-described using
   [Conventional Commits](https://www.conventionalcommits.org)
   (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, etc.).
3. **Scope changes**: A commit should ideally touch only one app unless the
   change is cross-cutting (e.g., shared proto or package changes).
4. **Regenerate after proto changes**:
   ```bash
   make proto   # buf generate → writes into packages/proto/gen/ (gitignored)
   ```
5. **Regenerate mocks after Go interface changes**:
   ```bash
   cd apps/api && make generate-mocks
   ```
6. **Check locally before pushing**:
   ```bash
   make lint       # ESLint (JS/TS) + golangci-lint (Go)
   make typecheck  # tsc --noEmit across all TS workspaces
   make test       # all tests (Docker required for integration tests)
   ```

## Code Style

### TypeScript / Next.js

- **Strict TS**: `noUncheckedIndexedAccess`, `verbatimModuleSyntax` — no
  `@ts-ignore`, `@ts-expect-error`, or `as any`.
- **Import order**: Perfectionist ESLint plugin, sorted by line-length
  descending. Use the `@/*` path alias inside `apps/dashboard`.
- **No inline styles**: Tailwind CSS utility classes only.
- **State management**: TanStack Query for server state; Zustand
  (`stores/registry-store.ts`) for invalidation counters only.
- **API calls**: Always use `useClient(Service)` from `lib/use-client.ts` —
  never fetch the API directly (except `use-documentation.ts` via `/api/docs`
  proxy).
- **Auto-fix**:
  ```bash
  cd apps/dashboard && bun run lint:fix
  ```

### Go

- **Linter**: `golangci-lint` via `go tool golangci-lint` (pinned in
  `go.mod` tool directives). Auto-fix with `cd apps/api && make lint-fix`.
- **Mocks**: `gomock` for service/handler mocks; `testcontainers` for
  repository-layer integration tests.
- **SQL**: Hand-built SQL only — no ORMs.
- **No cross-runtime deps**: Do not add Node.js/JS dependencies to the Go
  service.
- **Subprocess calls**: Do not use `exec.Command` outside the
  `CommandRunner` interface in `pkg/sdkgenerator`.

### Protobuf

- Schemas live in `packages/proto/proto/<domain>/v1/*.proto` (double-nested).
- **Never commit `gen/` output** — it is gitignored (ADR-0003) and rebuilt
  on demand.
- CI runs `buf lint` + `buf breaking` against `main` on every PR touching
  `packages/proto/**`.

## Testing

### API (Go)

```bash
# From monorepo root
make test               # turbo run test --affected

# From apps/api/ directly
go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...
```

Repository-layer tests use **testcontainers** (real Postgres — Docker
required). Service and handler tests use **gomock**.

### Dashboard (Bun)

```bash
# From monorepo root
turbo run test --filter=hasir-dashboard

# From apps/dashboard/ directly
bun test
bun test --watch
bun test --coverage
bun test <filename>   # e.g. bun test components/login-form
```

Test stack: **bun test + happy-dom + @testing-library/react**. No Vitest,
Jest, MSW, or Playwright. Tests are co-located with source (`*.test.tsx` /
`*.test.ts`). Always add or update tests for changed code.

### All tests

```bash
make test
```

## Database Migrations

- Migrations live in `apps/api/migrations/` as raw SQL pairs
  (`NNNNNN_name.{up,down}.sql`).
- Applied automatically at API startup via `golang-migrate`.
- **Never use an ORM** for migrations — raw SQL files only.
- `migrations_test.go` (testcontainers) is the canonical schema assertion.

## Pull Requests

1. PRs target `main` — it is always deployable.
2. **Title format**: `[scope] Brief description`
   (e.g. `[api] Fix JWT renewal race`, `[dashboard] Add org invite UI`).
3. Description should explain what the change does, why it's needed, and how
   it was tested.
4. Keep PRs focused — one logical change per PR.
5. **Required before opening a PR**:
   - `make lint` — must pass
   - `make typecheck` — must pass
   - `make test` — all tests green
   - `make proto` — if `.proto` files changed
   - `make generate-mocks` — if Go interfaces changed
6. Do not commit generated code (`packages/proto/gen/`) or unencrypted
   secrets (use SOPS).

### Review Process

- At least one approval is required before merging.
- First-time contributors: a maintainer will review within a few business
  days.
- Address review feedback with additional commits — we squash on merge.

## Issues

- **Bug reports**: Include reproduction steps, expected vs actual behavior,
  and environment details.
- **Feature requests**: Explain the use case and why it belongs in Hasir
  rather than as an external tool.
- **Security issues**: Do **not** file a public issue. See `SECURITY.md`.

## Common Gotchas

- **Proto double-nesting**: schemas are at `packages/proto/proto/<domain>/v1/`
  not `proto/<domain>/v1/`.
- **Turbo `--affected`**: relies on git diff vs the default base ref; in a
  fresh clone, run without `--affected` first.
- **testcontainers**: API integration tests require Docker running locally.
- **`config.json` secrets**: `apps/api/ssh_host_key` and `config.sops.json`
  are committed — treat them as sensitive.
- **SSH transport**: Git-over-SSH paths have zero test coverage.
- **`turbo.json` globalDependencies**: references `CODEOWNERS` which does not
  exist — ignore the warning.

## Code of Conduct

All contributors are expected to follow our
[Code of Conduct](CODE_OF_CONDUCT.md). Be respectful, constructive, and
professional.

## Questions?

Open a discussion or reach out to **me@lynicis.dev**.
