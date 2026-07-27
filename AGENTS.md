# PROJECT KNOWLEDGE BASE

**Regenerated:** 2026-07-27 (init-deep update)

## OVERVIEW
Hasir platform monorepo — self-hosted protobuf schema registry: Go API (ConnectRPC, PostgreSQL, JWT, Git-over-SSH) + Next.js dashboard + landing site + Buf protos + Helm/Docker deploys. Turborepo orchestrates; Bun 1.3.14 package manager.

## STRUCTURE
```
./
├── apps/
│   ├── api/           # Go 1.26 API + Git-over-SSH (see apps/api/AGENTS.md)
│   ├── dashboard/     # Next.js 16 management UI (see apps/dashboard/AGENTS.md)
│   └── landing/       # Next.js 16 marketing site (port 3001, Vercel deploy)
├── packages/
│   ├── proto/         # Protobuf schemas + buf (see packages/proto/AGENTS.md)
│   ├── shared/        # TS shared utils (log, otel)
│   ├── config/        # Shared eslint preset + tsconfig.base
│   └── tooling/       # CLI bins: proto-gen.mjs, release.mjs
├── deploy/
│   ├── helm/          # K8s chart (see deploy/helm/AGENTS.md)
│   └── docker/        # Compose + nginx (see deploy/docker/AGENTS.md)
├── docs/              # adr/ (9 ADRs), plans/ (dated design docs)
└── scripts/           # setup-sops.sh only
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| API domain logic | `apps/api/internal/{user,registry,organization}` | ConnectRPC services |
| SDK generation | `apps/api/pkg/sdkgenerator` | buf CLI orchestration |
| Dashboard pages | `apps/dashboard/app` | App Router, `(authenticated)` group |
| Proto schemas | `packages/proto/proto/**/v1/` | Double-nested path |
| CI/CD | `.github/workflows/` | 9 workflows, 3 release channels |
| Migrations | `apps/api/migrations` | 18 up/down pairs, raw SQL |

## CODE MAP
- Entry: `apps/api/main.go` — config → migrate up → OTel tracer → pg repos → email queue (10 workers) → registry service → SDK job queue → Connect handlers + raw `/git/` `/sdk/` `/docs/` → h1+h2c server → gliderlabs SSH (:2222) → graceful shutdown.
- Hot symbols: `authentication.MustGetUserID` (~18 callers), `registry.SshOperationRead/Write` (~11), `triggerSdkGenerationAfterPush` (duplicated in SSH + HTTP paths), `internal.GlobalHandler` route-registration iface.

## CONVENTIONS
- Turborepo orchestrates everything; Go services join via thin `package.json` wrappers.
- Single eslint preset in `packages/config` (Perfectionist sort-imports, line-length desc); strict TS (`noUncheckedIndexedAccess`, `verbatimModuleSyntax`).
- Raw SQL migrations only; buf for all proto work; changesets for per-app versioning (`hasir-api@x.y.z`).
- Config: dev = `apps/api/config.json` (`MODE=development`); prod = `HASIR_*` env vars; secrets via SOPS.

## ANTI-PATTERNS (THIS PROJECT)
- DO NOT use ORM-level migrations — raw SQL in `apps/api/migrations/`.
- DO NOT add JS dependencies to Go services.
- DO NOT use `@ts-ignore`, `@ts-expect-error`, or `as any`.
- DO NOT commit plaintext secrets or unencrypted configs (SOPS).
- DO NOT commit generated protobuf code (ADR-0003) or generated `buf.yaml`/`buf.gen.yaml`.
- DO NOT create long-lived release branches — `main` is always deployable.

## COMMANDS
```bash
make setup | dev | build | test | lint | typecheck | proto | docker | vuln | clean
make release            # changeset interactive (app=api bump=patch for non-interactive)
go test ./...           # API tests        bun test  # dashboard tests
make generate-mocks     # regenerate Go mocks (apps/api)
```

## RELEASE CHANNELS (three, independent)
- **Apps**: changesets → tag `hasir-<app>@x.y.z` → `docker.yml` → GHCR
- **Helm**: tag `hasir-helm@*` → `helm-release.yml` (chart-releaser)
- **Proto**: tag `proto/v*` → `proto-release.yml` → BSR push

## NOTES
- Image-name mismatch: bake/helm push `ghcr.io/lynicis/hasir-{api,dashboard}`; `deploy/docker` compose pulls `ghcr.io/lynicis/{api,dashboard}` (broken).
- `turbo.json` globalDependencies lists `CODEOWNERS` — file does not exist (stale ref).
- CI coverage steps are continue-on-error; tests do not gate merges.
- `apps/landing` deploys to Vercel on push to main (`apps/landing/**`).

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **hasir** (3528 symbols, 9410 relationships, 182 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

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
