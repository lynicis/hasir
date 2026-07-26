# SOPS Encrypted Configuration and Runtime Secrets Protection Design

## Goal
Implement a secure configuration pipeline using Mozilla SOPS to store encrypted configurations in Git, deploy them securely via Helm, and prevent runtime leakage in the Go API.

## 1. Tooling & Encryption
- **Mozilla SOPS**: Used to encrypt and decrypt configuration files.
- **Key Management**: Use `age` keys for local development and AWS KMS / GCP KMS for staging and production environments.
- **Repository Artifacts**: The file `config.enc.yaml` (or `.json`) will be committed to the repository. Unencrypted configuration files will be added to `.gitignore`.

## 2. Deployment Integration
- **Helm Secrets**: We will use the `helm-secrets` plugin to decrypt SOPS files on the fly during deployment.
- **Kubernetes Secrets**: The decrypted configuration will be injected directly into Kubernetes Secrets during `helm upgrade` or `helm install`. The Go application will mount these secrets as environment variables or an in-memory volume.
- **Security Posture**: Unencrypted configurations will never touch the disk in the CI/CD pipeline or deployment artifacts.

## 3. Runtime Protection (Go API)
To prevent accidental leakage through logging or panics:
- **Struct Tagging**: Sensitive fields in `apps/api/pkg/config/config.go` will be tagged (e.g., `secret:"true"`).
- **Redaction Logic**: Implement a custom `String()` method for configuration structs that masks sensitive values (e.g., returning `***REDACTED***`).
- **Structured Logging (`slog`)**: Ensure the logging pipeline strips out any values explicitly identified as secrets before writing to standard output.
- **Memory Clearing**: Where highly sensitive material (like TLS private keys) is processed, ensure byte slices are cleared post-initialization if possible.

## 4. Architectural Updates
- Update `docs/ARCHITECTURE.md` to reflect the migration from External Secrets Operator (ESO) to SOPS (`helm-secrets`).
- Provide documentation on developer onboarding with `age` and SOPS.

## Next Steps
- Update Helm charts to integrate with `helm-secrets`.
- Refactor the Go configuration parsing in `apps/api/pkg/config` to support runtime redaction.
- Create developer scripts for managing local secrets with `sops`.
