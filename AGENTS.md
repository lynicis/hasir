# AGENTS.md

> Hasir — self-hosted protobuf schema registry. Go 1.26 API + Next.js 16 dashboard + protobuf schemas + Helm/Docker deploy. Turborepo build system, Bun 1.4 package manager.

Sub-project AGENTS.md files exist at `apps/api/AGENTS.md`, `apps/dashboard/AGENTS.md`, and `packages/proto/AGENTS.md`. The closest file takes precedence for any given location.

---

## Quick Start

```bash
# 1. Install all dependencies and generate protobuf code
make setup          # bun install + buf generate + helm dep update

# 2. Start local Postgres (required for API)
cd apps/api && make run-postgres   # docker run -p 5432:5432 postgres:alpine

# 3. Copy and edit dev config (already committed with safe defaults)
# apps/api/config.json — set MODE=development; loaded automatically in dev

# 4. Start all apps in parallel
make dev            # turbo run dev --parallel (API :8080, Dashboard :3000)
```

## Monorepo Structure

```
./
├── apps/
│   ├── api/           # Go 1.26 API + Git-over-SSH (:8080, SSH :2222)
│   ├── dashboard/     # Next.js 16 management UI (:3000)
│   └── landing/       # Next.js 16 marketing site (:3001, Vercel deploy)
├── packages/
│   ├── proto/             # Protobuf schemas + buf workspace → @hasir/proto
│   ├── ui/                # Shared component library (shadcn/ui + Radix)
│   ├── shared/            # TS shared utils (log, otel)
│   ├── tooling/           # CLI bins (proto-gen.mjs, release.mjs)
│   ├── eslint-config/     # Shared ESLint preset
│   └── typescript-config/ # Shared tsconfig.base
├── deploy/
│   ├── helm/          # Kubernetes Helm chart
│   └── docker/        # Docker Compose + nginx
├── openspec/          # OpenSpec standard: specs/, changes/ (active + archive), config.yaml
├── docs/              # Legacy documentation & ADRs
└── scripts/
```

## Build

```bash
make build              # affected packages only (turbo --affected)
bun run build           # all packages unconditionally

# Per-app from monorepo root
turbo run build --filter=hasir-api
turbo run build --filter=hasir-dashboard
```

## Testing

### API (Go)

```bash
# From monorepo root
make test               # turbo run test --affected

# From apps/api/ directly
go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...

# Testcontainers-based integration tests spin up a real Postgres — Docker required.
# Repository-layer tests use testcontainers; service/handler tests use gomock.
```

### Dashboard (Bun)

```bash
# From monorepo root
turbo run test --filter=hasir-dashboard

# From apps/dashboard/ directly
bun test
bun test --watch
bun test --coverage
bun test <filename>          # e.g. bun test components/login-form
```

Test stack: **bun test + happy-dom + @testing-library/react**. No Vitest, Jest, MSW, or Playwright.
Tests are co-located with source (`*.test.tsx` / `*.test.ts`). Always add or update tests for changed code.

### All tests

```bash
make test
```

## Linting & Type Checking

```bash
make lint               # ESLint (JS/TS) + golangci-lint (Go)
make typecheck          # tsc --noEmit across all TS workspaces

# Per scope
turbo run lint --filter=hasir-api
turbo run lint --filter=hasir-dashboard
turbo run typecheck --filter=hasir-dashboard

# Auto-fix (JS/TS only)
cd apps/dashboard && bun run lint:fix

# Go auto-fix
cd apps/api && make lint-fix
```

## Code Style

### TypeScript / Next.js

- **Strict TS**: `noUncheckedIndexedAccess`, `verbatimModuleSyntax` — no `@ts-ignore`, `@ts-expect-error`, or `as any`.
- **Import order**: Perfectionist ESLint plugin, sorted by line-length descending. `@/*` path alias for `apps/dashboard`.
- **No inline styles**: Tailwind CSS utility classes only.
- **State**: TanStack Query for server state; Zustand (`stores/registry-store.ts`) for invalidation counters only.
- **API calls**: Always use `useClient(Service)` from `lib/use-client.ts` — never fetch the API directly (except `use-documentation.ts` via `/api/docs` proxy).

### Go

- **golangci-lint** via `go tool golangci-lint` (pinned in `go.mod` tool directives).
- **gomock** for service/handler mocks; **testcontainers** for repository-layer integration tests.
- Regenerate mocks after interface changes: `cd apps/api && make generate-mocks`.
- Hand-built SQL only — no ORMs.
- Do not add Node.js/JS dependencies to the Go service.
- Do not use `exec.Command` outside the `CommandRunner` interface in `pkg/sdkgenerator`.

## Protobuf / buf

```bash
make proto              # buf generate → writes into packages/proto/gen/ (gitignored)
```

- Schemas live in `packages/proto/proto/<domain>/v1/*.proto` (double-nested path).
- **Never commit `gen/` output** — it is gitignored (ADR-0003) and rebuilt on demand.
- CI runs `buf lint` + `buf breaking` against `main` on every PR touching `packages/proto/**`.
- Releases publish to the Buf Schema Registry via the `proto/v*` tag.

## Generating Mocks (Go)

```bash
cd apps/api && make generate-mocks
```

Mocks are generated per domain (user, registry, organization, email, authorization) using `go tool mockgen`.

## Database Migrations

- Migrations live in `apps/api/migrations/` as raw SQL pairs (`NNNNNN_name.{up,down}.sql`).
- Applied automatically at API startup via `golang-migrate` (`file://migrations`).
- **Never use an ORM for migrations** — raw SQL files only.
- `migrations_test.go` (testcontainers) is the canonical schema assertion.

