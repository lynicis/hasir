# 0001. Monorepo Structure and Tooling

## Status
Accepted

## Date
2026-07-14

## Decision
We will migrate the five standalone repositories (api, dashboard, hasir-proto, helm-charts, and docker-images) into a single monorepo. We will use Turborepo for task orchestration and Bun workspaces for package management.

## Problem being solved
Managing five separate repositories creates significant operational overhead. It makes cross-repository changes difficult, complicates dependency management, and slows down local development. We need a unified repository structure that allows us to build, test, and deploy services independently while maintaining a single source of truth.

## Alternatives considered
* **Git Submodules**: Keeping separate repositories and linking them via git submodules. This was rejected because submodules are notoriously difficult to manage, prone to developer error, and do not solve the cross-repository dependency management problem.
* **Nx or Bazel**: Using Nx or Bazel instead of Turborepo. Bazel was rejected due to its steep learning curve and high configuration overhead. Nx was rejected because Turborepo fits our Go and Next.js stack more naturally with less configuration.
* **pnpm**: Using pnpm instead of Bun workspaces. While pnpm is excellent, Bun workspaces provide faster installation speeds and align with our choice of Bun as the primary JavaScript runtime.

## Trade-offs
Choosing Turborepo and Bun workspaces means we gain speed and simplicity, but we must manage a larger repository and handle occasional Node.js compatibility edge cases in Bun.

## Benefits
* Single source of truth for all code, configuration, and deployment assets.
* Faster builds and tests with Turborepo caching and parallel execution.
* Simplified dependency management and workspace linking with Bun workspaces.
* Improved developer experience with a single command to start the entire stack.

## Drawbacks
* Larger repository size and history.
* Increased complexity in CI/CD configuration to ensure we only build and test changed projects.
* Potential Node.js compatibility issues with Bun for certain legacy packages.

## Why chosen
Turborepo and Bun workspaces provide the best balance of speed, simplicity, and developer experience for our specific stack. They integrate well with Go modules and Next.js, allowing us to scale to 20, 50, or more services without significant configuration overhead.
