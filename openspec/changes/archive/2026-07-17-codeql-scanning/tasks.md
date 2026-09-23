# Tasks: CodeQL Code Scanning

## 1. Create CodeQL Workflow

- [x] 1.1 Create `.github/workflows/codeql.yml` with matrix strategy (`go`, `javascript-typescript`), `security-events: write` permissions, and Buf proto pre-generation step — verify: valid YAML (`yamllint`), GitHub Actions schema validates

## 2. Integration Verification

- [x] 2.1 Push to a feature branch and confirm CodeQL workflow appears in GitHub Actions — verify: both language analyses complete without errors
- [x] 2.2 Run `make vuln` and confirm no new vulnerabilities introduced — verify: exits 0
- [x] 2.3 Run `bun changeset` — verify: `.changeset/` file present
