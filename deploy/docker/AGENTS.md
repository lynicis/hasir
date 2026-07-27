# DOCKER DEPLOY KNOWLEDGE BASE

## OVERVIEW
Single-server Docker Compose stack: PostgreSQL 16, API, dashboard, nginx reverse proxy, certbot. Operator-facing alternative to the Helm chart.

## STRUCTURE
```
deploy/docker/
├── docker-compose.yaml            # 5 services
├── docker-compose.override.yaml   # Dev: MODE=development, config.json mount
├── docker-bake.hcl                # Multi-stage buildx targets
├── nginx/
│   ├── Dockerfile + nginx.conf
│   ├── templates/default.conf.template
│   └── includes/                  # security-headers, ssl, gzip, proxy-settings
└── scripts/                       # setup.sh, deploy.sh, backup.sh
```

## SERVICES / PORTS
`postgres:16-alpine` 5432 · `hasir-api` 8080 + 2222 (Git SSH) · `hasir-dashboard` 3000 · `nginx` 80/443 (rate limiting, SSL termination via Let's Encrypt/certbot).

## CONVENTIONS
- Build images via `make docker` (root) — wraps `docker buildx bake`; keep targets in `docker-bake.hcl`.
- The nginx image is built locally from `nginx/` — not pulled.
- The compose override auto-applies in dev: mounts `apps/api/config.json`, sets `MODE=development`.

## ANTI-PATTERNS
- NO plaintext secrets in compose files or committed `.env`.
- Do not bypass bake with ad-hoc `docker build` commands.

## NOTES
- **IMAGE-NAME MISMATCH**: bake/helm push `ghcr.io/lynicis/hasir-{api,dashboard}`, but compose pulls `ghcr.io/lynicis/{api,dashboard}` — compose image refs are broken until aligned.
- The certbot container handles Let's Encrypt issuance/renewal; nginx config is envsubst-templated on deploy.
