# AGENTS.md

## Project Overview

Go API service using Connect-RPC (gRPC-compatible over HTTP). Protocol buffer definitions are hosted at `buf.build/hasir/hasir` with generated Go code imported as dependencies. PostgreSQL database with `pgx/v5`, JWT authentication, SSH server for git operations, and OpenTelemetry tracing.

## Commands

### Build & Run

```bash
make dev                    # MODE=development go run main.go (reads config.json)
GOOS=linux GOARCH=amd64 go build -o api ./main.go  # production build
make run-postgres           # start local PostgreSQL via Docker
```

### Test

```bash
go test ./...                                           # all tests
go test -v -coverprofile=coverage.txt ./...             # verbose + coverage
go test -v -run TestService_Register ./internal/user/   # single function
go test -v -run TestService_Register/happy_path ./internal/user/  # subtest
go test -v ./pkg/postgres/user/                         # single package
go test -v ./pkg/postgres/...                           # integration (Docker)
go test -v ./migrations_test.go                         # migration tests
```

### Lint & Format

```bash
make lint                   # golangci-lint run ./...
make lint-fix               # golangci-lint run --fix ./...
make sec                    # gosec -exclude-dir=sdk -exclude-dir=repos ./...
gofmt -w .                  # format all Go files
goimports -w .              # fix import grouping
```

### Mock Generation

```bash
make generate-mocks         # regenerates all mock files via mockgen (uber/mock)
```

After changing any interface in `internal/` or `pkg/`, run `make generate-mocks` to update `*_mock.go` files. Never edit mock files manually.

## Project Structure

```
internal/                    # Domain-specific business logic (3 bounded contexts)
  user/                      # handler.go, service.go, repository.go, model.go, queue.go
  organization/              # Same structure
  registry/                  # Same structure
pkg/                         # Shared libraries
  authentication/            # JWT interceptor, claims extraction
  authorization/             # Role-based access (owner/author/reader)
  config/                    # Koanf-based config (JSON in dev, env vars in prod)
  email/                     # SMTP service with HTML templates
  log/                       # Zap logger initialization
  postgres/                  # PostgreSQL repository implementations
  proto/                     # Shared protobuf type mapping helpers
  sdkgenerator/              # SDK code generation engines
migrations/                  # SQL migration files (up/down pairs)
main.go                      # Application entry point, wiring, graceful shutdown
```

## Architecture Conventions

Each domain in `internal/` follows this layered pattern:
- **handler.go** -- Thin Connect-RPC handler; translates protobuf requests/responses to domain calls. Implements the generated Connect service interface.
- **service.go** -- Business logic. Defines a `Service` interface and unexported `service` struct.
- **repository.go** -- Data access interface only (no implementation). Implementations live in `pkg/postgres/<domain>/`.
- **model.go** -- Domain DTOs with `db:"column"` struct tags for pgx scanning.
- **queue.go** -- Async job queue interface (where applicable).
- **`*_mock.go`** -- Generated mocks. Do not edit.

## Code Style

### Imports

Three groups separated by blank lines, enforced by `goimports`:
1. Standard library
2. Third-party packages
3. Internal packages (`hasir-api/...`)

```go
import (
    "context"
    "errors"

    userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
    "connectrpc.com/connect"

    "hasir-api/pkg/authentication"
)
```

Use aliases to avoid conflicts: `userv1`, `internalOrganization`, `postgresUser`, `gossh`.

### Naming

- **Packages:** lowercase single words (`user`, `registry`, `config`)
- **Interfaces:** PascalCase, exported (`Service`, `Repository`, `Queue`, `MemberRoleChecker`)
- **Concrete types:** unexported structs implementing interfaces (`service`, `handler`)
- **Constructors:** `New<Type>` returning pointer to concrete type (`NewService`, `NewHandler`, `NewPgRepository`)
- **DTOs:** `<Entity>DTO` suffix (`UserDTO`, `RepositoryDTO`, `OrganizationDTO`)
- **Sentinel errors:** `Err<Description>` using connect error codes (`ErrInternalServer`, `ErrMissingToken`)
- **Constants:** typed string constants for enums (`MemberRole`, `InviteStatus`, `Visibility`)
- **Struct tags:** `db:"column_name"` for pgx, `koanf:"field"` for config

### Error Handling

- Return `connect.NewError(connect.Code..., errors.New("message"))` from services and handlers.
- Define reusable sentinel errors at package level: `var ErrInternalServer = connect.NewError(connect.CodeInternal, errors.New("..."))`.
- Check error types with `errors.As(err, &connectErr)` and `errors.Is(err, target)`.
- Use `zap.L().Fatal(...)` only for unrecoverable startup failures.
- Use `zap.L().Error(...)` / `zap.L().Warn(...)` with structured fields for operational errors.
- Add `// #nosec` comments for gosec false positives (validated paths, hardcoded safe commands).

### Logging

Always use the global zap logger with structured fields:
```go
zap.L().Info("message", zap.String("key", value), zap.Error(err))
```

### Formatting

- `gofmt` and `goimports` are enforced via pre-commit hooks.
- No additional style linter rules beyond default `golangci-lint` config.
- Files must end with a newline, no trailing whitespace.

## Testing Conventions

### Strategy
- **Repository layer:** Integration tests with real PostgreSQL via `testcontainers-go`.
- **Service layer:** Unit tests with mocked repositories (gomock).
- **Handler layer:** Unit tests with mocked services (gomock).

### Patterns
- **Table-driven tests** with subtests: `t.Run("case name", func(t *testing.T) { ... })`.
- **gomock** for mocking: `gomock.NewController(t)`, `NewMock<Interface>(ctrl)`, `.EXPECT().Method(...).Return(...).Times(n)`.
- **testify** for assertions: prefer `assert.NoError`, `assert.Equal`, `assert.Error`, `require.NoError`.
- Use `t.Context()` for context in tests (Go 1.26+).
- Mark test helpers with `t.Helper()`.
- Tests live in the same package as the code they test (white-box testing).

### Test File Naming
- Unit tests: `<file>_test.go` alongside the source (e.g., `service_test.go`)
- Integration tests: `repository_test.go` in `pkg/postgres/<domain>/`
- Migration tests: `migrations_test.go` at project root

## CI/CD Pipeline

On push/PR to `master`: lint (golangci-lint v2.6.2) -> test (with coverage) -> static analysis (gosec) -> build. All four must pass. Coverage uploaded to Codecov (excludes `main.go`, `*_mock.go`, `*_test.go`).

Pre-commit hooks (configured in `.pre-commit-config.yaml`): trailing-whitespace, end-of-file-fixer, check-merge-conflict, gofmt, goimports, go test, golangci-lint.

## GitNexus — Code Intelligence

This project is indexed by GitNexus as **api** (1849 symbols, 5935 relationships, 97 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

### Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "master"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.

### Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

### Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/api/context` | Codebase overview, check index freshness |
| `gitnexus://repo/api/clusters` | All functional areas |
| `gitnexus://repo/api/processes` | All execution flows |
| `gitnexus://repo/api/process/{name}` | Step-by-step execution trace |

### CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **api** (1845 symbols, 5931 relationships, 97 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "master"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/api/context` | Codebase overview, check index freshness |
| `gitnexus://repo/api/clusters` | All functional areas |
| `gitnexus://repo/api/processes` | All execution flows |
| `gitnexus://repo/api/process/{name}` | Step-by-step execution trace |

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
