# Tasks: Changesets Integration

## 1. Install and Configure Changesets

- [x] 1.1 Run `bun add -d @changesets/cli` — verify: `@changesets/cli` appears in root `package.json` devDependencies
- [x] 1.2 Run `bunx changeset init` — verify: `.changeset/config.json` and `.changeset/README.md` created
- [x] 1.3 Set `.changeset/config.json` to use `access: "restricted"` and `baseBranch: "main"` — verify: valid JSON
- [x] 1.4 Commit: `chore: setup changesets for monorepo versioning`

## 2. CI/CD Pipeline

- [x] 2.1 Create `.github/workflows/changeset.yml` with `changesets/action@v1`, `contents: write`, `pull-requests: write` — verify: valid YAML
- [x] 2.2 Update `.github/workflows/release.yml` on-push-tags to `hasir-api@*.*.*` / `hasir-dashboard@*.*.*` — verify: existing Docker build job triggers on new tag format
- [x] 2.3 Commit: `ci: configure changesets github action and update release triggers`

## 3. Cleanup

- [x] 3.1 Delete `scripts/release.sh` — verify: file absent from git
- [x] 3.2 Update `package.json` `"release"` script to `"changeset publish"` — verify: `bun run release --dry-run` delegates to changeset
- [x] 3.3 Commit: `chore: remove old release script`

## 4. Integration Verification

- [x] 4.1 Run `make lint` and `make typecheck` — verify: both pass
- [x] 4.2 Run `bun changeset` as a smoke test — verify: interactive prompt appears; exit without writing
- [x] 4.3 Verify `make vuln` exits 0 — no new vulnerabilities
