# System Context: Hasir Platform

This document provides a comprehensive overview of the Hasir platform codebase context for developers and AI assistants.

---

## 1. System Overview

Hasir is a self-hosted platform for managing Protocol Buffer (protobuf) schemas.

- **Schema Registry**: Version-controlled repository for protobuf definitions, with access control at organization and repository level (reader / author / owner roles).
- **Git-over-SSH**: Custom SSH server (`gliderlabs/ssh`, port `:2222`) supporting Git push/pull natively. Developers push `.proto` updates directly via standard Git CLI.
- **Git-over-HTTP**: Raw `/git/` HTTP handler for authenticated Git operations over HTTPS.
- **Automatic SDK Generation**: On every schema push, Hasir validates definitions via `buf` and automatically generates and packages client SDKs (TypeScript, Go, Java, C#, Rust, etc.) using a configurable worker-pool job queue.
- **Documentation Hosting**: Auto-generated API docs served under `/docs/` per repository.
- **Management Dashboard**: Next.js 16 Web UI for managing organizations, repositories, users, SSH public keys, and viewing generated SDKs and docs.
- **Landing Site**: Marketing site (`apps/landing`) deployed independently to Vercel.

---

## 2. Tech Stack

| Layer | Technology |
|---|---|
| **API** | Go 1.26, ConnectRPC (Connect + gRPC + gRPC-Web), `pgx/v5`, raw SQL migrations, `buf` tooling integration |
| **SSH Server** | `gliderlabs/ssh` + `golang.org/x/crypto/ssh` (port `:2222`) |
| **Frontend** | Next.js 16, React, TypeScript, Bun 1.3.14, Tailwind CSS v4, shadcn/ui + Radix primitives |
| **UI Package** | `packages/ui` — shared component library (60+ components) consumed by `apps/dashboard` |
| **Proto** | Buf CLI, protobuf schemas in `packages/proto/proto/` |
| **Observability** | OpenTelemetry (OTLP gRPC export), Zap structured logging |
| **Docker Deploy** | Nginx (reverse proxy, SSL/Let's Encrypt) + PostgreSQL + API + Dashboard containers |
| **Kubernetes** | Helm chart at `deploy/helm/charts/hasir` |
| **Monorepo** | Turborepo orchestration, Bun package manager |
| **CI/CD** | GitHub Actions (9 workflows) — CI, release, Docker, Helm, Proto, CodeQL, Changesets, Landing deploy |

---

## 3. Directory Structure

```text
.
├── .github/
│   └── workflows/          # 9 GitHub Actions workflows (ci, docker, helm-release, proto-release, changeset, codeql, deploy-landing, helm, proto)
├── .agents/                # AI agent skills and workflows for this repo
├── apps/
│   ├── api/                # Go 1.26 API + Git-over-SSH/HTTP service
│   │   ├── internal/       # Domain packages: user, registry, organization
│   │   ├── pkg/            # Reusable packages (see §4)
│   │   ├── migrations/     # 20 raw SQL migration pairs (up/down)
│   │   └── main.go         # Entrypoint: config → migrate → OTel → repos → queues → handlers → graceful shutdown
│   ├── dashboard/          # Next.js 16 management UI
│   │   ├── app/            # App Router pages
│   │   │   ├── (authenticated)/  # Protected routes: dashboard, organization/[id], repository/[repositoryId], profile, invite
│   │   │   ├── api/        # Next.js API routes (auth, docs)
│   │   │   ├── login/      # Authentication pages
│   │   │   ├── register/
│   │   │   ├── forgot-password/
│   │   │   └── reset-password/
│   │   ├── components/     # Dashboard-specific UI components
│   │   ├── lib/            # Client utilities (session, interceptors, ConnectRPC client, hooks)
│   │   └── stores/         # Zustand client state (registry-store)
│   └── landing/            # Next.js 16 marketing site (Vercel deploy)
├── deploy/
│   ├── docker/             # Docker Compose + Nginx templates + Certbot scripts
│   └── helm/               # Kubernetes Helm chart
├── docs/
│   ├── adr/                # Architecture Decision Records
│   └── plans/              # Dated design docs and implementation plans
├── openspec/               # OpenSpec change artifacts (specs/, changes/, archive/)
├── packages/
│   ├── proto/              # Buf config + protobuf definitions (organization, registry, shared, user services)
│   │   └── proto/          # Double-nested: packages/proto/proto/<service>/v1/
│   ├── ui/                 # Shared component library (shadcn/ui + Radix, 60+ components)
│   ├── shared/             # TS shared utilities (log, otel)
│   ├── eslint-config/      # Shared ESLint preset (Perfectionist sort-imports, line-length desc)
│   ├── typescript-config/  # Shared tsconfig.base
│   └── tooling/            # CLI bins: proto-gen.mjs, release.mjs
├── scripts/                # setup-sops.sh only
├── AGENTS.md               # AI developer guidelines, hot symbols, anti-patterns, commands
├── CONTEXT.md              # This file
├── DESIGN.md               # Design tokens and visual specifications
└── README.md
```

---

## 4. API Package Reference (`apps/api/pkg/`)

| Package | Purpose |
|---|---|
| `authentication` | JWT middleware (`AuthInterceptor`), `MustGetUserID` helper (~18 callers) |
| `authorization` | RBAC helper — `IsUserOwner` / `IsUserMember`, roles: `reader`, `author`, `owner` |
| `csrf` | CSRF protection interceptor (`CsrfInterceptor`) — validates `Origin` vs `DashboardUrl` / `PublicUrl` |
| `idempotency` | Idempotency interceptor — deduplicates requests keyed by `Idempotency-Key` header, backed by Postgres |
| `config` | App config struct — dev uses `config.json`, prod uses `HASIR_*` env vars |
| `sdkgenerator` | `buf` CLI orchestration for multi-language SDK generation |
| `postgres` | `pgx/v5`-based connection pool helpers |
| `email` | Email queue (10 workers) |
| `log` | Zap logger initialisation |
| `proto` | Proto utility helpers |

---

## 5. ConnectRPC Interceptor Chain

Every unary request passes through (in order):

1. **CSRF** — `csrf.CsrfInterceptor` (origin validation)
2. **Protovalidate** — `bufbuild/protovalidate-go` (schema-level field validation)
3. **Idempotency** — `idempotency.IdempotencyInterceptor` (dedup for `IdempotencyLevel == Idempotent`)
4. **Auth** — `authentication.AuthInterceptor` (JWT verification, injects `userId` into context)
5. **OTel** (optional) — `otelconnect` tracing, enabled when `cfg.Otel.Enabled`

---

## 6. Dashboard Client Architecture

| Module | Location | Role |
|---|---|---|
| Session management | `lib/session.ts` | Server-side JWT cookie handling |
| Session provider | `lib/session-provider.tsx` | React context wrapper for auth state |
| ConnectRPC client | `lib/use-client.ts` | Factory for typed Connect clients |
| Auth interceptor | `lib/auth-interceptor.ts` | Attaches JWT bearer token to outgoing requests |
| CSRF interceptor | `lib/csrf-interceptor.ts` | Reads CSRF token from cookie, adds to request header |
| Idempotency interceptor | `lib/idempotency-interceptor.ts` | Adds `Idempotency-Key` header for safe retries |
| Query retry | `lib/query-retry.ts` | TanStack Query retry policy for Connect errors |
| Registry store | `stores/registry-store.ts` | Zustand store for client-side registry state |
| Repository context | `lib/repository-context.tsx` | React context for selected repository |

---

## 7. Key Workflows & CLI Commands

Run from monorepo root unless workspace-specific execution is required:

```bash
make setup        # Install deps, lint/proto tools, set up config files
make dev          # Start Go API + Next.js dev servers concurrently
make build        # Build Go API + compile Next.js frontend
make test         # go test ./... + bun test across monorepo
make lint         # golangci-lint + eslint + buf lint
make typecheck    # tsc --noEmit for TypeScript apps
make proto        # buf generate on packages/proto/
make docker       # Build Docker images via bake
make vuln         # govulncheck vulnerability scan
make clean        # Remove build artefacts
make release      # Changeset interactive (or: app=api bump=patch)
make generate-mocks  # Regenerate Go testify mocks (apps/api)
```

---

## 8. Release Channels (three independent)

| Channel | Trigger | Workflow | Target |
|---|---|---|---|
| **Apps** | Changesets → tag `hasir-<app>@x.y.z` | `docker.yml` | GHCR (`ghcr.io/lynicis/hasir-{api,dashboard}`) |
| **Helm** | Tag `hasir-helm@*` | `helm-release.yml` | GitHub chart-releaser |
| **Proto** | Tag `proto/v*` | `proto-release.yml` | Buf Schema Registry (BSR) |

---

## 9. Coding & Contribution Rules

- **Type Safety**: TypeScript must be clean — strictly no `as any`, `@ts-ignore`, or `@ts-expect-error`.
- **Database Migrations**: Raw SQL only under `apps/api/migrations/`. Never ORM migrations. Never modify applied migration files.
- **Proto generation**: Never commit generated protobuf code (ADR-0003). Never commit generated `buf.yaml` / `buf.gen.yaml`.
- **Secrets**: Never commit plaintext secrets. Use SOPS for encrypted configs.
- **Git flow**: Trunk-based on `main` — always deployable. No long-lived release branches.
- **Dependencies**: Never add JS/TS dependencies to Go services.
- **Refactoring**: Before changing Go or TS symbols, check dependents with graphify or grep to avoid regressions.
