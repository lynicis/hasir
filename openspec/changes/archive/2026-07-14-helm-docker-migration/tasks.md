# Tasks: Helm and Docker Restructuring

## 1. Helm Chart — Remove plaintext secrets

- [x] 1.1 Set `secrets.jwtSecret`, `secrets.smtp.*`, and `postgresql.auth.password` to `""` in `values.yaml` — verify: `helm lint deploy/helm/hasir` passes
- [x] 1.2 Add `postgresql.auth.existingSecret: "hasir-secrets"` and `postgresql.auth.secretKeys.userPasswordKey: "HASIR_POSTGRESQL_PASSWORD"` — verify: lint passes
- [x] 1.3 Create `deploy/helm/hasir/values-prod.yaml.example` with annotated ESO-based config — verify: valid YAML

## 2. Docker — Relocate Dockerfiles

- [x] 2.1 Create `deploy/docker/api/Dockerfile` as copy of `apps/api/Dockerfile` — verify: `docker build -f deploy/docker/api/Dockerfile .` succeeds
- [x] 2.2 Create `deploy/docker/dashboard/Dockerfile` as copy of `apps/dashboard/Dockerfile` — verify: build succeeds
- [x] 2.3 Delete `apps/api/Dockerfile` and `apps/dashboard/Dockerfile` — verify: neither path exists in git

## 3. docker-bake.hcl — Update paths

- [x] 3.1 Change `dockerfile` for `api` and `dashboard` targets to new paths — verify: `docker buildx bake --print` shows correct paths

## 4. docker-compose.yml — Fix dev bindings

- [x] 4.1 Add `ports: ["5432:5432"]` to postgres service — verify: `psql` connects from host
- [x] 4.2 Add build blocks and expose `8080:8080` / `3000:3000` for api/dashboard — verify: services reachable from host
- [x] 4.3 Update nginx build context to `./shared/nginx` and certbot volumes to `./shared/certbot/` — verify: compose up succeeds

## 5. .env.example

- [x] 5.1 Create `deploy/docker/.env.example` from legacy `docker-images/.env.example` — verify: file present and documents all required vars

## 6. Integration Verification

- [x] 6.1 Run `helm lint deploy/helm/hasir` — verify: 0 errors, 0 failures
- [x] 6.2 Run `make vuln` and grep changed files for hardcoded secrets — verify: exits 0, no hits
- [x] 6.3 Run `bun changeset` to create changeset entry — verify: `.changeset/` file present
