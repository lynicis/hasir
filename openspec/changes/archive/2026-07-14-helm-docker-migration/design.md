# Design: Helm and Docker Restructuring

## Context

The monorepo initially preserved the original per-repo Dockerfile locations. A separate `docker-images` repository had housed all Docker assets. Post-merge, Dockerfiles needed to move to `deploy/docker/` and the Helm chart needed to integrate with External Secrets Operator to eliminate plaintext secrets from `values.yaml`.

## Goals / Non-Goals

**Goals:**
- Remove all plaintext secrets from Helm chart values
- Centralise all Docker build assets under `deploy/docker/`
- Make local compose development work with proper port bindings and volumes

**Non-Goals:**
- Application code changes
- Introducing a new CI pipeline for Docker builds (existing `docker.yml` updated only to use new paths)

## Decisions

### Decision 1: Dockerfile relocation

Move `apps/api/Dockerfile` → `deploy/docker/api/Dockerfile` and `apps/dashboard/Dockerfile` → `deploy/docker/dashboard/Dockerfile`. Update `docker-bake.hcl` accordingly. Rationale: deployment assets belong under `deploy/`, not in application source trees.

### Decision 2: ESO for secrets

Reference `existingSecret: "hasir-secrets"` in `postgresql.auth` and remove `secrets.jwtSecret`, `secrets.smtp.*` plaintext fields from `values.yaml`. The `ExternalSecret` resource syncs from AWS Secrets Manager at deploy time. Rationale: ADR-0007 mandates no plaintext secrets in git.

### Decision 3: docker-compose port binding

Expose `5432:5432` for Postgres, `8080:8080` for API, `3000:3000` for Dashboard to enable direct service access during local development without nginx as gateway.

## PostgreSQL Migration

Not required — no schema changes in this change.
