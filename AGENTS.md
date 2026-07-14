# AGENTS.md

## Project Overview

Monorepo for the hasir platform — a Go API + Next.js dashboard + protocol buffer definitions + Helm deployment configs + Docker Compose stack.

### Structure

```
apps/api/          Go API service (Connect-RPC, PostgreSQL, JWT)
apps/dashboard/    Next.js dashboard (React, shadcn/ui, Bun)
proto/             Protocol buffer definitions (Buf)
deploy/helm/       Helm chart for Kubernetes deployment
docker/            Docker Compose stack (nginx, certbot, dev)
packages/          Shared configs (eslint, tsconfig, tooling)
scripts/           Build/release/lint utilities
docs/              Architecture docs and ADRs
```

## Commands

### API (Go)

```bash
cd apps/api && make dev       # MODE=development go run main.go
cd apps/api && make test      # go test ./...
cd apps/api && make build     # GOOS=linux go build
cd apps/api && make lint      # golangci-lint run
```

### Dashboard (Next.js)

```bash
cd apps/dashboard && bun dev           # Next.js dev server
cd apps/dashboard && bun test          # bun test
cd apps/dashboard && bun run build     # next build
cd apps/dashboard && bun run lint      # next lint
```

### Proto

```bash
cd proto && buf lint           # Lint proto files
cd proto && buf generate       # Generate code
```

### Helm

```bash
scripts/helm-lint.sh           # Lint Helm chart
```

### Docker

```bash
docker compose -f docker/docker-compose.yml up -d
```

### Monorepo

```bash
scripts/setup.sh               # First-time setup
scripts/clean.sh               # Clean all artifacts
make build                      # Build all apps
make test                       # Test all apps
```

## Git Workflow

- Trunk-based development on `main`
- Use `git filter-repo` for history rewrites
- Tag format: `<app>/<semver>` (e.g., `api/v1.0.0-alpha`)
