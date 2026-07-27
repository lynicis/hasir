# HELM CHART KNOWLEDGE BASE

## OVERVIEW
Kubernetes deployment for Hasir. Chart at `deploy/helm/charts/hasir/` (NOT `deploy/helm/hasir/` — older docs are stale). Own release channel via chart-releaser.

## STRUCTURE
```
deploy/helm/
├── charts/hasir/
│   ├── Chart.yaml            # chart v0.1.1
│   ├── values.yaml           # All configuration
│   ├── templates/            # 9: api-deployment, api-service, api-pvc,
│   │                         #    dashboard-deployment, dashboard-service,
│   │                         #    configmap, secrets, ingress, _helpers
│   └── charts/               # Vendored bitnami postgresql-15.5.3.tgz subchart
└── cr.yaml                   # chart-releaser config
```

## VALUES MAP
- `global.domain` — ingress host base.
- `api`: service 8080, SSH exposed as LoadBalancer on 2222 (separate from ingress); PVCs `sshKeys` 1Gi, `repos` 5Gi, `sdk` 5Gi.
- `dashboard`: service 3000.
- `ingress.className`: **traefik** (NOT nginx).
- `config.smtp`; `secrets.{jwtSecret,smtp,sopsAgeKey}` — SOPS-managed.

## CONVENTIONS
- Images: `ghcr.io/lynicis/hasir-api` + `hasir-dashboard` (match docker-bake; NOT the compose names).
- `make setup` runs `helm dependency update` (vendored subchart).
- CI (`helm.yml`): `ct lint` + kind install test on `deploy/helm/**` changes.

## RELEASE
- Tag `hasir-helm@*` → `helm-release.yml` → chart-releaser publishes from `charts_dir: deploy/helm/charts`.

## ANTI-PATTERNS
- NEVER commit plaintext secrets — values reference SOPS-decrypted secrets only.
- Do not restructure the chart dir layout — chart-releaser and `ct` configs depend on it.

## NOTES
- PostgreSQL runs as the vendored subchart, not an external DB.
- Rollback: `helm rollback <release> <revision>` (see docs/RELEASE.md).
