# Changesets Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate the changesets tool to manage versions and changelogs, replacing the custom bash scripts and automating PR creation for releases.

**Architecture:** Use `@changesets/cli` for version bumping and changelog generation. Configure independent versioning. Add a GitHub action to handle the "Version Packages" PR and publish steps automatically.

**Tech Stack:** Bun, Changesets, GitHub Actions.

---

### Task 1: Install Changesets and Initialize

**Files:**
- Modify: `package.json`
- Modify: `bun.lock`
- Create: `.changeset/config.json`
- Create: `.changeset/README.md`

**Step 1: Install @changesets/cli**

Run: `bun add -d @changesets/cli`
Expected: Installs `@changesets/cli` into `devDependencies` and updates `bun.lock`.

**Step 2: Initialize Changesets**

Run: `bunx changeset init`
Expected: Creates the `.changeset` directory with a `config.json` and a `README.md`.

**Step 3: Configure Independent Versioning**

Modify `.changeset/config.json` to have:
```json
{
  "$schema": "https://unpkg.com/@changesets/config@3.0.0/schema.json",
  "changelog": "@changesets/cli/changelog",
  "commit": false,
  "fixed": [],
  "linked": [],
  "access": "restricted",
  "baseBranch": "main",
  "updateInternalDependencies": "patch",
  "ignore": []
}
```
*(No changes strictly required for independent mode if `fixed` is empty, but verify that it matches standard config).*

**Step 4: Commit**

```bash
git add package.json bun.lock .changeset
git commit -m "chore: setup changesets for monorepo versioning"
```

---

### Task 2: Configure CI/CD Pipeline

**Files:**
- Create: `.github/workflows/changeset.yml`
- Modify: `.github/workflows/release.yml`

**Step 1: Create changeset.yml**

Create `.github/workflows/changeset.yml`:
```yaml
name: Changesets
on:
  push:
    branches:
      - main

permissions:
  contents: write
  pull-requests: write

jobs:
  version:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest
      - run: bun install
      - name: Create Release Pull Request or Publish
        id: changesets
        uses: changesets/action@v1
        with:
          publish: bunx changeset publish
          version: bunx changeset version
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Step 2: Update release.yml triggers**

Modify `.github/workflows/release.yml`:
Change the `on.push.tags` to match the changesets format (e.g., `hasir-api@*`, `hasir-dashboard@*`). 
*Note: Since Changesets publishes tags and GitHub releases natively via the action if configured, we might not need the custom changelog generation in `release.yml`, but we keep `release.yml` to trigger the Docker builds based on the new tag format.*

```yaml
on:
  push:
    tags:
      - 'hasir-api@*.*.*'
      - 'hasir-dashboard@*.*.*'
```

Remove the "Create Release" job step from `release.yml` since `changesets/action` handles GitHub releases. If `release.yml` is also doing Docker builds, ensure those remain.

**Step 3: Commit**

```bash
git add .github/workflows/changeset.yml .github/workflows/release.yml
git commit -m "ci: configure changesets github action and update release triggers"
```

---

### Task 3: Cleanup Old Release Logic

**Files:**
- Delete: `scripts/release.sh`
- Modify: `package.json`
- Modify: `docs/RELEASE.md`

**Step 1: Delete scripts/release.sh**

Run: `rm scripts/release.sh`

**Step 2: Update package.json scripts**

Modify `package.json` to change the `"release"` script:
```json
"release": "changeset publish"
```

**Step 3: Update documentation**

Modify `docs/RELEASE.md` to reflect that we now use `bunx changeset` to create changeset files, and releases are fully automated via the "Version Packages" PR.

**Step 4: Commit**

```bash
git add scripts/release.sh package.json docs/RELEASE.md
git commit -m "chore: remove old release script and update docs for changesets"
```
