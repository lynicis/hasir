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

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **hasir** (3356 symbols, 9143 relationships, 174 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "main"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({search_query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/hasir/context` | Codebase overview, check index freshness |
| `gitnexus://repo/hasir/clusters` | All functional areas |
| `gitnexus://repo/hasir/processes` | All execution flows |
| `gitnexus://repo/hasir/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
