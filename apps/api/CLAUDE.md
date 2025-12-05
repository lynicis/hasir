# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hasir API is a Git registry backend service built with Go 1.25.4, using Connect-RPC (gRPC-compatible) over HTTP/2, PostgreSQL for data persistence, and Protocol Buffers for API definitions. The API manages users, organizations, organization members, and Git repositories with JWT-based authentication.

## Development Commands

### Running the Server
```bash
# Development mode (uses config.json)
make dev

# Production mode (uses environment variables with HASIR_ prefix)
go run main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/user

# Run a specific test
go test ./internal/user -run TestServiceRegister
```

### Code Quality
```bash
# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Generate mocks (after modifying interfaces)
make generate-mocks
```

### Pre-commit Hooks
The repository uses pre-commit hooks that run:
- gofmt and goimports for formatting
- go test ./... for testing
- golangci-lint for linting

## Architecture

### Domain-Driven Design Structure

The codebase follows a domain-driven structure with three main domains in `internal/`:
- **user**: User registration, authentication, and management
- **organization**: Organization CRUD, member invitations, and email job processing
- **registry**: Git repository management

Each domain follows a consistent layered pattern:
```
domain/
├── handler.go          # Connect-RPC handlers (HTTP layer)
├── handler_test.go
├── service.go          # Business logic layer
├── service_test.go
├── service_mock.go     # Generated mock for testing
├── repository.go       # Data access layer (PostgreSQL)
├── repository_test.go
├── repository_mock.go  # Generated mock for testing
└── model.go           # Domain models and DTOs
```

### Global Handler Pattern

All domain handlers implement `internal.GlobalHandler`:
```go
type GlobalHandler interface {
    RegisterRoutes() (string, http.Handler)
}
```

This allows main.go to register all routes in a uniform way without knowing domain-specific details.

### Shared Packages in `pkg/`

- **auth**: JWT-based authentication interceptor for Connect-RPC
  - Uses context keys `UserIDKey` and `UserEmailKey` to pass authenticated user info
  - Public methods (Register, Login, RenewTokens) bypass authentication
  - Use `auth.MustGetUserID(ctx)` to extract user ID from context in handlers

- **config**: Configuration management with two modes
  - Development: reads from `config.json` (when `MODE=development`)
  - Production: reads from environment variables with `HASIR_` prefix
  - Use `config.NewConfigReader().Read()` to get configuration

- **email**: SMTP email service for organization invitations
  - Configured via `smtp` section in config

- **organization**: Helper utilities for organization domain

- **proto**: Protocol Buffer definitions (generated, do not edit)

### Key Architectural Patterns

1. **Repository Pattern**: All data access goes through repository interfaces, making it easy to mock for testing

2. **Dependency Injection**: Dependencies are injected through constructors (e.g., `NewService(repo Repository)`)

3. **Connect-RPC Interceptors**: Used for cross-cutting concerns
   - `validate.NewInterceptor()`: Request validation using protobuf constraints
   - `auth.NewAuthInterceptor()`: JWT authentication
   - `otelconnect.NewInterceptor()`: OpenTelemetry tracing (when enabled)

4. **Background Job Processing**: Organization repository runs a background worker that processes email jobs
   - Started in main.go: `organizationPgRepository.StartEmailJobProcessor(ctx, emailService, 10, 5*time.Second)`
   - Gracefully stopped on shutdown

### Database Migrations

Migrations are in `migrations/` using golang-migrate:
- Format: `{version}_{description}.up.sql` and `{version}_{description}.down.sql`
- Applied automatically on server startup in main.go
- Migration client URL format: `postgres://user:pass@host:port/db?sslmode=disable`

### Testing Strategy

All tests use:
- **testcontainers-go** for PostgreSQL integration tests (repository layer)
- **go.uber.org/mock** for mocking service and repository interfaces (handler and service tests)
- Table-driven tests for comprehensive coverage

Mock generation uses mockgen and is run via `make generate-mocks`.

### Git Repository Storage

Git repositories are cloned and stored in the `repos/` directory:
- Each repository is stored in a subdirectory named by its UUID
- Managed by the registry domain

### Configuration

Configuration is loaded via koanf library with two modes:
- **Development** (when `MODE=development`): Loads from `config.json`
- **Production**: Loads from environment variables prefixed with `HASIR_`
  - Example: `HASIR_POSTGRESQL_HOST=localhost`

See `config.example.json` for all available configuration options.

### OpenTelemetry Tracing

When `otel.enabled` is true:
- OTLP traces are exported to the configured endpoint
- Tracing spans are automatically added to all Connect-RPC calls
- pgx database queries are traced via exaring/otelpgx

### Connect-RPC API

The API uses Connect-RPC protocol (compatible with gRPC) over HTTP/2:
- Protocol Buffers generated from buf.build/hasir/hasir
- Handlers return `*connect.Response[T]` and `connect.NewError()` for errors
- Standard HTTP/2 with unencrypted support for development

## Development Workflow

1. **Modifying a domain**: Update service interface → regenerate mocks → update implementation → update tests
2. **Adding a new RPC method**: Update proto definitions in buf.build → update handler → update service → update repository if needed
3. **Database changes**: Create new migration files with sequential version numbers
4. **Adding dependencies**: Use `go get` and run `go mod tidy`

## Testing Notes

- Repository tests use real PostgreSQL via testcontainers (slower but thorough)
- Service and handler tests use mocks (fast unit tests)
- Run specific tests to avoid spinning up containers unnecessarily
- Tests automatically clean up containers and connections
