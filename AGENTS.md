# AGENTS.md

## Project Overview

Monorepo for the Hasir platform — a Go API + Next.js dashboard + protocol buffer definitions + Helm deployment configs + Docker Compose stack.

### Architecture & Quirks

- **Turborepo Orchestration**: All builds, tests, and linting are orchestrated by Turborepo from the root. 
- **Go via Turborepo**: Go services (`apps/api/`) remain standard Go modules but use a minimal `package.json` to participate in the Turborepo task graph. Do not add JS dependencies to Go services.
- **Protobufs**: The schema directory is flattened at `proto/` directly. 
- **Database Migrations**: Always use raw SQL migration files under `apps/api/migrations/`. Do not use ORM-level migrations to prevent schema drift.
- **TypeScript Strictness**: Strictly avoid `as any`, `@ts-ignore`, or `@ts-expect-error`.

### Structure

```text
apps/api/          Go API service (Connect-RPC, PostgreSQL, JWT, Git-over-SSH)
apps/dashboard/    Next.js dashboard (React, shadcn/ui, Bun)
proto/             Protocol buffer definitions (Buf)
deploy/helm/       Helm chart for Kubernetes deployment
docker/            Docker Compose stack (nginx, certbot, dev)
packages/          Shared configs (eslint, tsconfig, tooling)
scripts/           Build/release/lint utilities
docs/              Architecture docs and ADRs
```

## Commands

Run all primary commands from the **monorepo root**. Turborepo will orchestrate the underlying dependencies.

### Local Development & Verification

```bash
make setup                # Installs dependencies, lint/proto tools, setups config files
make dev                  # Starts Go API and Next.js dev servers concurrently
make build                # Builds Go API and compiles Next.js frontend
make test                 # Runs go test and bun test across the monorepo
make lint                 # Runs golangci-lint, eslint, and buf lint
make proto                # Triggers 'buf generate' on the proto/ directory
```

### Docker Deployments

```bash
docker compose -f docker/docker-compose.yml up -d
```

## Git Workflow & Releases

- **Trunk-based development** on the `main` branch. 
- **Independent Versioning**: Services are versioned independently. 
- **Tag format**: `<app>/<semver>` (e.g., `api/v1.0.0-alpha`)

### Releasing a Service

To release a service, use the automated bump script which handles git tags and triggers GitHub Actions:

```bash
bun run scripts/release.sh <app> <patch|minor|major>
# Example: bun run scripts/release.sh api minor
```

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **hasir** (3507 symbols, 9352 relationships, 179 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

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
