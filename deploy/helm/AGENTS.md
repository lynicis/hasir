# Helm Chart Instructions (`deploy/helm`)

## Commands
| Task | Command |
|------|---------|
| Update chart dependencies | `helm dependency update charts/hasir` |
| Lint chart | `helm lint charts/hasir` |
| Render templates locally | `helm template hasir charts/hasir` |

## Code Map
| Path | Purpose |
|------|---------|
| `charts/hasir/Chart.yaml` | Chart metadata and PostgreSQL subchart dependency |
| `charts/hasir/values.yaml` | Default chart values (Traefik ingress, PVCs, SOPS secret refs) |
| `charts/hasir/templates/` | Kubernetes manifests (deployments, services, PVCs, configmap, secrets, ingress) |
| `cr.yaml` | `chart-releaser` configuration |

## Key Conventions
- Edit chart files under `deploy/helm/charts/hasir/` (required by `cr.yaml` and `ct` in `.github/workflows/helm.yml`).
- Reference SOPS-managed secrets in `values.yaml`; never commit plaintext credentials.
- Release via `hasir-helm@*` tags (`helm-release.yml`), separate from app and proto releases.
