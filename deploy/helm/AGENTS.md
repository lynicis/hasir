# AGENTS.md — helm-charts

Single Helm chart for deploying the **Hasir stack** (API + Dashboard + PostgreSQL) on Kubernetes.

## Quick commands

```bash
# Prerequisites
helm repo add bitnami https://charts.bitnami.com/bitnami

# Before any helm command
helm dependency update

# Install
helm install hasir . --namespace hasir --create-namespace

# Override values
helm install hasir . -f my-values.yaml
```

No Makefile, no test/lint runners, no typecheck — all `helm` CLI.

## Chart structure

- `Chart.yaml` — v2 Helm chart, depends on Bitnami PostgreSQL 15.5.3
- `values.yaml` — all config surface (domain, images, persistence, SMTP, secrets, TLS)
- `templates/` — standard resources: Deployments (api, dashboard), Services (api-http, api-ssh, dashboard), PVCs (ssh-keys, repos, sdk), ConfigMap, Secrets, Ingress
- ConfigMap drives all env vars (API connects to DB/dashboard via config)

## Architecture notes

- **API** container exposes HTTP (8080) and SSH (2222) ports. SSH gets a separate `LoadBalancer` service for git push.
- **Dashboard** (port 3000) is a static/Next.js app reading `NEXT_PUBLIC_API_URL` from ConfigMap.
- **Ingress** routes `/api` → API service, `/` → Dashboard service. TLS via cert-manager + letsencrypt.
- **Persistence**: 3 PVCs (ssh-keys 1Gi, repos 5Gi, sdk 5Gi) gated by `api.persistence.enabled`.

## CI

GitHub Actions on push to `main`: `helm/chart-releaser-action` publishes to GitHub Pages. No lint/test step.

## Secrets

`secrets.yaml` template reads from `values.yaml` and base64-encodes. In production, override via `--set` or a separate values file — never commit real secrets to `values.yaml`.

## GitNexus

This repo is indexed as **helm-charts** but consists only of YAML templates (no code). GitNexus impact/context tools are not useful here.
