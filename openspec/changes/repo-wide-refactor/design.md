# Design: Repo-Wide Structural Refactoring

## Migration & Protobuf Lifecycle
- **PostgreSQL Migration**: Not required (purely structural refactor).
- **Protobuf Lifecycle**: No `.proto` files touched.
- **Config / Secrets**: No new config values or secrets introduced.

## Phase 1: Go API (`apps/api`)
1. **Shared Postgres Pool (`pkg/postgres/pool.go`) & `main.go`**:
   - Extract `MustNewPool` and `NewTracer` from the 4 duplicated repository constructors (`user`, `organization`, `registry`, `idempotency`).
   - Add `New...RepositoryWithPool(pool, traceProvider)` constructors and retain `New...Repository(cfg, traceProvider)` for backward compatibility in tests.
   - Share one `*pgxpool.Pool` in `apps/api/main.go` and close it in `gracefulShutdown`.
2. **Pagination Helpers (`pkg/pagination/pagination.go`)**:
   - Extract `Normalize(req *shared.Pagination)` and `ComputePageMeta(totalCount, page, pageSize int)` to replace 6 duplicated blocks across `internal/{user,organization,registry}`.
3. **Registry Service Decomposition (`internal/registry/`)**:
   - Split `service.go` (1,297 LOC) into `service.go` (CRUD), `service_git.go` (Git inspection & SSH access), `service_sdk.go` (SDK orchestration & docs), and `service_sdk_deps.go` (checkout, SDK git commit, Go/JS dependency install).
   - Extend `sdkgenerator.CommandRunner` (`RunInDir`, `CombinedOutputInDir`, `OutputInDir`) and inject it into `registry.service` so no `exec.Command` calls exist outside `pkg/sdkgenerator` or transport handlers.
4. **Registry Transport & SSH Server Split (`internal/registry/` & `main.go`)**:
   - Split `handler.go` (1,079 LOC) into `handler.go` (ConnectRPC), `handler_git_ssh.go`, `handler_git_http.go`, `handler_sdk.go`, and `handler_docs.go`.
   - Extract `startSshServer` / `loadOrGenerateHostKey` from `main.go` into `internal/registry/ssh_server.go` with unit tests (`ssh_server_test.go`).
   - Reuse JWT claim verification helper from `pkg/authentication` inside `DocumentationHttpHandler`.
5. **Postgres Repository Split & Queue Helpers (`pkg/postgres/`)**:
   - Move `go-git` filesystem methods out of `pkg/postgres/registry/repository.go` into `pkg/postgres/registry/git_repository.go`.
   - Extract shared job queue helper (`pkg/postgres/queue_helper.go`) for job status transitions and retry/fail handling.

## Phase 2: Dashboard & Shared UI (`apps/dashboard`, `packages/ui`)
1. **Connect Transport Consolidation**:
   - Extract `apps/dashboard/lib/transport.ts` with the single `createConnectTransport` instance (`[csrfInterceptor, authInterceptor, idempotencyInterceptor]`).
   - Import `transport` in both `app/providers.tsx` and `lib/use-client.impl.ts`; make `lib/use-client.ts` re-export `useClient` from `./use-client.impl` (eliminating drift).
2. **ConnectError Toast & Validation Schemas**:
   - Add `apps/dashboard/lib/connect-error.ts` (`handleConnectErrorToast`) and `apps/dashboard/lib/validation-schemas.ts` for shared Zod v4 rules.
3. **Component & Style Decomposition**:
   - Split `apps/dashboard/components/ssh-api-key-panel.tsx` (597 LOC) into `ssh-keys-section.tsx` and `api-keys-section.tsx` (client components preserving existing UI/tests).
   - Extract pure sorting/tree utilities from `packages/ui/src/components/file-tree.tsx` (510 LOC) into `packages/ui/src/components/file-tree-utils.ts`.
   - Deduplicate `packages/ui/src/styles/dashboard.css` by importing `./globals.css`.
