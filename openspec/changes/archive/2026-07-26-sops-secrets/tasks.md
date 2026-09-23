# Tasks: SOPS Encrypted Secrets

## 1. Go Config — Redaction Struct Tags

- [x] 1.1 Add `sops:"secret"` tags to sensitive fields in `apps/api/pkg/config/config.go` (DB URL, JWT Secret, SMTP password, SSH host key) — verify: `go vet ./pkg/config/...` passes
- [x] 1.2 Implement `String() string` using reflection to mask `sops:"secret"` tagged fields with `[REDACTED]` — verify: `TestConfigRedaction` passes
- [x] 1.3 Implement `slog.LogValuer` interface on `Config` using the same reflection logic — verify: `slog` output contains no credential values
- [x] 1.4 Commit: `feat(api/config): add struct tags and reflection for secret redaction`

## 2. SOPS Decryption on Config Load

- [x] 2.1 Add `go.mozilla.org/sops/v3` to `apps/api/go.mod` via `go get` — verify: `go mod tidy` succeeds
- [x] 2.2 Update config load path: if file ends in `.sops.json`, call `decrypt.Data(raw, "json")` before passing to koanf — verify: `TestSOPSConfigLoad` passes with a test encrypted file
- [x] 2.3 Run `go test -coverprofile=coverage.out ./pkg/config/...` — verify: coverage ≥85%
- [x] 2.4 Commit: `feat(api/config): support SOPS encrypted configuration files`

## 3. Age Key Setup and Config Encryption

- [x] 3.1 Create `scripts/setup-sops.sh` that generates an `age` key and prints `export SOPS_AGE_KEY_FILE=...` — verify: script exits 0; key file created
- [x] 3.2 Encrypt `apps/api/config.json` with SOPS: `sops --encrypt --age <pubkey> config.json > config.sops.json` — verify: `sops -d apps/api/config.sops.json` decrypts successfully
- [x] 3.3 Commit: `chore: setup SOPS scripts and encrypt local development config`

## 4. Helm Chart — Age Key Mount

- [x] 4.1 Update `deploy/helm/hasir/values.yaml` to declare `sops.ageKeySecret: ""` (populated by `values-prod.yaml`) — verify: `helm lint` passes
- [x] 4.2 Add Secret template and volume mount for age key; set `SOPS_AGE_KEY_FILE` env var on the API container — verify: `helm template deploy/helm/hasir` renders correctly
- [x] 4.3 Commit: `feat(helm): add support for mounting age keys for SOPS decryption`

## 5. Integration Verification

- [x] 5.1 Run `go test -cover ./...` in `apps/api` — verify: all packages ≥85% coverage
- [x] 5.2 Run `make vuln` and grep changed files for hardcoded secrets — verify: exits 0, no hits
- [x] 5.3 Run `bun changeset` — verify: `.changeset/` entry for `hasir-api` patch bump
