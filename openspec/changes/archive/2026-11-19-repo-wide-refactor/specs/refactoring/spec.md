# Spec: Repo-Wide Structural Refactoring

## ADDED Requirements

### Requirement: Behavior Preservation Across Refactored Boundaries
The refactored Go API services, handlers, repositories, and Dashboard components SHALL preserve all pre-refactor functional outputs, status codes, error mappings, and persistence side effects.

#### Scenario: Unit, integration, and UI parity after structural split
- **WHEN** any ConnectRPC handler, HTTP/SSH Git/SDK/Docs endpoint, or background job queue is invoked after decomposition
- **THEN** its response payload, error codes, side effects, and ordering SHALL match pre-refactor behavior
- **AND** existing unit tests (`go test`), integration tests, and Dashboard tests (`bun test`) SHALL pass without weakening assertions.

### Requirement: Single CommandRunner Boundary & Shared Database Pool
The Go API SHALL route external CLI invocations in `registry.service` through `sdkgenerator.CommandRunner` and share a single PostgreSQL connection pool initialized in `main.go`.

#### Scenario: Registry service delegates external commands to CommandRunner
- **WHEN** `registry.service` invokes external CLI commands (`git`, `tar`, `go`, `npm`, `npx`)
- **THEN** it SHALL invoke `sdkgenerator.CommandRunner` rather than calling `os/exec.Command` directly.

#### Scenario: Single shared pgxpool lifecycle
- **WHEN** `apps/api/main.go` boots
- **THEN** it SHALL initialize a single `*pgxpool.Pool` via `postgres.MustNewPool`, share it across all four Postgres repositories, and close it during `gracefulShutdown`.

### Requirement: Single Connect Transport & Deduplicated UI Helpers
The Dashboard and shared UI package SHALL share a single Connect transport instance and a single CSS token definition.

#### Scenario: Unified Connect transport in Dashboard
- **WHEN** `apps/dashboard` creates Connect clients in `providers.tsx` or `useClient`
- **THEN** both SHALL share the single transport instance in `lib/transport.ts` with `[csrfInterceptor, authInterceptor, idempotencyInterceptor]`.

#### Scenario: Single CSS theme source of truth
- **WHEN** `packages/ui` imports `dashboard.css`
- **THEN** it SHALL delegate directly to `globals.css` rather than duplicating CSS token definitions.
