# Helm and Docker Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Author Helm chart and Docker files for the hasir-monorepo, removing plaintext secrets and setting up ExternalSecret and local development docker-compose.

**Architecture:** Use External Secrets Operator + AWS Secrets Manager for Helm secrets. Move Dockerfiles to `docker/api/` and `docker/dashboard/` and configure `docker-bake.hcl` and `docker-compose.yml` to point to them.

**Tech Stack:** Helm, Docker, Docker Compose, External Secrets Operator, AWS Secrets Manager.

---

### Task 1: Remove plaintext secrets from values.yaml

**Files:**
- Modify: `deploy/helm/hasir/values.yaml`

**Step 1: Remove plaintext secrets**
Set `secrets.jwtSecret`, `secrets.smtp.username`, `secrets.smtp.password`, and `postgresql.auth.password` to `""`.

**Step 2: Configure PostgreSQL to use existing secret**
Set `postgresql.auth.existingSecret` to `"hasir-secrets"` and `postgresql.auth.secretKeys.userPasswordKey` to `"HASIR_POSTGRESQL_PASSWORD"`.

**Step 3: Verify**
Run `helm lint deploy/helm/hasir` to ensure the chart is valid.

---

### Task 2: Add values-prod.yaml.example

**Files:**
- Modify: `deploy/helm/hasir/values-prod.yaml.example`

**Step 1: Add comments and configuration**
Add comments explaining that secrets are managed via External Secrets Operator and AWS Secrets Manager.

**Step 2: Verify**
Ensure the file exists and is valid YAML.

---

### Task 3: Create Dockerfiles in docker/api/ and docker/dashboard/

**Files:**
- Create: `docker/api/Dockerfile`
- Create: `docker/dashboard/Dockerfile`
- Delete: `apps/api/Dockerfile`
- Delete: `apps/dashboard/Dockerfile`

**Step 1: Create docker/api/Dockerfile**
Copy the contents of `apps/api/Dockerfile` to `docker/api/Dockerfile` and update the path comment.

**Step 2: Create docker/dashboard/Dockerfile**
Copy the contents of `apps/dashboard/Dockerfile` to `docker/dashboard/Dockerfile` and update the path comment.

**Step 3: Delete old Dockerfiles**
Remove `apps/api/Dockerfile` and `apps/dashboard/Dockerfile`.

---

### Task 4: Update docker-bake.hcl

**Files:**
- Modify: `docker/shared/docker-bake.hcl`

**Step 1: Update Dockerfile paths**
Change `dockerfile` paths for `api` and `dashboard` targets to point to `docker/api/Dockerfile` and `docker/dashboard/Dockerfile`.

---

### Task 5: Update docker-compose.yml for development

**Files:**
- Modify: `docker/docker-compose.yml`

**Step 1: Update paths and build blocks**
- Add `ports: - "5432:5432"` to `postgres`.
- Add `build` block to `hasir-api` and change `expose` to `ports` for `8080`.
- Add `build` block to `hasir-dashboard` and change `expose` to `ports` for `3000`.
- Update `nginx` build context to `./shared/nginx` and volumes to `./shared/certbot/...`.
- Update `certbot` volumes to `./shared/certbot/...`.

---

### Task 6: Copy .env.example

**Files:**
- Create: `docker/.env.example`

**Step 1: Copy .env.example**
Copy the contents of `docker-images/.env.example` to `docker/.env.example`.
