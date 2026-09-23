# Design: SOPS Encrypted Secrets

## Context

ADR-0007 selected External Secrets Operator. The SOPS design (2026-07-26) supersedes it by noting that `helm-secrets` + SOPS provides the same no-plaintext-in-git guarantee with no cluster operator to maintain. The Go config layer needs to decrypt secrets at startup and must not accidentally log them.

## Goals / Non-Goals

**Goals:**
- Encrypt `config.json` with SOPS using `age` public key; commit `config.sops.json`
- Decrypt transparently at API startup via `go.mozilla.org/sops/v3/decrypt`
- Add runtime redaction of sensitive fields via `sops:"secret"` struct tags and reflection
- Configure `helm-secrets` to decrypt values during `helm upgrade`

**Non-Goals:**
- Migrating away from `koanf` for config loading
- Encrypting non-sensitive fields
- Using GPG instead of `age`

## Decisions

### Decision 1: `age` for local; KMS for prod

`age` keys for local development are generated via `scripts/setup-sops.sh` and loaded from `SOPS_AGE_KEY_FILE`. Production uses AWS KMS or GCP KMS configured in `.sops.yaml`.

### Decision 2: Reflection-based redaction

Config struct fields tagged `sops:"secret"` are masked in `String()` and `LogValue()` by iterating struct fields via reflection. This avoids per-field manual redaction and works automatically when new sensitive fields are added with the tag.

### Decision 3: SOPS decrypt in koanf load path

Detect the `.sops.json` extension in the file-load step. If detected, pass the raw bytes through `decrypt.Data(raw, "json")` before handing them to koanf. This keeps the config API unchanged — callers always receive a decrypted `Config` struct.

### Decision 4: Helm — mount age key as Kubernetes Secret

The `age` private key is stored in a Kubernetes Secret and mounted as a file into the API container via `SOPS_AGE_KEY_FILE`. This avoids storing the key in `values.yaml`.

## PostgreSQL Migration

Not required.
