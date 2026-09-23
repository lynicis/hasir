# Design: Changesets Integration

## Context

The legacy `scripts/release.sh` calculated version bumps via positional arguments (`api patch`) and pushed tags manually. It produced no changelogs and required human knowledge of which packages changed. Changesets automates this by tracking intent-driven changeset files authored at PR time.

## Goals / Non-Goals

**Goals:**
- Replace manual tag pushing with Changesets-driven "Version Packages" PR flow
- Generate `CHANGELOG.md` per app automatically on release
- Update release workflow triggers to the new tag format

**Non-Goals:**
- Publishing npm packages publicly (`access: "restricted"` in config)
- Linking package versions (each app versions independently)

## Decisions

### Decision 1: Independent mode

`.changeset/config.json` uses default independent mode (no `fixed` groups). Each app (hasir-api, hasir-dashboard) versions independently.

### Decision 2: changesets/action for PR automation

`changesets/action@v1` runs on every `main` push. It creates/updates the Version Packages PR when changesets are present, and publishes (tags + GitHub Releases) when the PR is merged.

### Decision 3: Release workflow trigger migration

`release.yml` on-push-tags changed from `api/v*.*.*` to `hasir-api@*.*.*` and `hasir-dashboard@*.*.*` to match Changesets native tag format. The Docker build jobs remain unchanged.

### Decision 4: Remove scripts/release.sh

The script is deleted. The `make release` Makefile target is updated to delegate to `bunx changeset`.

## PostgreSQL Migration

Not required.
