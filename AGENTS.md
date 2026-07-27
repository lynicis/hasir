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
