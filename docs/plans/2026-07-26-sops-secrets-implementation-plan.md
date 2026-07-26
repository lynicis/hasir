# SOPS Encrypted Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement encrypted secrets management for the Hasir application using Mozilla SOPS and Go reflection-based redaction, keeping secrets safe at rest and at runtime.

**Architecture:** We will integrate SOPS for encrypting `config.json` via age, configure Helm to use helm-secrets for deployments, and enhance the Go application's config layer (`apps/api/pkg/config/config.go`) to decrypt the config on load and redact sensitive fields in memory and logs using custom struct tags.

**Tech Stack:** Mozilla SOPS, age, Go 1.26, slog, Helm, helm-secrets.

---

### Task 1: Update Go Config Struct with Redact Tags

**Files:**
- Modify: `apps/api/pkg/config/config.go`

**Step 1: Write the failing test**

```go
// apps/api/pkg/config/config_test.go
package config_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"apps/api/pkg/config"
)

func TestConfigRedaction(t *testing.T) {
    cfg := &config.Config{
        Database: config.DatabaseConfig{
            URL: "postgres://user:password@localhost/db",
        },
        JWT: config.JWTConfig{
            Secret: "super-secret-key",
        },
    }

    // This should print redacted values
    str := cfg.String()
    assert.NotContains(t, str, "password")
    assert.NotContains(t, str, "super-secret-key")
    assert.Contains(t, str, "[REDACTED]")
}
```

**Step 2: Run test to verify it fails**

Run: `cd apps/api && go test ./pkg/config/... -v`
Expected: FAIL (because `String()` doesn't exist or doesn't redact)

**Step 3: Write minimal implementation**

Modify `apps/api/pkg/config/config.go` to add `sops:"secret"` tags to sensitive fields (like DB URL, JWT Secret) and implement the `LogValue` method for `slog` and `String` method for standard printing, iterating via reflection to mask fields with that tag.

```go
// ... add tags in struct ...
type DatabaseConfig struct {
    URL string `koanf:"url" sops:"secret"`
    // ...
}
type JWTConfig struct {
    Secret string `koanf:"secret" sops:"secret"`
    // ...
}

// ... implement LogValue() slog.Value and String() string using reflection to mask "secret" tags
```

**Step 4: Run test to verify it passes**

Run: `cd apps/api && go test ./pkg/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/api/pkg/config/config.go apps/api/pkg/config/config_test.go
git commit -m "feat(api/config): add struct tags and reflection for secret redaction"
```

---

### Task 2: Implement SOPS Decryption on Config Load

**Files:**
- Modify: `apps/api/pkg/config/config.go`

**Step 1: Write the failing test**

```go
// apps/api/pkg/config/sops_test.go
// Create a dummy sops encrypted file and test that Load() successfully decrypts it.
```

**Step 2: Run test to verify it fails**

Run: `cd apps/api && go test ./pkg/config/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Update `apps/api/pkg/config/config.go` to intercept the file reading process. If the file has a `.sops.json` or `.sops.yaml` extension, invoke the `go.mozilla.org/sops/v3/decrypt` package to decrypt the byte slice before passing it to `koanf.Load()`. Add `go.mozilla.org/sops/v3` to `go.mod`.

**Step 4: Run test to verify it passes**

Run: `cd apps/api && go mod tidy && go test ./pkg/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/api/pkg/config/config.go apps/api/go.mod apps/api/go.sum
git commit -m "feat(api/config): support SOPS encrypted configuration files"
```

---

### Task 3: Setup Age Key and Encrypt Example Config

**Files:**
- Create: `scripts/setup-sops.sh`
- Create: `apps/api/config.sops.json` (encrypted version of config.json)

**Step 1: Write minimal implementation**

Create a helper script `scripts/setup-sops.sh` to generate an `age` key for local development and output the `SOPS_AGE_KEY_FILE` export command. Then use `sops --encrypt --age <public_key> config.json > config.sops.json`.

**Step 2: Verify**

Run: `sops -d apps/api/config.sops.json`
Expected: Outputs the decrypted configuration.

**Step 3: Commit**

```bash
git add scripts/setup-sops.sh apps/api/config.sops.json
git commit -m "chore: setup SOPS scripts and encrypt local development config"
```

---

### Task 4: Update Helm Chart for SOPS Integration

**Files:**
- Modify: `deploy/helm/hasir/values.yaml`
- Modify: `deploy/helm/hasir/templates/deployment.yaml` (or secret template)

**Step 1: Write minimal implementation**

Update the Helm chart to support mounting the `SOPS_AGE_KEY_FILE` as a Kubernetes Secret and passing the path to the application container via an environment variable, so the Go application can decrypt the config at startup in the cluster.

**Step 2: Verify**

Run: `helm lint deploy/helm/hasir`
Expected: No errors.

**Step 3: Commit**

```bash
git add deploy/helm/hasir/
git commit -m "feat(helm): add support for mounting age keys for SOPS decryption"
```
