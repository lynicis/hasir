# Proposal: Helm and Docker Restructuring

## Why

The original repository stored Dockerfiles inside each app directory (`apps/api/Dockerfile`, `apps/dashboard/Dockerfile`) and kept production secrets in plaintext in `values.yaml`. This violated separation-of-concerns and security requirements. Docker build assets needed to live under `deploy/docker/` and secrets needed to be removed from the chart.

## What Changes

- Plaintext secrets removed from `deploy/helm/hasir/values.yaml`; replaced with External Secrets Operator references
- Dockerfiles moved from `apps/api/` and `apps/dashboard/` to `deploy/docker/api/` and `deploy/docker/dashboard/`
- `deploy/docker/shared/docker-bake.hcl` updated to reference new Dockerfile paths
- `deploy/docker/docker-compose.yml` updated with correct build contexts, port bindings, and volume paths
- `deploy/docker/.env.example` created from the legacy `docker-images/.env.example`

## Capabilities

### Modified Capabilities
- `ci-cd`: Docker image build pipeline updated to use new Dockerfile paths
- `secrets-management`: Helm chart now uses `ExternalSecret` resources instead of hardcoded values

## Impact

- `deploy/helm/hasir/values.yaml`: Remove all plaintext secret fields; add ESO `existingSecret` references
- `deploy/helm/hasir/values-prod.yaml.example`: New file documenting expected production overrides
- `deploy/docker/api/Dockerfile`: Relocated from `apps/api/Dockerfile`
- `deploy/docker/dashboard/Dockerfile`: Relocated from `apps/dashboard/Dockerfile`
- `deploy/docker/shared/docker-bake.hcl`: Dockerfile path references updated
- `deploy/docker/docker-compose.yml`: Build contexts, port exposure, and volume paths corrected
- `deploy/docker/.env.example`: New file

## Implementation Branch

`helm-docker-migration` (archived — completed)
