# Contributing to Hasir

Thanks for your interest in contributing! This document covers the practical
guidelines for working with this monorepo.

## Table of Contents

- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Requests](#pull-requests)
- [Issues](#issues)
- [Code of Conduct](#code-of-conduct)

## Project Structure

```
hasir/
├── apps/
│   ├── api/            # Go API server
│   └── dashboard/      # Next.js dashboard
├── deploy/
│   └── helm/           # Helm chart deployment
├── docker/             # Docker Compose environment
├── proto/              # Protocol Buffers / Buf definitions
└── .github/
    └── workflows/      # CI/CD (monorepo-level only)
```

Each app has its own `Makefile`, dependency file, and tests. The root
`Makefile` delegates to each.

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose (for local environment)
- Helm (for deployment work)
- Buf CLI (for protobuf changes)

### Quick Setup

```bash
make install        # install toolchain dependencies
make dev            # start local development environment
```

See each app's `README.md` for app-specific setup.

## Development Workflow

1. **Branch**: Create a feature branch from `main`.
2. **Small commits**: Keep commits atomic and well-described.
3. **Changes scoped to apps**: A commit should ideally touch only one app
   unless the change is cross-cutting (e.g., shared proto changes).
4. **Update generated code**: If you change `.proto` files, regenerate:
   ```bash
   make proto
   ```
5. **Check CI locally before pushing**:
   ```bash
   make lint
   make test
   ```

## Code Style

- **Go**: Follow standard `gofmt` / `golangci-lint` rules. Project has a
  `.golangci.yaml` at `apps/api/`.
- **TypeScript/React**: Follow the project's ESLint + Prettier config in
  `apps/dashboard/`.
- **Protobuf**: Follow Buf's lint and breaking-change rules.
- **Commit messages**: Use conventional commits
  (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, etc.).

## Testing

```bash
make test       # run all tests across all apps
```

- Go tests use the standard `testing` package with mockgen-generated mocks.
- Dashboard tests use Vitest.
- Always write tests for new functionality.
- Update existing tests when behavior changes.

## Pull Requests

1. PRs should target `main`.
2. Title should follow conventional commits.
3. Description should explain:
   - What the change does
   - Why it's needed
   - How it was tested
4. Keep PRs focused — one logical change per PR.
5. Ensure CI passes before requesting review.

### Review Process

- At least one approval is required before merging.
- If you're a first-time contributor, a maintainer will review within
  a few business days.
- Address review feedback with additional commits — we squash on merge.

## Issues

- **Bug reports**: Include reproduction steps, expected vs actual behavior,
  and environment details.
- **Feature requests**: Explain the use case and why it belongs in Hasir
  rather than as an external tool.
- **Security issues**: Do **not** file a public issue. See `SECURITY.md`.

## Code of Conduct

All contributors are expected to follow our
[Code of Conduct](CODE_OF_CONDUCT.md). Be respectful, constructive, and
professional.

## Questions?

Open a discussion or reach out to **me@lynicis.dev**.