## Security Scanning

```bash
make vuln               # govulncheck (Go) + bun audit (JS) across all apps

# Go security scan (from apps/api)
make sec                # gosec (excludes sdk/ and repos/ dirs)
```

## Environment Variables

Dev: `apps/api/config.json` loaded when `MODE=development` (committed with safe defaults).
Prod: `HASIR_*` environment variables. All globals tracked in `turbo.json globalEnv`.

Key env vars:
| Variable | Purpose |
|---|---|
| `MODE` | `development` loads `config.json`; prod uses `HASIR_*` vars |
| `HASIR_POSTGRESQL_*` | Postgres connection (host, port, username, password, database) |
| `HASIR_JWTSECRET` | JWT signing key |
| `HASIR_SMTP_*` | SMTP settings (host, port, username, password, from, useTLS) |
| `HASIR_SERVER_PUBLICURL` | Public URL of the API server |
| `HASIR_SSH_*` | SSH server config (enabled, port, hostKeyPath) |
| `HASIR_DASHBOARDURL` | Dashboard origin (CORS / redirect) |
| `NEXT_PUBLIC_API_URL` | API URL for dashboard client-side Connect transport |
| `SESSION_SECRET` | iron-session cookie encryption key (dashboard) |
| `BUF_TOKEN` | Buf Schema Registry token for proto releases |

## Docker / Deployment

```bash
make docker             # buildx bake all images via deploy/docker/shared/docker-bake.hcl
```

Images: `ghcr.io/lynicis/hasir-api`, `ghcr.io/lynicis/hasir-dashboard`.

Local compose:
```bash
cd deploy/docker
cp config.example.json config.json   # fill in real values
docker compose up
```

Helm:
```bash
helm dependency update deploy/helm/charts/hasir
helm install hasir deploy/helm/charts/hasir -f your-values.yaml
```

## Releasing

Three independent release channels — do **not** mix them:

| Channel | Tag pattern | Workflow |
|---|---|---|
| Apps (API / Dashboard) | `hasir-<app>@x.y.z` | `docker.yml` → GHCR |
| Helm chart | `hasir-helm@*` | `helm-release.yml` (chart-releaser) |
| Protobuf schemas | `proto/v*` | `proto-release.yml` → BSR push |

```bash
# Interactive changeset (prompts for app + bump)
make release            # bunx changeset

# Non-interactive
app=api bump=patch make release
```

## OpenSpec Standard (Planning & Specifications)

Hasir follows the **OpenSpec standard** for all architecture planning, requirements, and change lifecycles.
Configuration and quality gates live in [`openspec/config.yaml`](openspec/config.yaml).

- **Specs as Single Source of Truth**: Core domain capabilities and architecture specifications live under `openspec/specs/<capability>/spec.md`. Always check existing specs before designing or modifying system behavior.
- **Never create ad-hoc plans**: Do not create standalone design docs or planning files (e.g. in `docs/plans/` or root). All planning must be an OpenSpec change.
- **OpenSpec Change Workflow**:
  - `openspec new change "<name>"` or `/openspec-propose <name>`: Create a new scoped change under `openspec/changes/<name>/`.
  - Artifact sequence: `proposal.md` → `specs/<capability>/spec.md` (deltas with WHEN/THEN/AND) → `design.md` → `tasks.md`.
  - `/openspec-apply-change <name>`: Work through tasks sequentially with verification.
  - `/openspec-verify-change <name>`: Verify implementation satisfies requirements.
  - `/openspec-archive-change <name>`: Move completed changes to `openspec/changes/archive/`.
- **Quality Gates (enforced before archiving)**:
  - Branching: Dedicated worktree (`git worktree add ../hasir-<change-name> -b <change-name>`).
  - Test coverage: >=85% for all touched packages (`go test -cover ./...` and `bun test --coverage`).
  - Protobuf: `make proto` + `buf lint` + `buf breaking` on schema edits.
  - Idempotency: Explicit tests for write operations.
  - Migrations: Raw SQL pairs in `apps/api/migrations/`, verified in `migrations_test.go`.
  - Changeset: Every user-facing PR ends with `bun changeset`.
  - Secrets: Zero plaintext secrets in Git (use Mozilla SOPS with `age`).

## PR Guidelines

- **Title format**: `[scope] Brief description` (e.g. `[api] Fix JWT renewal race`, `[dashboard] Add org invite UI`)
- **Required before opening a PR**:
  - `make lint` — must pass (ESLint + golangci-lint)
  - `make typecheck` — must pass
  - `make test` — all tests green
  - `make proto` — if `.proto` files changed
  - `make generate-mocks` — if Go interfaces changed
- **Commit messages**: Conventional Commits (`feat:`, `fix:`, `chore:`, etc.)
- `main` is always deployable — no long-lived feature branches.
- Do not commit generated code (`packages/proto/gen/`) or unencrypted secrets (use SOPS).

## Common Gotchas

- **Proto double-nesting**: schemas are at `packages/proto/proto/<domain>/v1/` not `proto/<domain>/v1/`.
- **Turbo `--affected`**: relies on git diff vs the default base ref; in a fresh clone, run without `--affected` first.
- **testcontainers**: API integration tests require Docker running locally.
- **`config.json` secrets**: `apps/api/ssh_host_key` and `config.sops.json` are committed — treat them as sensitive.
- **Dashboard duplicate file**: `lib/use-client.impl.ts` is byte-identical to `use-client.ts` — edit both or reconcile.
- **SSH transport**: Git-over-SSH paths have zero test coverage.
- **`turbo.json` globalDependencies**: references `CODEOWNERS` which does not exist — ignore the warning.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
