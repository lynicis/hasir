# AGENTS.md

## Project Overview

Go API service using Connect-RPC (gRPC-compatible over HTTP). Protocol buffer definitions are hosted at `buf.build/hasir/hasir` with generated Go code imported as dependencies. PostgreSQL database with `pgx/v5`, JWT authentication, SSH server for git operations, and OpenTelemetry tracing.

## Build & Run Commands

```bash
# Run locally (development mode, reads config.json)
make dev                    # MODE=development go run main.go

# Build for production
GOOS=linux GOARCH=amd64 go build -o api ./main.go

# Start local PostgreSQL (Docker)
make run-postgres
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run all tests with verbose output and coverage
go test -v -coverprofile=coverage.txt ./...

# Run a single test function
go test -v -run TestService_Register ./internal/user/

# Run a single subtest
go test -v -run TestService_Register/happy_path ./internal/user/

# Run all tests in a specific package
go test -v ./pkg/postgres/user/

# Run only integration tests (requires Docker for testcontainers)
go test -v ./pkg/postgres/...
go test -v ./migrations_test.go
```

## Lint & Format Commands

```bash
make lint                   # golangci-lint run ./...
make lint-fix               # golangci-lint run --fix ./...
make sec                    # gosec -exclude-dir=sdk -exclude-dir=repos ./...
gofmt -w .                  # format all Go files
goimports -w .              # fix import grouping
```

## Mock Generation

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
- Use `t.Context()` for context in tests (Go 1.24+).
- Mark test helpers with `t.Helper()`.
- Tests live in the same package as the code they test (white-box testing).

### Test File Naming
- Unit tests: `<file>_test.go` alongside the source (e.g., `service_test.go`)
- Integration tests: `repository_test.go` in `pkg/postgres/<domain>/`
- Migration tests: `migrations_test.go` at project root

## CI Pipeline

On push/PR to `master`: lint (golangci-lint v2.6.2) -> test (with coverage) -> static analysis (gosec) -> build. All four must pass. Coverage uploaded to Codecov (excludes `main.go`, `*_mock.go`, `*_test.go`).

## Pre-commit Hooks

Configured in `.pre-commit-config.yaml`: trailing-whitespace, end-of-file-fixer, check-merge-conflict, gofmt, goimports, go test, golangci-lint. All hooks run before each commit.
