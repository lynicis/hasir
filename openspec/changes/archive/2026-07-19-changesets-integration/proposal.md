# Proposal: Changesets Integration

## Why

The project used a custom `scripts/release.sh` bash script for versioning, which required manual version bumping and produced no changelogs. Switching to Changesets provides structured, automated versioning with generated `CHANGELOG.md` files and a "Version Packages" PR workflow that integrates natively with GitHub Actions.

## What Changes

- `@changesets/cli` installed as a devDependency; `.changeset/config.json` configured for `independent` mode
- `.github/workflows/changeset.yml` created: opens/updates "Version Packages" PR automatically on every `main` push
- `.github/workflows/release.yml` trigger updated from `api/v*.*.*` tags to `hasir-api@*.*.*` / `hasir-dashboard@*.*.*`
- `scripts/release.sh` deleted; `package.json` `"release"` script updated to `changeset publish`
- `docs/RELEASE.md` updated to reflect new Changesets-based workflow

## Capabilities

### Modified Capabilities
- `release-strategy`: Tag format migrated from `api/v<semver>` to `hasir-api@<semver>`; release process fully automated

## Impact

- `package.json`: `@changesets/cli` devDep added; `release` script updated
- `.changeset/config.json`: New file
- `.github/workflows/changeset.yml`: New file
- `.github/workflows/release.yml`: Tag trigger updated
- `scripts/release.sh`: Deleted
- `docs/RELEASE.md`: Updated (now superseded by `openspec/specs/release-strategy/spec.md`)

## Implementation Branch

`changesets-integration` (archived — completed)
