# API SERVICE KNOWLEDGE BASE

## OVERVIEW
Go 1.26 API service providing ConnectRPC, PostgreSQL, JWT authentication, and a custom Git-over-SSH server.

## STRUCTURE
```
apps/api/
├── internal/       # Application domain packages (user, registry, organization)
├── pkg/            # Reusable packages (sdkgenerator, postgres, authentication, config)
├── migrations/     # PostgreSQL SQL migrations
├── main.go         # Application entry point
└── Makefile        # Go service commands
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Domain Logic | `internal/` | Core business logic |
| Shared Utils | `pkg/` | Reusable cross-domain packages |
| DB Migrations| `migrations/` | Raw SQL only |

## CONVENTIONS
- Run `make dev` or `make build` from the monorepo root (Turborepo managed).
- Go services use a thin `package.json` wrapper for Turborepo.

## ANTI-PATTERNS
- DO NOT use ORMs for migrations; use raw SQL files in `migrations/`.
- DO NOT add JS/Node.js dependencies to this Go service.
