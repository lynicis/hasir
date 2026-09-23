# Proposal: Replace Nginx with Traefik in Helm Ingress

## Why

The Helm chart's ingress was configured with Nginx-specific annotations and `className: nginx`. The target cluster uses Traefik as its ingress controller. Keeping the Nginx configuration would cause ingress routing to fail silently on any Traefik-based cluster.

## What Changes

- `deploy/helm/charts/hasir/values.yaml` ingress block updated: `className` changed to `traefik`, Nginx annotations replaced with Traefik-compatible annotations (`traefik.ingress.kubernetes.io/router.tls`, `cert-manager.io/cluster-issuer`)

## Capabilities

### Modified Capabilities
- `ci-cd`: Helm chart ingress now targets Traefik

## Impact

- `deploy/helm/charts/hasir/values.yaml`: `ingress.className` and `ingress.annotations` updated

## Implementation Branch

`replace-nginx-with-traefik` (archived — completed)
