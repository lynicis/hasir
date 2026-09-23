# Design: Replace Nginx with Traefik in Helm Ingress

## Context

The original ingress in `values.yaml` was written for an Nginx ingress controller. The production cluster was subsequently migrated to Traefik. This change is a minimal values-only update — no template changes required.

## Goals / Non-Goals

**Goals:**
- Switch `className: nginx` to `className: traefik`
- Replace `nginx.ingress.kubernetes.io/*` annotations with Traefik equivalents
- Keep TLS termination via cert-manager

**Non-Goals:**
- Changing Helm templates
- Adding Traefik-specific middleware CRDs

## Decisions

### Decision 1: Minimal values-only change

The ingress template already uses the standard `ingressClassName` field and generic annotation map. Only `values.yaml` needs updating — no template edits required.

### Decision 2: Retain cert-manager annotation

`cert-manager.io/cluster-issuer: letsencrypt-prod` is controller-agnostic and must remain.

## PostgreSQL Migration

Not required.
