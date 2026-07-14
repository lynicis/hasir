# 0008. Vercel Remote Cache for Turborepo

## Status
Accepted

## Date
2026-07-14

## Decision
We will use Vercel Remote Cache as the remote caching backend for Turborepo. We will configure it in our CI/CD pipelines and local development environments using the TURBO_REMOTE_CACHE_SIGNATURE_KEY and VERCEL_TOKEN environment variables.

## Problem being solved
As the monorepo grows to 20, 50, or more services, build and test times will increase significantly. Local caching helps individual developers, but it does not share build artifacts across the team or in CI/CD pipelines. We need a shared, remote cache to speed up builds and tests across all environments.

## Alternatives considered
* **Self-Hosted S3-Compatible Cache**: Setting up a custom caching server using MinIO or AWS S3. This was rejected because it requires setting up, maintaining, and securing a custom caching server, which adds operational overhead.
* **Local Caching Only**: Relying solely on local caching. This was rejected because it does not share build artifacts across the team or in CI/CD pipelines, leading to redundant builds and slow CI/CD runs.

## Trade-offs
Using Vercel Remote Cache introduces a dependency on Vercel's infrastructure and potential cost implications. However, it requires zero maintenance and integrates seamlessly with Turborepo.

## Benefits
* Zero maintenance overhead.
* Seamless integration with Turborepo.
* Significant reduction in build and test times in CI/CD pipelines and local development.
* Shared cache across all developers and CI/CD runners.

## Drawbacks
* Dependency on Vercel's infrastructure.
* Potential cost implications as cache usage scales.
* Requires managing access tokens for developers and CI/CD runners.

## Why chosen
Vercel Remote Cache is the native remote caching solution for Turborepo. It requires no setup or maintenance, providing immediate value and performance improvements with minimal configuration. This allows the platform team to focus on other migration tasks rather than managing caching infrastructure.
