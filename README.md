<div align="center">

# Hasir

**A self-hosted platform for managing protobuf APIs, SDK generation, and Git-native workflows.**

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-15-000000?logo=next.js&logoColor=white)](https://nextjs.org)
[![Buf](https://img.shields.io/badge/Buf-Protobuf-4353FF?logo=buf&logoColor=white)](https://buf.build)
[![Bun](https://img.shields.io/badge/Bun-1.3-F9F1E1?logo=bun&logoColor=black)](https://bun.sh)
[![codecov](https://codecov.io/gh/lynicis/hasir/graph/badge.svg)](https://codecov.io/gh/lynicis/hasir)

</div>

---

## Overview

Hasir is a cloud-native platform that combines a Go API, a Next.js dashboard, and protocol buffer contracts into a single monorepo. It provides Git-over-SSH access, automatic SDK generation from protobuf definitions, JWT authentication, and OpenTelemetry observability — all orchestrated with Turborepo and Bun.

> [!NOTE]
> The original standalone repositories (`api`, `dashboard`, `proto`, `docker-images`) under the [protohasir](https://github.com/protohasir) organization are retired. They remain available for older versions, but all active development continues here.

## Features

- **ConnectRPC API** — Type-safe RPC services built with Connect for Go, backed by PostgreSQL (pgx)
- **Git-over-SSH** — Native SSH server for Git push/pull workflows using go-git
- **SDK Generation** — Automatic client SDK generation from Buf modules with configurable workers
- **Next.js Dashboard** — Admin UI with Radix primitives, TanStack Query, and iron-session auth
- **Protobuf Contracts** — Shared `.proto` definitions with Buf-managed code generation for Go and TypeScript
- **OpenTelemetry** — Distributed tracing and structured logging via zap
- **Kubernetes-Ready** — Helm charts, multi-stage Docker builds, and Docker Compose for local development

## Repository Structure

```
apps/api/            Go API service (ConnectRPC, PostgreSQL, JWT, SSH)
apps/dashboard/      Next.js dashboard (React, shadcn/ui, Tailwind)
proto/               Protocol buffer definitions (Buf)
packages/            Shared configs (eslint, tsconfig, UI components)
deploy/helm/         Helm chart for Kubernetes deployment
docker/              Docker Compose stack (nginx, certbot)
scripts/             Build, release, and lint utilities
docs/                Architecture docs and ADRs
```

## Prerequisites

| Tool       | Version   |
| ---------- | --------- |
| [Bun](https://bun.sh) | >= 1.3.14 |
| [Go](https://go.dev)  | >= 1.26   |
| [Node.js](https://nodejs.org) | >= 22 |
| [Buf](https://buf.build/docs/installation) | Latest |
| [Docker](https://docs.docker.com/get-docker/) | With Buildx |
| [Helm](https://helm.sh) | >= 3.14 *(optional, for K8s deployment)* |

## Quick Start

```bash
# Clone the repository
git clone https://github.com/lynicis/hasir.git
cd hasir

# Install dependencies, generate proto, update Helm deps
make setup

# Copy example configs
cp apps/api/config.example.json apps/api/config.json
cp apps/dashboard/.env.example apps/dashboard/.env.local

# Start API + dashboard in parallel
make dev
```

The API runs at `http://localhost:8080` and the dashboard at `http://localhost:3000`.

> [!TIP]
> You can also start services individually:
> ```bash
> cd apps/api && make dev         # API only
> cd apps/dashboard && bun dev    # Dashboard only
> ```

## Available Commands

All top-level commands are available through `make`:

| Command           | Description                                      |
| ----------------- | ------------------------------------------------ |
| `make setup`      | Install deps, generate proto, update Helm charts |
| `make dev`        | Start API + dashboard via Turborepo              |
| `make build`      | Build all affected workspaces                    |
| `make test`       | Run tests across affected workspaces             |
| `make lint`       | ESLint + golangci-lint across the monorepo       |
| `make typecheck`  | TypeScript type checking                         |
| `make proto`      | Regenerate code from `.proto` definitions        |
| `make docker`     | Build all Docker images (Buildx Bake)            |
| `make helm-lint`  | Lint and validate Helm charts                    |
| `make clean`      | Remove build artifacts and caches                |
| `make release`    | Tag and release a service (`app=api bump=patch`) |

## Configuration

### API (`apps/api/config.json`)

Copy `config.example.json` and fill in your values:

| Key                  | Description                              |
| -------------------- | ---------------------------------------- |
| `server.port`        | HTTP listen port (default: `8080`)       |
| `ssh.port`           | SSH server port (default: `2222`)        |
| `postgresql.*`       | Database connection settings             |
| `smtp.*`             | Email delivery (registration, invites)   |
| `jwtSecret`          | Secret for signing JWT tokens            |
| `sdkGeneration.*`    | Worker count, poll interval, Buf config  |
| `otel.*`             | OpenTelemetry trace endpoint             |

### Dashboard (`apps/dashboard/.env.local`)

| Variable                | Description                    |
| ----------------------- | ------------------------------ |
| `NEXT_PUBLIC_API_URL`   | API base URL                   |
| `NEXT_PUBLIC_BASE_URL`  | Dashboard public URL           |
| `SESSION_SECRET`        | iron-session encryption secret |

## Deployment

### Docker Compose

```bash
cd docker
cp .env.example .env
docker compose up -d
```

Includes nginx reverse proxy and optional certbot for TLS certificates.

### Kubernetes (Helm)

```bash
helm dependency update deploy/helm/hasir
helm install hasir deploy/helm/hasir -f deploy/helm/hasir/values.yaml
```

## Documentation

| Document | Description |
| -------- | ----------- |
| [Architecture](docs/ARCHITECTURE.md) | System design and component boundaries |
| [Migration Guide](docs/MIGRATION.md) | Database and breaking-change migrations |
| [Release Strategy](docs/RELEASE.md) | Versioning, tagging, and release workflow |
| [ADRs](docs/adr/) | Architecture Decision Records |
