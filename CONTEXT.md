# System Context: Hasir Platform

This document provides a comprehensive overview of the Hasir platform codebase context for developers and AI assistants.

---

## 1. System Overview

Hasir is a self-hosted platform for managing Protocol Buffer (protobuf) schemas.
- **Schema Registry**: Serves as a version-controlled repository for protobuf definitions.
- **Git-over-SSH**: Implements a custom SSH server in Go that supports Git push/pull natively. Developers push `.proto` updates directly to Hasir via standard Git CLI.
- **Automatic SDK Generation**: On every schema push, Hasir validates definitions (via `buf`) and automatically generates and packages client SDKs (TypeScript, Go, etc.) using configurable worker pools.
- **Management Dashboard**: A Next.js Web UI for managing organizations, repositories, users, and SSH public keys.

---

## 2. Tech Stack

- **Go (Backend - `apps/api`)**: Go 1.26, ConnectRPC (Connect, gRPC, gRPC-Web compatible), PostgreSQL (`pgx/v5` driver and SQL migrations), custom Git SSH server (`golang.org/x/crypto/ssh`), golangci-lint, and Buf tooling integration.
- **Next.js (Frontend - `apps/dashboard`)**: Next.js 16, React, TypeScript, Bun package manager, Tailwind CSS, shadcn/ui components.
- **Schema & Codegen (`proto`)**: Buf CLI, Protobuf definitions.
- **Infrastructure & Deployment**:
  - **Docker Compose**: Single-server setup in `docker/` containing Nginx (reverse proxy, rate-limiting, SSL termination via Let's Encrypt/Certbot), PostgreSQL, API, and Dashboard containers.
  - **Kubernetes**: Helm chart located in `deploy/helm/hasir`.
  - **CI/CD**: GitHub Actions workflows in `.github/workflows/` (`ci.yml` and `release.yml`).

---

## 3. Directory Structure

```text
.
├── .github/
│   └── workflows/          # GitHub Actions CI/CD workflows
├── apps/
│   ├── api/                # Go API & Git-over-SSH service
│   │   ├── internal/       # Application domain packages (user, registry, organization)
│   │   ├── pkg/            # Reusable packages (sdkgenerator, postgres, authentication, config)
│   │   ├── migrations/     # PostgreSQL SQL migrations
│   │   ├── main.go         # Application entry point
│   │   └── Makefile        # Go service commands
│   └── dashboard/          # Next.js web application
│       ├── app/            # Next.js page components
│       ├── components/     # UI/UX layout and visual components
│       └── package.json    # Next.js workspace configurations
├── deploy/
│   └── helm/               # Kubernetes Helm deployment configuration
├── docker/
│   ├── base/               # Base Dockerfiles for build optimization
│   ├── shared/             # Docker Compose setup, nginx configs, certificates
│   └── README.md           # Docker deployment documentation
├── docs/                   # Architecture documents, ADRs, release guides
├── packages/               # Shared TS packages (eslint, tsconfig, shared UI/utils)
├── proto/                  # Buf configuration and protobuf schemas
├── scripts/                # Development, setup, and build scripts
├── AGENTS.md               # AI developer guidelines and commands
├── CLAUDE.md               # GitNexus configuration and workflow rules
├── DESIGN.md               # Design tokens and visual specifications
└── README.md               # Main repository readme
```

---

## 4. Key Workflows & CLI Commands

Always run these commands from the monorepo root unless workspace-specific execution is required:

- **Setup & Installation**:
  ```bash
  make setup                # Installs dependencies, lint/proto tools, setups config files
  ```
- **Local Development**:
  ```bash
  make dev                  # Starts Go API and Next.js dev servers concurrently
  ```
- **Building Services**:
  ```bash
  make build                # Builds Go API and compiles Next.js frontend
  ```
- **Testing**:
  ```bash
  make test                 # Runs go test and bun test across the monorepo
  ```
- **Linting & Validation**:
  ```bash
  make lint                 # Runs golangci-lint, eslint, and buf lint
  ```
- **Protobuf Compilation**:
  ```bash
  make proto                # Triggers 'buf generate' on the proto/ directory
  ```

---

## 5. Coding & Contribution Rules

- **Type Safety**: TypeScript definitions must be clean. Strictly avoid `as any`, `@ts-ignore`, or `@ts-expect-error`.
- **Database Migrations**: Always use raw SQL migration files under `apps/api/migrations/`. Avoid schema drift.
- **Git Flow**:
  - Trunk-based development on `main` branch.
  - Release tagging format: `<app>/<semver>` (e.g. `api/v1.0.0`, `dashboard/v1.0.0`).
- **Refactoring & Impact**: Prior to changing Go or TS symbols, review dependents using GitNexus graph commands (`impact` tool). Ensure no regressions are introduced.
