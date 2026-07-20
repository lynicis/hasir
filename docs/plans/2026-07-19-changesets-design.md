# Changesets Integration Design

## Overview
Replace the custom `scripts/release.sh` versioning logic with [Changesets](https://github.com/changesets/changesets) to properly manage package versions and generate automated `CHANGELOG.md` files for the monorepo applications.

## 1. Initial Setup
- Install `@changesets/cli` at the monorepo root.
- Initialize with `.changeset/config.json`.
- Set configuration to operate in independent mode so that `hasir-api` and `hasir-dashboard` are versioned independently.

## 2. Tag Format Migration
- Transition from manual `<app>/v<version>` tags (e.g. `api/v1.0.0`) to Changesets native `<package-name>@<version>` format (e.g., `hasir-api@1.0.0`).
- Update the GitHub Actions workflow triggers in `.github/workflows/release.yml` to trigger on `hasir-api@*` and `hasir-dashboard@*`.

## 3. CI/CD Pipeline Update
- Create `.github/workflows/changeset.yml` using `changesets/action`.
- The action will automatically create/update a "Version Packages" Pull Request when standard changesets are pushed to the `main` branch.
- Upon merging the Version Packages PR, the action will tag the repository and create the GitHub Releases automatically (which then trigger the Docker build workflow).

## 4. Cleanup
- Delete the deprecated `scripts/release.sh`.
- Remove or update the `"release"` script in `package.json`.
- Update `docs/RELEASE.md` and `docs/ARCHITECTURE.md` (or `AGENTS.md`) to reflect the new release process via Changesets.
