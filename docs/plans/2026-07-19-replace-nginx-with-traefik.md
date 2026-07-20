# Replace Nginx with Traefik in Helm Chart Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Switch the Helm chart ingress configuration to use Traefik instead of Nginx.

**Architecture:** Update the default ingress configurations in `values.yaml` to drop the Nginx-specific class name and annotations, replacing them with standard Traefik compatibility equivalents.

**Tech Stack:** Helm, Kubernetes Ingress, Traefik

---

### Task 1: Update Ingress Configuration in `values.yaml`

**Files:**
- Modify: `deploy/helm/charts/hasir/values.yaml`

**Step 1: Write the minimal implementation**

Update the `ingress` block in `deploy/helm/charts/hasir/values.yaml` to remove Nginx annotations and update the class name.

```yaml
# Target replacement block in values.yaml:
ingress:
  enabled: true
  className: "traefik"
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    traefik.ingress.kubernetes.io/router.tls: "true"
```

**Step 2: Verify the syntax**

Run: `helm lint deploy/helm/charts/hasir`
Expected: PASS with "1 chart(s) linted, 0 chart(s) failed"

**Step 3: Commit**

```bash
git add deploy/helm/charts/hasir/values.yaml
git commit -m "chore(helm): replace nginx with traefik in ingress values"
```
