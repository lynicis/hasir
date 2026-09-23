# Spec: Secrets Management

> Derived from ADR-0007 (2026-07-14, superseded), SOPS design (2026-07-26) — Status: Accepted (SOPS replaces ESO)

> **Note**: ADR-0007 originally selected External Secrets Operator + AWS Secrets Manager. The 2026-07-26 SOPS design supersedes that decision: the project now uses Mozilla SOPS with `age` keys for local development and `helm-secrets` for cluster deployment.

## Purpose

Define secrets management policies, ensuring zero plaintext secrets in Git, using Mozilla SOPS with age keys for local development, runtime secret redaction in Go, and helm-secrets for deployment.

## Requirements

### Requirement: No Plaintext Secrets in Git

No secret value (JWT signing key, database password, SMTP credentials, TLS private keys) SHALL be committed to the repository in plaintext form.

#### Scenario: Config file encryption

- **WHEN** `apps/api/config.json` (or `config.sops.json`) is committed
- **THEN** all sensitive fields are encrypted with SOPS using the project's `age` public key
- **AND** `git grep` over the commit history finds no unencrypted credential values

#### Scenario: CI secrets injection

- **WHEN** CI runs in GitHub Actions
- **THEN** secrets are injected via GitHub Actions secrets as `HASIR_*` environment variables
- **AND** no plaintext secret appears in workflow logs or artifact outputs

### Requirement: SOPS-Based Config Encryption

The Go API configuration SHALL support loading an encrypted `config.sops.json` file, decrypting it at startup using the `age` key available via `SOPS_AGE_KEY_FILE` or `SOPS_AGE_KEY`.

#### Scenario: Development startup

- **WHEN** `MODE=development` and `SOPS_AGE_KEY_FILE` is set
- **THEN** the API loads `apps/api/config.sops.json`, decrypts it with SOPS, and starts successfully
- **AND** no plaintext secret is written to disk or printed to stdout

#### Scenario: Sensitive field redaction

- **WHEN** the config struct is printed via `%v`, `%s`, or passed to `slog`
- **THEN** fields tagged `sops:"secret"` display as `[REDACTED]` instead of their values
- **AND** `cfg.String()` contains no credential data

### Requirement: Helm Deployment via helm-secrets

Cluster deployments SHALL use `helm-secrets` to decrypt SOPS-encrypted values files on the fly. Kubernetes Secrets SHALL never contain plaintext values baked into the chart.

#### Scenario: Helm deployment with secrets

- **WHEN** `helm secrets upgrade hasir deploy/helm/charts/hasir -f values.sops.yaml` is run
- **THEN** `helm-secrets` decrypts `values.sops.yaml` using the cluster's KMS key
- **AND** the resulting Kubernetes Secret is created without touching an unencrypted file on disk

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Encryption tool | Mozilla SOPS + `age` | Simpler than GPG; no key server; dev-friendly |
| Cluster key | AWS KMS / GCP KMS (prod), `age` (local) | Cloud-native rotation without operator overhead |
| Runtime redaction | Reflection over `sops:"secret"` struct tags | Prevents accidental logging of credentials |
| Helm integration | `helm-secrets` plugin | Decrypts on-the-fly; no intermediate plaintext file |

## Non-Goals

- External Secrets Operator (superseded by SOPS/helm-secrets)
- Manual secret rotation scripts
- Storing secrets in `config.json` for production environments
