# API Service Instructions (`apps/api`)

## Commands
| Task | Command |
|------|---------|
| Start local Postgres | `make run-postgres` |
| Run API in dev mode | `make dev` |
| Test single package | `go test -count=1 ./internal/registry/...` |
| Test single test case | `go test -count=1 -run TestName ./internal/user/...` |
| Lint package | `go tool golangci-lint run ./internal/organization/...` |
| Regenerate gomock files | `make generate-mocks` |
| Security scan | `make sec` |

## Code Map
| Area | Location |
|------|----------|
| Entrypoint & interceptor chain | `main.go` |
| Domain handlers & services | `internal/{user,organization,registry}/` |
| Git HTTP/SSH & SDK/Docs handlers | `internal/registry/handler_{git_http,git_ssh,sdk,docs}.go`, `internal/registry/ssh_server.go` |
| Postgres repositories & queues | `pkg/postgres/{user,organization,registry,idempotency}/` |
| Interceptors & auth | `pkg/{csrf,idempotency,authentication,authorization}/` |
| SDK codegen runners | `pkg/sdkgenerator/` |
| Raw SQL migrations | `migrations/` (asserted in `migrations_test.go`) |

## Key Conventions
- Share the single `*pgxpool.Pool` initialized via `postgres.MustNewPool` in `main.go` across all repositories.
- Maintain unary Connect interceptor order: `[csrf, validate, idempotency, auth, otel?]`.
- Use `gomock` (`*_mock.go`) for service/handler unit tests and `testcontainers` (real Postgres) for `pkg/postgres` and `migrations_test.go`.
- Use `MODE=development` to load `config.json` locally; production reads `HASIR_*` environment variables.
