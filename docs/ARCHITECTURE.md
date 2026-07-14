# Hasir Platform Monorepo Architecture Documentation

## 1. Executive Summary

This document defines the architecture and execution blueprint for migrating the Hasir platform from five standalone repositories into a single unified monorepo. The primary goal is to establish a production-ready repository structure that supports independent builds, testing, and deployments while scaling to dozens of services.

### Scope Boundaries

#### In Scope
* Repository structure and workspace layout.
* Build, test, and lint orchestration using Turborepo.
* Git history migration preserving commit logs and tags.
* Protocol Buffer codegen pipeline using buf.
* Docker asset organization and multi-stage builds.
* Helm chart organization and deployment configuration.
* CI/CD pipeline design using GitHub Actions.
* Local development workflow.
* Release, versioning, and branching strategy.

#### Out of Scope
* Application business logic changes.
* Observability platform selection, though baseline health endpoints and structured logging are defined.
* Authentication and authorization architecture.
* Cloud provider or Kubernetes distribution selection.
* Replacing any element of the fixed technology stack.

---

## 2. Assumptions & Environment Context

We base our design on several key assumptions to ensure consistency across the platform.

* **Git Hosting**: GitHub is the primary repository host.
* **Container Registry**: GitHub Container Registry (GHCR) is used for storing Docker images. It integrates natively with GitHub Actions and provides a unified permission model.
* **Cloud Provider**: A standard managed Kubernetes service is assumed to be already chosen.
* **Current Scale**: The five source repositories have small-to-medium Git history. If we had very large histories (multi-GB, 50k+ commits), we would filter out large binary blobs before migration.
* **License**: The platform is proprietary and all rights are reserved.
* **Migration End-State**: The five source repositories will be archived and marked read-only once the migration is verified.
* **Tool Versions**:
  * Bun: `1.3.14` (verify before executing, as newer versions may be available in mid-2026).
  * Turborepo: `2.6.0` (verify before executing).
  * Go: `1.26` (verify before executing).
  * Node.js LTS: `22.0.0` (verify before executing).
  * Kubernetes: `1.30` (verify before executing).
  * Helm: `3.14` (verify before executing).
  * buf: `1.42.0` (verify before executing).
  * GitHub Actions runner image: `ubuntu-latest` (verify before executing).

---

## 3. High-Level Architecture

The Hasir monorepo uses Turborepo for task orchestration and Bun for package management. Go services remain standard Go modules, participating in the Turborepo task graph via thin package wrappers.

```
+-----------------------------------------------------------------+
|                           Turborepo                             |
|  (Orchestrates builds, tests, lints, and protobuf generation)   |
+-----------------------------------------------------------------+
       |                                 |
       v                                 v
+-----------------------+       +---------------------------------+
|      apps/api/        |       |        apps/dashboard/          |
|  (Go 1.26 Backend)    |       |  (Next.js Frontend via Bun)     |
+-----------------------+       +---------------------------------+
       |                                 |
       +----------------+----------------+
                        |
                        v
            +-----------------------+
            |        proto/         |
            |  (buf Protobuf Gen)   |
            +-----------------------+
```

### Key Architectural Decisions

* **Flattened Proto Directory**: We flatten the protobuf directory to `proto/` directly. This avoids redundant nesting and simplifies import paths.
* **Go Integration**: Go services use a minimal `package.json` to expose scripts wrapping `go build` and `go test`. This allows Turborepo to manage dependencies without forcing Go into the Node.js ecosystem.
* **Secrets Handling**: We use External Secrets Operator with AWS Secrets Manager. Plaintext secrets are never committed to the repository.

---

## 4. Final Repository Tree

The monorepo structure is organized as follows:

