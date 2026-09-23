# Proposal: SOPS Encrypted Secrets

## Why

The project initially used External Secrets Operator (ESO) for secret management (ADR-0007). ESO requires running an operator in the cluster and cloud-specific IAM configuration. SOPS with `age` keys provides an equivalent security posture with lower operational overhead: secrets are encrypted at rest in git, decrypted by the app at startup (dev) or by `helm-secrets` at deploy time (prod).

## What Changes

- `apps/api/pkg/config/config.go`: Sensitive fields tagged with `sops:"secret"`; `String()` and `slog.LogValuer` methods added for redaction via reflection
- `apps/api/pkg/config/config.go`: SOPS decryption integrated into the config load path for `.sops.json` files
- `scripts/setup-sops.sh`: New helper to generate `age` keys and produce an export command for `SOPS_AGE_KEY_FILE`
- `apps/api/config.sops.json`: Encrypted version of `config.json` committed to repo
- `deploy/helm/hasir/`: Updated to mount `age` key as a Kubernetes Secret and pass `SOPS_AGE_KEY_FILE` to the container

## Capabilities

### Modified Capabilities
- `secrets-management`: ESO replaced by SOPS + `age`; runtime config redaction added

## Impact

- `apps/api/pkg/config/config.go`: Struct tags + redaction methods + SOPS decrypt on load
- `apps/api/go.mod` / `go.sum`: `go.mozilla.org/sops/v3` added
- `scripts/setup-sops.sh`: New helper script
- `apps/api/config.sops.json`: Encrypted config committed
- `deploy/helm/hasir/values.yaml` and templates: `age` key secret mount added

## Implementation Branch

`sops-secrets` (archived — completed)
