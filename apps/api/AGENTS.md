# API SERVICE KNOWLEDGE BASE

## OVERVIEW
Go 1.26 API: ConnectRPC services, PostgreSQL (pgx/v5), JWT auth, gliderlabs Git-over-SSH server, SDK-generation job queue. ~51 Go files.

## STRUCTURE
```
apps/api/
├── internal/
│   ├── user/           # Register/Login/tokens, bcrypt+JWT, API+SSH keys
│   ├── organization/   # Owner-only ops, invites, email queue, pg_trgm search
│   ├── registry/       # Repo lifecycle on disk ./repos, go-git, SDK orchestration
│   └── handler.go      # GlobalHandler iface: RegisterRoutes() (string, http.Handler)
├── pkg/
│   ├── sdkgenerator/   # buf SDK generation (see pkg/sdkgenerator/AGENTS.md)
│   ├── postgres/       # pgxpool+otelpgx repos; FOR UPDATE SKIP LOCKED job queues
│   ├── authentication/ # JWT interceptor + public-procedure whitelist
│   ├── authorization/  # MemberRoleChecker, OrgRepositoryAdapter
│   ├── config/         # koanf; JsonConfig+SOPS (dev) vs EnvConfig HASIR_* (prod)
│   ├── email/          # SMTP 465/587/plain, embedded HTML templates
│   ├── log/            # zap JSON, init side-effect
│   └── proto/          # Visibility bidirectional maps
├── migrations/         # 18 up/down SQL pairs (golang-migrate, file://)
└── main.go             # Entry: full wiring + graceful shutdown
```

## ENTRY FLOW (main.go)
`config.Read` → `migrate.Up` (ErrNoChange ok) → `initTracer` → 3 pg repos → `EmailJobQueue.Start` (10 workers, 5s) → `registry.NewService` → `SdkGenerationJobQueue.Start` (registryService is BOTH generator AND trigger processor) → user/org services → interceptors [validate, auth, otel?] → GlobalHandler routes on one mux + raw `/git/` `/sdk/` `/docs/` → `http.Server` h1+h2c → `startSshServer` if `cfg.Ssh.Enabled` (public-key auth via `userRepo.GetUserBySshPublicKey`, RSA-4096 host key) → gracefulShutdown 10s.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Business logic | `internal/{user,registry,organization}` | service.go per domain |
| HTTP/SSH transports | `internal/registry/handler.go` | 1041 LOC, 5 transports |
| DB queries | `pkg/postgres/<domain>/` | Hand-built SQL |
| Job queues | `pkg/postgres/{organization,registry}/queue.go` | Polling workers |
| Migrations | `migrations/` | Raw SQL, runs at startup |
| SSH server | `main.go` startSshServer + registry Git handlers | Port 2222 |

## CONVENTIONS
- golangci-lint defaults (config nearly empty); tools pinned via `go tool` directives (golangci-lint, mockgen, gosec, govulncheck).
- Tests: repository layer = testcontainers (real postgres); service/handler = `go.uber.org/mock` gomock (`make generate-mocks`); table-driven; `go test -count=1 -coverprofile`.
- Mocks generated per domain: user, registry, organization, email, authorization.
- Migrations: raw SQL pairs `NNNNNN_name.{up,down}.sql`; golang-migrate runs `file://migrations` at startup; `migrations_test.go` (testcontainers) is the canonical schema assertion (12 tables + matview + pg_trgm).
- Run `make` targets from the monorepo root (Turborepo); thin `package.json` wrapper only.

## ANTI-PATTERNS
- DO NOT use ORMs for migrations; raw SQL files in `migrations/` only.
- DO NOT add JS/Node.js dependencies to this Go service.
- DO NOT use `exec.Command` outside the `CommandRunner` interface (sdkgenerator).
- DO NOT hardcode BSR module mappings (see `pkg/sdkgenerator/AGENTS.md`).

## NOTES
- Hotspots: `internal/registry/service.go` 1169 LOC; `handler.go` 1041 LOC; `pkg/postgres/registry/queue.go` 542 LOC.
- `ssh_host_key` + `config.sops.json` are committed in this dir — treat as sensitive.
- Git-over-SSH paths (`startSshServer`, `GitSshHandler`) have zero test coverage.
- gosec excludes `sdk/`, `repos/` (generated/user content).