```text
.
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── apps/
│   ├── api/
│   │   ├── .golangci.yaml
│   │   ├── Dockerfile
│   │   ├── Makefile
│   │   ├── go.mod
│   │   └── package.json
│   └── dashboard/
│       ├── Dockerfile
│       ├── bunfig.toml
│       ├── eslint.config.mjs
│       ├── next.config.ts
│       ├── package.json
│       └── tsconfig.json
├── deploy/
│   └── helm/
│       └── hasir/
│           ├── Chart.yaml
│           ├── values.yaml
│           └── templates/
├── docker/
│   ├── api/
│   ├── base/
│   │   ├── Dockerfile.go
│   │   └── Dockerfile.node
│   ├── dashboard/
│   └── shared/
│       ├── docker-bake.hcl
│       ├── nginx/
│       └── certbot/
├── docs/
│   ├── ARCHITECTURE.md
│   ├── MIGRATION.md
│   ├── RELEASE.md
│   └── adr/
├── packages/
│   ├── config/
│   ├── shared/
│   └── tooling/
├── proto/
│   ├── buf.gen.yaml
│   ├── buf.lock
│   ├── buf.yaml
│   └── organization/
│       └── v1/
├── scripts/
│   ├── clean.sh
│   ├── docker-build.sh
│   ├── helm-lint.sh
│   ├── proto-gen.sh
│   ├── release.sh
│   └── setup.sh
├── CODEOWNERS
├── package.json
├── bunfig.toml
├── turbo.json
└── LICENSE
```

---

## 23. CI/CD Architecture

Our CI/CD pipeline is built on GitHub Actions. It uses Turborepo's remote caching to optimize build times.

### Workflow Design

* **Continuous Integration (`ci.yml`)**: Runs on every pull request and push to the main branch. It executes linting, typechecking, unit tests, and protobuf validation.
* **Continuous Delivery (`release.yml`)**: Runs when a release tag is pushed. It builds Docker images and publishes them to GHCR.

### Optimization Strategies

* **Remote Caching**: We use Vercel Remote Cache to share build artifacts across developer machines and CI runners.
* **Filtering**: Turborepo filters tasks using the `--filter` flag. Only modified workspaces and their dependents are built or tested.
* **Concurrency**: We configure concurrency groups to cancel in-progress runs when new commits are pushed to the same branch.

---

## 24. Release Strategy

We use independent per-application versioning combined with trunk-based development.

### Versioning Model

Each application is versioned independently using semantic versioning. Git tags are prefixed with the application name, such as `api/v1.4.0` or `dashboard/v2.1.0`. This approach allows us to deploy services without coupling their release cycles.

### Branching Model

We follow trunk-based development. Developers merge short-lived feature branches into `main` after passing CI checks.

### Release Process

1. A developer runs `scripts/release.sh <app> <patch|minor|major>`.
2. The script calculates the next version, creates a git tag, and pushes it to GitHub.
3. The release workflow builds the Docker image, tags it with the version, and publishes it to GHCR.
4. The Helm chart is updated with the new image tag.

---

## 25. Rollback Strategy

Rollbacks must be fast and reliable to minimize downtime.

### Application Rollback

To roll back an application, we redeploy the previous stable Docker image tag. This is done by updating the image tag in the Helm values file and running the deployment pipeline.

### Infrastructure Rollback

Helm releases can be rolled back using the `helm rollback` command. This reverts the Kubernetes resources to their previous state.

---

## 26. Future Scalability Considerations

The monorepo is designed to scale to 50+ services and multiple frontend applications.

* **Task Graph Scaling**: Turborepo handles large task graphs efficiently. Remote caching keeps build times low even as the codebase grows.
* **Go Module Isolation**: Go services remain independent modules. This prevents dependency conflicts between services.
* **Code Ownership**: The `CODEOWNERS` file routes pull request reviews to the correct teams automatically.

---

## 27. Risks and Mitigations

### Cache Poisoning
* **Risk**: Incorrect cache hits due to missing task inputs.
* **Mitigation**: We define strict inputs in `turbo.json` and include all configuration files in global dependencies.

### Repository Size Growth
* **Risk**: Large repository size slows down git operations.
* **Mitigation**: We enforce a strict policy against committing large binary files. We use shallow clones in CI pipelines.

### Shared Package Coupling
* **Risk**: Changes to shared packages trigger builds across all applications.
* **Mitigation**: We keep shared packages small and focused. We use interface boundaries to decouple services.
