# Tasks: Replace Nginx with Traefik in Helm Ingress

## 1. Update Ingress Configuration

- [x] 1.1 In `deploy/helm/charts/hasir/values.yaml`, set `ingress.className: "traefik"` — verify: field updated
- [x] 1.2 Replace `nginx.ingress.kubernetes.io/*` annotations with `traefik.ingress.kubernetes.io/router.tls: "true"` and retain `cert-manager.io/cluster-issuer: letsencrypt-prod` — verify: no nginx annotations remain

## 2. Verification

- [x] 2.1 Run `helm lint deploy/helm/charts/hasir` — verify: 1 chart linted, 0 failures
- [x] 2.2 Commit: `chore(helm): replace nginx with traefik in ingress values`
- [x] 2.3 Run `bun changeset` — verify: `.changeset/` entry present
