# Agent Instructions

## Package Manager
- Use **Bun** (`bun@1.4.4`): `bun install`
- Go workspaces (`go.work`): `apps/api` and `packages/proto`

## Commands
| Task | Command |
|------|---------|
| Bootstrap dependencies & codegen | `make setup` |
| Start dev servers (API + Dashboard) | `make dev` |
| Test single Dashboard file | `cd apps/dashboard && bun test path/to/file.test.tsx` |
| Test single Go package | `cd apps/api && go test -count=1 ./internal/registry/...` |
| Lint single TS/JS file | `bunx eslint path/to/file.tsx` |
| Lint single Go package | `cd apps/api && go tool golangci-lint run ./internal/registry/...` |
| Typecheck single TS workspace | `turbo run typecheck --filter=hasir-dashboard` |
| Regenerate Protobuf SDKs (`packages/proto/gen/`) | `make proto` |
| Regenerate Go mocks (`*_mock.go`) | `cd apps/api && make generate-mocks` |
| Full PR verification gate | `make lint && make typecheck && make test` |
| Update knowledge graph after code edits | `graphify update .` |

## External References
| Need | File |
|------|------|
| Setup, dev workflow & PR checklist | `CONTRIBUTING.md` |
| Domain model & architecture | `CONTEXT.md`, `docs/ARCHITECTURE.md` |
| Architecture Decision Records | `docs/adr/` |
| OpenSpec planning & quality gates | `openspec/config.yaml`, `openspec/specs/` |
| Protobuf service schemas | `packages/proto/proto/` |
| Design system & UI tokens | `DESIGN.md` |
| Release channels & runbooks | `docs/RELEASE.md` |
| Security policy | `SECURITY.md` |
| AI tooling & OpenSpec commands | `AI_USAGE.md` |
| Knowledge graph report | `graphify-out/GRAPH_REPORT.md` |

## Key Conventions
- Plan multi-file changes in `openspec/changes/<name>/` per `openspec/config.yaml`; never create ad-hoc plan docs.
- Query `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` first when `graphify-out/graph.json` exists; run `graphify update .` after code edits.
- Edit protobuf schemas under `packages/proto/proto/<domain>/v1/*.proto` and run `make proto`; never edit or commit `packages/proto/gen/`.
- Regenerate Go mocks with `cd apps/api && make generate-mocks` after changing Go interfaces; never hand-edit `*_mock.go`.
- Write database migrations as raw SQL pairs (`NNNNNN_name.{up,down}.sql`) in `apps/api/migrations/` and verify with `apps/api/migrations_test.go`; never use an ORM.
- Route external CLI execution in Go through `sdkgenerator.CommandRunner`; never call `exec.Command` directly.
- Call backend services in `apps/dashboard` via `useClient(Service)` from `apps/dashboard/lib/use-client.ts`; use `apps/dashboard/stores/registry-store.ts` for query invalidation counters only.
- Encrypt secrets with SOPS (`apps/api/config.sops.json`, `apps/api/ssh_host_key` are sensitive); never commit plaintext secrets or `.env` files.
- Create release changesets via `make release`; keep app (`hasir-<app>@*`), Helm (`hasir-helm@*`), and proto (`proto/v*`) release channels separate.

## Commit Attribution
AI commits MUST include:
```
Co-Authored-By: (the agent's name and attribution byline)
```
