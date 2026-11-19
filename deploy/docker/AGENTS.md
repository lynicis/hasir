# Docker Deploy Instructions (`deploy/docker`)

## Commands
| Task | Command |
|------|---------|
| Bake all images | `make docker` (from repo root) |
| Validate compose config | `docker compose -f docker-compose.yml config` |

## Code Map
| Path | Purpose |
|------|---------|
| `docker-compose.yml` | Production stack (`postgres`, `hasir-api`, `hasir-dashboard`, `nginx`, `certbot`) |
| `docker-compose.override.yml` | Local dev overrides (`MODE=development`, mounts `apps/api/config.json`) |
| `shared/docker-bake.hcl` | Multi-stage `docker buildx bake` targets |
| `nginx/` | Reverse proxy image and envsubst templates |
| `scripts/` | Operator scripts (`setup.sh`, `deploy.sh`, `backup.sh`) |

## Key Conventions
- Build images through `shared/docker-bake.hcl` (`make docker`); do not use ad-hoc `docker build` commands.
- Keep secrets out of compose files and `.env.example`; copy `.env.example` to `.env` (gitignored) for local runs.
