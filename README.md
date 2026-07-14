# Hasir Platform Monorepo

Welcome to the Hasir platform monorepo. This repository consolidates the Go API, Next.js dashboard, shared protobuf contracts, Helm charts, and Docker images into a single unified workspace.

---

## Table of Contents

1. [Repository Overview](#repository-overview)
2. [Repository Structure](#repository-structure)
3. [Local Setup](#local-setup)
4. [Development Workflow](#development-workflow)
   * [Running Services](#running-services)
   * [Adding a Go Service](#adding-a-go-service)
   * [Adding a Bun Application](#adding-a-bun-application)
   * [Adding Shared Packages](#adding-shared-packages)
   * [Adding Protobuf Definitions](#adding-protobuf-definitions)
5. [Docker Workflow](#docker-workflow)
6. [Helm Workflow](#helm-workflow)
7. [CI/CD Pipeline](#cicd-pipeline)
8. [Release Process](#release-process)
9. [Branching & Versioning Strategy](#branching--versioning-strategy)
10. [Ownership Map](#ownership-map)
11. [Secrets Handling](#secrets-handling)
12. [Troubleshooting](#troubleshooting)
13. [Scope Boundaries](#scope-boundaries)
14. [Additional Documentation](#additional-documentation)

---

## Repository Overview

The Hasir platform is a cloud-native system designed to scale to dozens of services. We use Turborepo for task orchestration and Bun for package management. Go services remain standard Go modules, participating in the Turborepo task graph via thin package wrappers.

---

## Repository Structure

```text
.
├── .github/             # GitHub Actions workflows
├── apps/                # Applications and services
│   ├── api/             # Go backend service
│   └── dashboard/       # Next.js frontend dashboard
├── deploy/              # Deployment configurations
│   └── helm/            # Helm charts for Kubernetes
├── docker/              # Dockerfiles and image definitions
├── docs/                # Architecture, migration, and release docs
├── packages/            # Shared packages and configurations
│   ├── config/          # Shared linting and tsconfig files
│   ├── shared/          # Shared TypeScript utilities
│   └── tooling/         # Shared build and development tools
├── proto/               # Protocol Buffer definitions
├── scripts/             # Helper scripts for development and CI
├── CODEOWNERS           # Code ownership mapping
├── package.json         # Root package configuration
├── bunfig.toml          # Bun configuration
└── turbo.json           # Turborepo configuration
```

---

## Local Setup

Ensure you have the following tools installed on your system:
* Bun `1.3.14` or higher (verify before executing).
* Go `1.26` or higher (verify before executing).
* Docker with Buildx support.
* Helm `3.14` or higher (verify before executing).

Run the setup script to install dependencies and configure the workspace:

```bash
./scripts/setup.sh
```

This script installs Bun dependencies, updates Helm chart dependencies, and downloads Go modules.

---

## Development Workflow

### Running Services

To start the local development environment, run:

```bash
bun run dev
```

This command starts the API and dashboard services in parallel.

### Adding a Go Service

1. Create a new directory under `apps/` (such as `apps/my-service/`).
2. Initialize a Go module: `go mod init hasir-my-service`.
3. Create a `package.json` file in the new directory to wrap Go commands:
   ```json
   {
     "name": "hasir-my-service",
     "version": "0.0.0",
     "private": true,
     "scripts": {
       "build": "go build -o dist/my-service main.go",
       "test": "go test ./...",
       "lint": "golangci-lint run"
     }
   }
   ```
4. Register the workspace in the root `package.json` if it is not covered by the wildcard.

### Adding a Bun Application

1. Create a new directory under `apps/` (such as `apps/my-app/`).
2. Initialize a package: `bun init`.
3. Configure `package.json` with the required scripts and dependencies.
4. Ensure the workspace is registered in the root `package.json`.

### Adding Shared Packages

1. Create a new directory under `packages/` (such as `packages/my-package/`).
2. Initialize a package: `bun init`.
3. Reference the package in other workspaces using the `workspace:*` protocol.

### Adding Protobuf Definitions

1. Add your `.proto` files to the `proto/` directory.
2. Run the generation script to regenerate Go and TypeScript files:
   ```bash
   ./scripts/proto-gen.sh
   ```

---

## Docker Workflow

We use multi-stage Docker builds to keep image sizes minimal.

### Building Images

To build all Docker images using Buildx Bake, run:

```bash
./scripts/docker-build.sh
```

You can also build a specific service image:

```bash
docker build -t hasir-api -f apps/api/Dockerfile .
```

---

## Helm Workflow

Helm charts are located in `deploy/helm/`.

### Linting and Validation

To lint and validate the Helm charts, run:

```bash
./scripts/helm-lint.sh
```

This script runs `helm lint` and `helm template` to ensure the charts are valid.

---

## CI/CD Pipeline

Our CI/CD pipeline is built on GitHub Actions.

* **Continuous Integration**: Runs on every pull request. It executes linting, testing, and validation.
* **Continuous Delivery**: Runs when a release tag is pushed. It builds and publishes Docker images to GHCR.

---

## Release Process

Releases are triggered by pushing a versioned git tag.

```bash
# Release a new patch version for the api service
bun run scripts/release.sh api patch
```

This script bumps the version, creates a tag (such as `api/v1.4.1`), and pushes it to GitHub.

---

## Branching & Versioning Strategy

* **Branching Model**: We follow trunk-based development. All feature branches are short-lived and merged directly into `main`.
* **Versioning Model**: We use independent per-application versioning. Each service is versioned and tagged independently.

---

## Secrets Handling

Plaintext secrets must never be committed to the repository.

* **Local Development**: Use `.env.local` files (which are gitignored) to store local secrets.
* **Production**: We use External Secrets Operator to fetch secrets from AWS Secrets Manager and inject them into Kubernetes pods.

---

## Troubleshooting

### Protobuf Generation Fails
Ensure you have `buf` installed and available in your PATH. Run `bun install` to install the local buf package.

### Turborepo Cache Issues
If you suspect cache corruption, run `bun run clean` to clear all build caches and node_modules.

---

## Scope Boundaries

This repository is designed to manage the platform infrastructure and build orchestration. It explicitly does not cover:
* Application business logic changes.
* Observability platform selection.
* Authentication and authorization architecture.

---

## Additional Documentation

For more detailed information, refer to the following documents:
* [Architecture Documentation](docs/ARCHITECTURE.md)
* [Migration Guide](docs/MIGRATION.md)
* [Release Strategy](docs/RELEASE.md)
